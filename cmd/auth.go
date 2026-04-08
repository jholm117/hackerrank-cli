package cmd

import (
	"fmt"
	"os"

	"github.com/jholm117/hackerrank-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// configPath is the config file path. Overridden in tests.
var configPath = config.DefaultPath()

// readToken reads a token without echoing. Overridden in tests.
var readToken = func() (string, error) {
	fmt.Fprint(os.Stderr, "Enter HackerRank API token: ")
	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save API token to config",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := readToken()
		if err != nil {
			return err
		}
		if token == "" {
			return fmt.Errorf("token cannot be empty")
		}

		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		cfg.Token = token
		if err := config.Save(cfg, configPath); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Token saved to %s\n", configPath)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored token",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		cfg.Token = ""
		if err := config.Save(cfg, configPath); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Token removed.")
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current auth state",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configPath)
		if err != nil {
			return err
		}
		token := config.ResolveToken(flagToken, cfg)
		if token == "" {
			fmt.Println("Not authenticated. Run: hr auth login")
			return nil
		}
		source := "config file"
		if flagToken != "" {
			source = "--token flag"
		} else if os.Getenv("HACKERRANK_API_TOKEN") != "" {
			source = "HACKERRANK_API_TOKEN env var"
		}
		fmt.Printf("Authenticated via %s\n", source)
		fmt.Printf("Token: %s...%s\n", token[:4], token[len(token)-4:])
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
	rootCmd.AddCommand(authCmd)
}
