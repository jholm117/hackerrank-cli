// cmd/interviews.go
package cmd

import (
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"

	"github.com/jholm117/hackerrank-cli/internal/api"
	"github.com/jholm117/hackerrank-cli/internal/output"
	"github.com/spf13/cobra"
)

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

func extractTitle(rawHTML string) string {
	re := regexp.MustCompile(`<h[1-6][^>]*>(.*?)</h[1-6]>`)
	m := re.FindStringSubmatch(rawHTML)
	if m == nil {
		re = regexp.MustCompile(`<p[^>]*>(.*?)</p>`)
		m = re.FindStringSubmatch(rawHTML)
	}
	if m == nil {
		return ""
	}
	return html.UnescapeString(strings.TrimSpace(htmlTagRe.ReplaceAllString(m[1], "")))
}

var interviewsCmd = &cobra.Command{
	Use:   "interviews",
	Short: "Manage interviews",
}

var interviewsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List interviews",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		c, err := newClient()
		if err != nil {
			return err
		}

		interviews, err := api.PaginateN[api.Interview](c, "/interviews", nil, limit)
		if err != nil {
			return err
		}

		if flagOutput == "json" {
			return output.WriteJSON(os.Stdout, interviews)
		}

		w := output.NewTableWriter(os.Stdout)
		w.SetHeader([]string{"ID", "TITLE", "STATUS", "CREATED"})
		for _, iv := range interviews {
			w.Append([]string{iv.ID, iv.Title, iv.Status, iv.CreatedAt})
		}
		w.Render()
		return nil
	},
}

var interviewsGetCmd = &cobra.Command{
	Use:   "get <interview-id>",
	Short: "Show interview details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		var iv api.Interview
		if err := c.Get("/interviews/"+args[0], nil, &iv); err != nil {
			return err
		}

		if flagOutput == "table" {
			w := output.NewTableWriter(os.Stdout)
			w.SetHeader([]string{"FIELD", "VALUE"})
			w.Append([]string{"ID", iv.ID})
			w.Append([]string{"Title", iv.Title})
			w.Append([]string{"Status", iv.Status})
			w.Append([]string{"Created", iv.CreatedAt})
			w.Append([]string{"URL", iv.URL})
			w.Render()
			return nil
		}

		return output.WriteJSON(os.Stdout, iv)
	},
}

var interviewsTranscriptCmd = &cobra.Command{
	Use:   "transcript <interview-id>",
	Short: "Get interview transcript",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		var transcript api.Transcript
		if err := c.Get("/interviews/"+args[0]+"/transcript", nil, &transcript); err != nil {
			return err
		}

		if flagOutput == "json" {
			return output.WriteJSON(os.Stdout, transcript)
		}

		for _, msg := range transcript.Messages {
			role := msg.Author
			if msg.Candidate {
				role = fmt.Sprintf("%s (candidate)", msg.Author)
			}
			fmt.Printf("[%s] %s:\n%s\n\n", msg.Timestamp, role, msg.Text)
		}
		return nil
	},
}

var interviewsCodeCmd = &cobra.Command{
	Use:   "code <interview-id>",
	Short: "Extract code from interview pads",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return err
		}

		var recording api.InterviewRecording
		if err := c.GetRaw("/api/interviews/"+args[0]+"/recordings/code", nil, &recording); err != nil {
			return err
		}

		if flagOutput == "json" {
			return output.WriteJSON(os.Stdout, recording)
		}

		for i, q := range recording.Data.Questions {
			if q.QType != "code" || len(q.Runs) == 0 {
				continue
			}
			run := q.Runs[len(q.Runs)-1]
			lang := run.Lang
			if lang == "" {
				lang = "unknown"
			}
			title := extractTitle(q.Question)
			if title != "" {
				fmt.Printf("## Pad %d: %s (%s)\n", i+1, title, lang)
			} else {
				fmt.Printf("## Pad %d (%s)\n", i+1, lang)
			}
			fmt.Println(strings.Repeat("─", 60))
			fmt.Println(run.Code)
			fmt.Println()
		}
		return nil
	},
}

func init() {
	interviewsListCmd.Flags().Int("limit", 20, "Max results to return (0 for all)")
	interviewsCmd.AddCommand(interviewsListCmd)
	interviewsCmd.AddCommand(interviewsGetCmd)
	interviewsCmd.AddCommand(interviewsTranscriptCmd)
	interviewsCmd.AddCommand(interviewsCodeCmd)
	rootCmd.AddCommand(interviewsCmd)
}
