package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const Description = `fluffy is a simple file-sharing web app. You can upload files, or paste text.

By default, the public instance of fluffy is used: https://fluffy.cc

If you'd like to instead use a different instance (for example, one run
internally by your company), you can specify the --server option.

To make that permanent, you can create a config file with contents similar to:

    {"server": "https://fluffy.my.corp"}

This file can be placed at either /etc/fluffy.json or $XDG_CONFIG_HOME/fluffy.json.
`
const defaultServer = "https://fluffy.cc"

type Settings struct {
	Server   string `json:"server"`
	Auth     bool   `json:"auth"`
	Username string `json:"username"`
}

func populateSettingsFromFile(configPath string, s *Settings) error {
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

func GetSettings() (*Settings, error) {
	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("fetching current user: %w", err)
	}

	s := Settings{
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

type Credentials struct {
	Username string
	Password string
}

func WrapWithAuth(
	fn func(cmd *cobra.Command, args []string, creds *Credentials) error,
) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var creds *Credentials
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
			creds = &Credentials{
				Username: user,
				Password: string(passwordBytes),
			}
		}

		return fn(cmd, args, creds)
	}
}

func Bold(s string) string {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return "\x1b[1m" + s + "\x1b[0m"
	} else {
		return s
	}
}

func AddCommonOpts(command *cobra.Command, settings *Settings) {
	command.Flags().String("server", settings.Server, "server to upload to")
	command.Flags().Bool("auth", settings.Auth, "use HTTP Basic auth")
	command.Flags().StringP("user", "u", settings.Username, "username for HTTP Basic auth")
	command.Flags().Bool("direct-link", false, "return direct link to the uploads")
}
