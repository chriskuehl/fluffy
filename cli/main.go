package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"

	"os/user"
	"path"
	"strings"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const description = `fluffy is a simple file-sharing web app. You can upload files, or paste text.

By default, the public instance of fluffy is used: https://fluffy.cc

If you'd like to instead use a different instance (for example, one run
internally by your company), you can specify the --server option.

To make that permanent, you can create a config file with contents similar to:

    {"server": "https://fluffy.my.corp"}

This file can be placed at either /etc/fluffy.json or $XDG_CONFIG_HOME/fluffy.json.
`
const defaultServer = "https://fluffy.cc"

type settings struct {
	Server   string `json:"server"`
	Auth     bool   `json:"auth"`
	Username string `json:"username"`
}

func populateSettingsFromFile(configPath string, s *settings) error {
	contents, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		} else {
			return err
		}
	}
	if err := json.Unmarshal(contents, s); err != nil {
		return fmt.Errorf("error parsing config file %s: %w", configPath, err)
	}
	return nil
}

func getSettings() (*settings, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("fetching current user: %w", err)
	}

	s := settings{
		Server:   defaultServer,
		Username: currentUser.Username,
	}

	for _, configPath := range []string{
		"/etc/fluffy.json",
		path.Join(xdg.ConfigHome, "fluffy.json"),
	} {
		err := populateSettingsFromFile(configPath, &s)
		if err != nil {
			return nil, fmt.Errorf("reading config file %s: %w", configPath, err)
		}
	}

	return &s, nil
}

type credentials struct {
	username string
	password string
}

func wrapWithAuth(
	fn func(cmd *cobra.Command, args []string, creds *credentials) error,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var creds *credentials
		flags := cmd.Flags()
		auth, _ := flags.GetBool("auth")

		if auth {
			server, _ := flags.GetString("server")
			user, _ := flags.GetString("user")
			fmt.Fprintf(os.Stderr, "Server: %s\n", server)
			fmt.Fprintf(os.Stderr, "Password for %s: ", user)
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("reading password: %w", err)
			}
			creds = &credentials{
				username: user,
				password: string(passwordBytes),
			}
		}

		return fn(cmd, args, creds)
	}
}

func regexHighlightFragment(regex *regexp.Regexp, content *bytes.Buffer) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content.String()))
	matches := []int{}
	for i := 0; scanner.Scan(); {
		line := scanner.Text()
		if regex.MatchString(line) {
			matches = append(matches, i)
		}
		i++
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanning: %w", err)
	}

	// Squash consecutive matches.
	groups := []string{}
	for i := 0; i < len(matches); i++ {
		start := matches[i]
		end := start
		for ; len(matches) > i+1 && matches[i+1] == end+1; i++ {
			end++
		}
		if start == end {
			groups = append(groups, fmt.Sprintf("L%d", start+1))
		} else {
			groups = append(groups, fmt.Sprintf("L%d-%d", start+1, end+1))
		}
	}

	return strings.Join(groups, ","), nil
}

