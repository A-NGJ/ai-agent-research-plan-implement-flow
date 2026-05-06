package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/A-NGJ/rpi/internal/frontmatter"
	"github.com/A-NGJ/rpi/internal/scanner"
	"github.com/spf13/cobra"
)

// ResumeResult is the session-level overview returned by rpi resume.
type ResumeResult struct {
	Artifacts  []ResumeArtifact   `json:"artifacts"`
	ActivePlan *ActivePlanSummary `json:"active_plan,omitempty"`
	Suggestion *Suggestion        `json:"suggestion,omitempty"`
}

// ResumeArtifact is a compact artifact entry in the resume output.
type ResumeArtifact struct {
	Path   string `json:"path"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Topic  string `json:"topic"`
}

// ActivePlanSummary is a compact plan context for the session resume.
type ActivePlanSummary struct {
	Path         string   `json:"path"`
	Topic        string   `json:"topic"`
	CurrentPhase string   `json:"current_phase"`
	Progress     Progress `json:"progress"`
	NextItems    []string `json:"next_items"`
}

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Show active work and suggested next steps",
	Long: `Return a session-level overview of active RPI work.

Lists all active and draft artifacts, shows the current implementation
context for the most recent active plan, and suggests the next pipeline
action. Designed for session start to quickly restore context.

Default output is a human-readable text summary. Use --format json for
the structured JSON shape consumed by the rpi_session_resume MCP tool.`,
	Example: `  # Human-readable summary
  rpi resume

  # Structured JSON (script-friendly, same shape as the MCP tool)
  rpi resume --format json`,
	Args: cobra.NoArgs,
	RunE: runResume,
}

func init() {
	addRpiDirFlag(resumeCmd)
	addFormatFlag(resumeCmd)
	rootCmd.AddCommand(resumeCmd)
}

func runResume(cmd *cobra.Command, args []string) error {
	result, err := assembleResume(rpiDirFlag)
	if err != nil {
		return err
	}

	format := formatFlag
	if format == "" {
		format = "text"
	}

	switch format {
	case "text":
		return renderResumeText(cmd, result)
	case "json":
		return renderResumeJSON(cmd, result)
	default:
		return fmt.Errorf("unknown format: %s (expected text or json)", format)
	}
}

func renderResumeJSON(cmd *cobra.Command, r *ResumeResult) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

// resumeArtifactTypeOrder is the canonical type order for the Artifacts section.
var resumeArtifactTypeOrder = []string{"design", "diagnosis", "plan", "research"}

func renderResumeText(cmd *cobra.Command, r *ResumeResult) error {
	if len(r.Artifacts) == 0 && r.ActivePlan == nil {
		fmt.Fprintln(cmd.OutOrStdout(), "No active work — start with /rpi-propose <topic>")
		return nil
	}

	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
	sectionWritten := false

	if len(r.Artifacts) > 0 {
		fmt.Fprintln(w, "Artifacts")
		grouped := make(map[string][]ResumeArtifact)
		for _, a := range r.Artifacts {
			grouped[a.Type] = append(grouped[a.Type], a)
		}
		seen := make(map[string]bool)
		for _, typ := range resumeArtifactTypeOrder {
			if rows, ok := grouped[typ]; ok {
				writeArtifactRows(w, rows)
				seen[typ] = true
			}
		}
		// Include any types not in the canonical order, sorted for determinism.
		var extra []string
		for typ := range grouped {
			if !seen[typ] {
				extra = append(extra, typ)
			}
		}
		sort.Strings(extra)
		for _, typ := range extra {
			writeArtifactRows(w, grouped[typ])
		}
		sectionWritten = true
	}

	if r.ActivePlan != nil {
		if sectionWritten {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, "Active Plan")
		fmt.Fprintf(w, "  %s\n", r.ActivePlan.Topic)
		progress := ""
		if r.ActivePlan.Progress.Total > 0 {
			pct := r.ActivePlan.Progress.Checked * 100 / r.ActivePlan.Progress.Total
			progress = fmt.Sprintf("\t%d/%d (%d%%)", r.ActivePlan.Progress.Checked, r.ActivePlan.Progress.Total, pct)
		}
		phase := r.ActivePlan.CurrentPhase
		if phase == "" {
			phase = "(no phase)"
		}
		fmt.Fprintf(w, "  %s%s\n", phase, progress)
		for _, item := range r.ActivePlan.NextItems {
			fmt.Fprintf(w, "  Next:\t%s\n", item)
		}
		sectionWritten = true
	}

	if r.Suggestion != nil {
		if sectionWritten {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, "Suggestion")
		fmt.Fprintf(w, "  Action:\t%s\n", r.Suggestion.Action)
		fmt.Fprintf(w, "  Why:\t%s\n", r.Suggestion.Reasoning)
		if r.Suggestion.Artifact != "" {
			fmt.Fprintf(w, "  Artifact:\t%s\n", r.Suggestion.Artifact)
		}
	}

	return w.Flush()
}

func writeArtifactRows(w *tabwriter.Writer, rows []ResumeArtifact) {
	for _, a := range rows {
		topic := a.Topic
		if topic == "" {
			topic = "-"
		}
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", a.Type, a.Status, topic, a.Path)
	}
}

// assembleResume builds a session-level overview: active artifacts, plan
// context, and a pipeline suggestion.
func assembleResume(rpiDir string) (*ResumeResult, error) {
	allArtifacts, err := scanner.Scan(rpiDir, scanner.Filters{})
	if err != nil {
		return nil, fmt.Errorf("scan artifacts: %w", err)
	}

	result := &ResumeResult{
		Artifacts: []ResumeArtifact{},
	}

	// Filter to active/draft artifacts for the list
	for _, a := range allArtifacts {
		if a.Status == nil {
			continue
		}
		status := *a.Status
		if status != "active" && status != "draft" {
			continue
		}
		topic := ""
		if a.Title != nil {
			topic = *a.Title
		}
		result.Artifacts = append(result.Artifacts, ResumeArtifact{
			Path:   a.Path,
			Type:   a.Type,
			Status: status,
			Topic:  topic,
		})
	}

	// Build active plan context for the most recent active plan
	activePlans := filterArtifacts(allArtifacts, "plan", "active")
	if len(activePlans) > 0 {
		sortByDateDesc(activePlans)
		plan := activePlans[0]
		result.ActivePlan = buildActivePlanSummary(plan)
	}

	// Pipeline suggestion
	suggestion, err := suggestNext(rpiDir, "")
	if err == nil {
		result.Suggestion = suggestion
	}

	return result, nil
}

// buildActivePlanSummary constructs a compact plan context from an artifact.
func buildActivePlanSummary(plan scanner.ArtifactInfo) *ActivePlanSummary {
	content, err := os.ReadFile(plan.Path)
	if err != nil {
		return nil
	}

	doc, err := frontmatter.Parse(plan.Path)
	if err != nil {
		return nil
	}

	topic, _ := doc.Frontmatter["topic"].(string)
	checkboxes := parseCheckboxes(string(content))
	currentPhase, nextItems := detectCurrentPhase(string(content))

	return &ActivePlanSummary{
		Path:         plan.Path,
		Topic:        topic,
		CurrentPhase: currentPhase,
		Progress: Progress{
			Checked: checkboxes.Checked,
			Total:   checkboxes.Total,
		},
		NextItems: nextItems,
	}
}
