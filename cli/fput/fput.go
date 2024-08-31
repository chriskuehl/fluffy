package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/chriskuehl/fluffy/cli/internal/cli"
)

// This will be set by the linker for release builds.
var version = "(deV)"

var command = &cobra.Command{
	Use:          "fput file [file ...]",
	Long:         "Upload files to fluffy.\n\n" + cli.Description,
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: cli.WrapWithAuth(func(cmd *cobra.Command, args []string, creds *cli.Credentials) error {
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
			req.SetBasicAuth(creds.Username, creds.Password)
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
				fmt.Println(cli.Bold(uploadedFile.Raw))
			}
		} else {
			fmt.Println(cli.Bold(result.Redirect))
		}
		return nil
	}),
}

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
}

func main() {
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
