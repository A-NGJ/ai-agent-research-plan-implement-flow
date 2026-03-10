package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	scanStatus     string
	scanType       string
	scanProposal   string
	scanReferences string
	scanArchivable bool
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Discover and filter artifacts in .thoughts/",
	Long:  "Walk .thoughts/ directory, parse frontmatter, and return artifacts matching filter criteria.",
	RunE:  runScan,
}

func init() {
	addFormatFlag(scanCmd)
	addThoughtsDirFlag(scanCmd)
	scanCmd.Flags().StringVar(&scanStatus, "status", "", "Filter by frontmatter status")
	scanCmd.Flags().StringVar(&scanType, "type", "", "Filter by artifact type (plan, proposal, research, etc.)")
	scanCmd.Flags().StringVar(&scanProposal, "proposal", "", "Filter by frontmatter proposal field")
	scanCmd.Flags().StringVar(&scanReferences, "references", "", "Find files that reference this path")
	scanCmd.Flags().BoolVar(&scanArchivable, "archivable", false, "Show only complete/superseded artifacts not in archive/")
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	filters := scanner.Filters{
		Status:     scanStatus,
		Type:       scanType,
		Proposal:   scanProposal,
		References: scanReferences,
		Archivable: scanArchivable,
	}

	results, err := scanner.Scan(thoughtsDirFlag, filters)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	format := formatFlag
	if format == "" {
		format = "json"
	}

	switch format {
	case "json":
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
	case "md":
		printScanTable(results)
	default:
		return fmt.Errorf("unknown format: %s (expected json or md)", format)
	}

	return nil
}

func printScanTable(results []scanner.ArtifactInfo) {
	if len(results) == 0 {
		fmt.Println("No artifacts found.")
		return
	}
	fmt.Println("| Path | Type | Status | Title | Date |")
	fmt.Println("|------|------|--------|-------|------|")
	for _, a := range results {
		status := "-"
		if a.Status != nil {
			status = *a.Status
		}
		title := "-"
		if a.Title != nil {
			title = strings.ReplaceAll(*a.Title, "|", "\\|")
		}
		date := "-"
		if a.Date != nil {
			date = *a.Date
		}
		fmt.Printf("| `%s` | %s | %s | %s | %s |\n", a.Path, a.Type, status, title, date)
	}
}
