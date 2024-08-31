package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/chriskuehl/fluffy/cli/internal/cli"
)

// This will be set by the linker for release builds.
var version = "(dev)"

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

var command = &cobra.Command{
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

` + cli.Description,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,

	RunE: cli.WrapWithAuth(func(cmd *cobra.Command, args []string, creds *cli.Credentials) error {
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
			req.SetBasicAuth(creds.Username, creds.Password)
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

		fmt.Println(cli.Bold(location))
		return nil
	}),
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

func init() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Fprintf(os.Stderr, "Failed to read build info\n")
		os.Exit(1)
	}
	command.Version = fmt.Sprintf("%s/%s", version, buildInfo.GoVersion)

	settings, err := cli.GetSettings()
	if err != nil {
		panic(fmt.Errorf("getting settings: %w", err))
	}

	cli.AddCommonOpts(command, settings)

	command.Flags().StringP("language", "l", "autodetect", "language for syntax highlighting")
	command.Flags().VarP(&regex, "regex", "r", "regex of lines to highlight")
	command.Flags().Bool("tee", false, "stream the stdin to stdout before creating the paste")
}

func main() {
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