// Variant of bufio.ScanLines which includes the newline character in the token.
//
// This is desired so that we don't erroneously insert a newline at the end of a final line which
// isn't present in the input. bufio.ScanLines has no way to differentiate whether or not the final
// line has a newline or not.
func ScanLinesWithEOL(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

var fpbCommand = &cobra.Command{
	Use: "fpb [file]",
	Long: `Paste text to fluffy.

Example usage:

    Paste a file:
        fpb some-file.txt

    Pipe the output of a command:
        some-command | fpb

    Specify a language to highlight text with:
        fpb -l python some_file.py
    (Default is to auto-detect the language. You can use "rendered-markdown" for Markdown.)

` + description,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,

	RunE: wrapWithAuth(func(cmd *cobra.Command, args []string, creds *credentials) error {
		server, _ := cmd.Flags().GetString("server")
		tee, _ := cmd.Flags().GetBool("tee")
		language, _ := cmd.Flags().GetString("language")
		directLink, _ := cmd.Flags().GetBool("direct-link")

		path := "-"
		if len(args) > 0 {
			path = args[0]
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.WriteField("language", language)

		content := &bytes.Buffer{}
		if path == "-" {
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Split(ScanLinesWithEOL)
			for scanner.Scan() {
				line := scanner.Text()
				if _, err := content.WriteString(line); err != nil {
					return fmt.Errorf("writing to buffer: %w", err)
				}
				if tee {
					fmt.Print(line)
				}
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("reading from stdin: %w", err)
			}
		} else {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("opening file: %w", err)
			}
			defer file.Close()
			if _, err := io.Copy(content, file); err != nil {
				return fmt.Errorf("copying file: %w", err)
			}
		}
		writer.WriteField("text", content.String())

		if err := writer.Close(); err != nil {
			return fmt.Errorf("closing writer: %w", err)
		}

		req, err := http.NewRequest("POST", server+"/paste?json", body)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		if creds != nil {
			req.SetBasicAuth(creds.username, creds.password)
		}
		q := req.URL.Query()
		q.Add("language", language)
		req.URL.RawQuery = q.Encode()

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("making request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "Unexpected status code: %d\n", resp.StatusCode)
			fmt.Fprintf(os.Stderr, "Error:\n")
			err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			if _, copyErr := io.Copy(os.Stderr, resp.Body); copyErr != nil {
				return fmt.Errorf("copying error: %w for %w", copyErr, err)
			}
			return err
		}

		var result struct {
			Redirect      string `json:"redirect"`
			UploadedFiles struct {
				Paste struct {
					Raw string `json:"raw"`
				} `json:"paste"`
			} `json:"uploaded_files"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}

		location := result.Redirect
		if directLink {
			location = result.UploadedFiles.Paste.Raw
		}

		if regex.r != nil {
			highlight, err := regexHighlightFragment(regex.r, content)
			if err != nil {
				return fmt.Errorf("highlighting: %w", err)
			}
			location += "#" + highlight
		}

		fmt.Println(bold(location))
		return nil
	}),
}

var fputCommand = &cobra.Command{
	Use:          "fput file [file ...]",
	Long:         "Upload files to fluffy.\n\n" + description,
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: wrapWithAuth(func(cmd *cobra.Command, args []string, creds *credentials) error {
		directLink, _ := cmd.Flags().GetBool("direct-link")

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for _, path := range args {
			part, err := writer.CreateFormFile("file", filepath.Base(path))
			if err != nil {
				return fmt.Errorf("creating form file: %w", err)
			}
			if path == "-" {
				if _, err := io.Copy(part, os.Stdin); err != nil {
					return fmt.Errorf("copying stdin: %w", err)
				}
			} else {
				file, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("opening file: %w", err)
				}
				defer file.Close()
				if _, err := io.Copy(part, file); err != nil {
					return fmt.Errorf("copying file: %w", err)
				}
			}
		}
		if err := writer.Close(); err != nil {
			return fmt.Errorf("closing writer: %w", err)
		}

		server, _ := cmd.Flags().GetString("server")
		req, err := http.NewRequest("POST", server+"/upload?json", body)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		if creds != nil {
			req.SetBasicAuth(creds.username, creds.password)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("making request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "Unexpected status code: %d\n", resp.StatusCode)
			fmt.Fprintf(os.Stderr, "Error:\n")
			err := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			if _, copyErr := io.Copy(os.Stderr, resp.Body); copyErr != nil {
				return fmt.Errorf("copying error: %w for %w", copyErr, err)
			}
			return err
		}

		var result struct {
			Redirect      string `json:"redirect"`
			UploadedFiles map[string]struct {
				Raw string `json:"raw"`
			} `json:"uploaded_files"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}

		if directLink {
			for _, uploadedFile := range result.UploadedFiles {
				fmt.Println(bold(uploadedFile.Raw))
			}
		} else {
			fmt.Println(bold(result.Redirect))
		}
		return nil
	}),
}

func bold(s string) string {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return "\x1b[1m" + s + "\x1b[0m"
	} else {
		return s
	}
}

type regexpValue struct {
	r *regexp.Regexp
}

func (v *regexpValue) String() string {
	if v.r == nil {
		return ""
	}
	return v.r.String()
}

func (v *regexpValue) Set(s string) error {
	r, err := regexp.Compile(s)
	if err != nil {
		return err
	}
	v.r = r
	return nil
}

func (v *regexpValue) Type() string {
	return "regex"
}

var regex = regexpValue{}

// This will be set by the linker for release builds.
var Version string

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		panic("could not read build info")
	}
	if Version == "" {
		Version = buildInfo.Main.Version
	}
	version := fmt.Sprintf("%s/%s", Version, buildInfo.GoVersion)
	fpbCommand.Version = version
	fputCommand.Version = version

	settings, err := getSettings()
	if err != nil {
		panic(fmt.Errorf("getting settings: %w", err))
	}

	addCommonOpts := func(command *cobra.Command) {
		command.Flags().String("server", settings.Server, "server to upload to")
		command.Flags().Bool("auth", settings.Auth, "use HTTP Basic auth")
		command.Flags().StringP("user", "u", settings.Username, "username for HTTP Basic auth")
		command.Flags().Bool("direct-link", false, "return direct link to the uploads")
	}

	addCommonOpts(fputCommand)
	addCommonOpts(fpbCommand)

	fpbCommand.Flags().StringP("language", "l", "autodetect", "language for syntax highlighting")
	fpbCommand.Flags().VarP(&regex, "regex", "r", "regex of lines to highlight")

	fpbCommand.Flags().Bool("tee", false, "stream the stdin to stdout before creating the paste")
}

func main() {
	command := fputCommand
	if strings.HasPrefix(filepath.Base(os.Args[0]), "fpb") {
		command = fpbCommand
	}
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
