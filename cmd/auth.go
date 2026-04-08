package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jholm117/hackerrank-cli/internal/config"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save API token to config",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("Enter HackerRank API token: ")
		reader := bufio.NewReader(os.Stdin)
		token, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		token = strings.TrimSpace(token)
		if token == "" {
			return fmt.Errorf("token cannot be empty")
		}

		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		cfg.Token = token
		if err := config.Save(cfg, path); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Token saved to %s\n", path)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored token",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}
		cfg.Token = ""
		if err := config.Save(cfg, path); err != nil {
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
		path := config.DefaultPath()
		cfg, err := config.Load(path)
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
