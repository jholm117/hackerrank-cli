// cmd/candidates.go
package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jholm117/hackerrank-cli/internal/api"
	"github.com/jholm117/hackerrank-cli/internal/output"
	"github.com/spf13/cobra"
)

var candidatesCmd = &cobra.Command{
	Use:   "candidates",
	Short: "Manage test candidates",
}

var candidatesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List candidates for a test",
	RunE: func(cmd *cobra.Command, args []string) error {
		testID, _ := cmd.Flags().GetString("test")
		if testID == "" {
			return fmt.Errorf("--test flag is required")
		}

		c, err := newClient()
		if err != nil {
			return err
		}

		candidates, err := api.Paginate[api.Candidate](c, "/tests/"+testID+"/candidates", nil)
		if err != nil {
			return err
		}

		if flagOutput == "json" {
			return output.WriteJSON(os.Stdout, candidates)
		}

		w := output.NewTableWriter(os.Stdout)
		w.SetHeader([]string{"ID", "NAME", "EMAIL", "SCORE", "STATUS", "DATE"})
		for _, cand := range candidates {
			w.Append([]string{
				cand.ID,
				cand.FullName,
				cand.Email,
				fmt.Sprintf("%.0f%%", cand.PercentageScore),
				fmt.Sprintf("%d", cand.Status),
				cand.AttemptStart,
			})
		}
		w.Render()
		return nil
	},
}

var candidatesGetCmd = &cobra.Command{
	Use:   "get <test-id> <candidate-id>",
	Short: "Show candidate details with submissions",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("additional_fields", "questions,questions.solves,questions.submission_result")

		var cand api.CandidateDetail
		if err := c.Get("/tests/"+args[0]+"/candidates/"+args[1], params, &cand); err != nil {
			return err
		}

		if flagOutput == "table" {
			w := output.NewTableWriter(os.Stdout)
			w.SetHeader([]string{"FIELD", "VALUE"})
			w.Append([]string{"ID", cand.ID})
			w.Append([]string{"Name", cand.FullName})
			w.Append([]string{"Email", cand.Email})
			w.Append([]string{"Score", fmt.Sprintf("%.0f%%", cand.PercentageScore)})
			w.Append([]string{"Started", cand.AttemptStart})
			w.Append([]string{"Ended", cand.AttemptEnd})
			w.Append([]string{"Questions", fmt.Sprintf("%d", len(cand.Questions))})
			w.Render()
			return nil
		}

		return output.WriteJSON(os.Stdout, cand)
	},
}

var candidatesCodeCmd = &cobra.Command{
	Use:   "code <test-id> <candidate-id>",
	Short: "Extract candidate source code",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		saveDir, _ := cmd.Flags().GetString("save")

		c, err := newClient()
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("additional_fields", "questions,questions.solves,questions.submission_result")

		var cand api.CandidateDetail
		if err := c.Get("/tests/"+args[0]+"/candidates/"+args[1], params, &cand); err != nil {
			return err
		}

		i := 0
		for qID, q := range cand.Questions {
			i++
			if !q.Answered {
				continue
			}

			lang := q.Answer.Language
			ext := langExtension(lang)
			name := fmt.Sprintf("q%d_%s", i, qID)

			if saveDir != "" {
				filename := name + ext
				path := filepath.Join(saveDir, filename)
				if err := os.MkdirAll(saveDir, 0o755); err != nil {
					return err
				}
				if err := os.WriteFile(path, []byte(q.Answer.Code), 0o644); err != nil {
					return err
				}
				fmt.Fprintf(os.Stderr, "Saved %s\n", path)
			} else {
				header := fmt.Sprintf("## Question %d: %s (%s) — %.0f/%.0f",
					i, qID, lang, q.Score, q.Score)
				fmt.Println(header)
				fmt.Println(strings.Repeat("─", 60))
				fmt.Println(q.Answer.Code)
				fmt.Println()
			}
		}
		return nil
	},
}

func langExtension(lang string) string {
	lang = strings.ToLower(lang)
	switch {
	case strings.Contains(lang, "python"):
		return ".py"
	case strings.Contains(lang, "java") && !strings.Contains(lang, "javascript"):
		return ".java"
	case strings.Contains(lang, "javascript"):
		return ".js"
	case strings.Contains(lang, "typescript"):
		return ".ts"
	case strings.Contains(lang, "go"):
		return ".go"
	case strings.Contains(lang, "ruby"):
		return ".rb"
	case strings.Contains(lang, "rust"):
		return ".rs"
	case strings.Contains(lang, "c++") || strings.Contains(lang, "cpp"):
		return ".cpp"
	case lang == "c":
		return ".c"
	default:
		return ".txt"
	}
}

func init() {
	candidatesListCmd.Flags().String("test", "", "Test ID (required)")
	candidatesListCmd.MarkFlagRequired("test")
	candidatesCodeCmd.Flags().String("save", "", "Directory to save code files")
	candidatesCmd.AddCommand(candidatesListCmd)
	candidatesCmd.AddCommand(candidatesGetCmd)
	candidatesCmd.AddCommand(candidatesCodeCmd)
	rootCmd.AddCommand(candidatesCmd)
}
