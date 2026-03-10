package main

import (
	"fmt"
	"os"
	"path/filepath"

	tmpl "github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/template"
	"github.com/spf13/cobra"
)

var (
	topicFlag     string
	ticketFlag    string
	designFlag    string
	researchFlag  string
	structureFlag string
	proposalFlag  string
	tagsFlag      string
	prefixFlag    string
	numberFlag    int
	writeFlag     bool
	forceFlag     bool
)

// typeDirs maps artifact type to its subdirectory under thoughts-dir.
var typeDirs = map[string]string{
	"research":      "research",
	"design":        "designs",
	"plan":          "plans",
	"ticket":        "tickets",
	"ticket-index":  "tickets",
	"structure":     "structures",
	"propose":       "proposals",
	"verify-report": "reviews",
	"spec":          "specs",
}

// validTypes lists all supported artifact types.
var validTypes = []string{
	"research", "design", "plan", "ticket", "ticket-index",
	"structure", "propose", "verify-report", "spec",
}

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold <type> [flags]",
	Short: "Generate artifact files from templates",
	Long: `Generate pre-filled artifact files from .tmpl templates.

Types: research, design, plan, ticket, ticket-index, structure, propose, verify-report, spec

By default, outputs rendered markdown to stdout. Use --write to create the file
at the correct path under .thoughts/.`,
	Args: cobra.ExactArgs(1),
	RunE: runScaffold,
}

func init() {
	addThoughtsDirFlag(scaffoldCmd)
	addTemplatesDirFlag(scaffoldCmd)
	scaffoldCmd.Flags().StringVar(&topicFlag, "topic", "", "Topic/title for the artifact")
	scaffoldCmd.Flags().StringVar(&ticketFlag, "ticket", "", "Ticket ID (e.g., cli-002)")
	scaffoldCmd.Flags().StringVar(&designFlag, "design", "", "Path to design document")
	scaffoldCmd.Flags().StringVar(&researchFlag, "research", "", "Path to research document")
	scaffoldCmd.Flags().StringVar(&structureFlag, "structure", "", "Path to structure document")
	scaffoldCmd.Flags().StringVar(&proposalFlag, "proposal", "", "Path to proposal document")
	scaffoldCmd.Flags().StringVar(&tagsFlag, "tags", "", "Comma-separated tags")
	scaffoldCmd.Flags().StringVar(&prefixFlag, "prefix", "", "Ticket prefix (for ticket/ticket-index types)")
	scaffoldCmd.Flags().IntVar(&numberFlag, "number", 0, "Ticket number (for ticket type)")
	scaffoldCmd.Flags().BoolVar(&writeFlag, "write", false, "Write to file instead of stdout")
	scaffoldCmd.Flags().BoolVar(&forceFlag, "force", false, "Allow overwriting existing files")

	rootCmd.AddCommand(scaffoldCmd)
}

func runScaffold(cmd *cobra.Command, args []string) error {
	artifactType := args[0]

	// Validate type
	if _, ok := typeDirs[artifactType]; !ok {
		fmt.Fprintf(os.Stderr, "error: unknown artifact type %q\nValid types: %v\n", artifactType, validTypes)
		os.Exit(2)
	}

	// Validate required flags per type
	if err := validateRequiredFlags(artifactType); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	// Build render context
	ctx := &tmpl.RenderContext{
		Type:      artifactType,
		Topic:     topicFlag,
		Ticket:    ticketFlag,
		Design:    designFlag,
		Research:  researchFlag,
		Structure: structureFlag,
		Proposal:  proposalFlag,
		Tags:      tagsFlag,
		Prefix:    prefixFlag,
		Number:    numberFlag,
	}

	// Set type label
	typeLabels := map[string]string{
		"research":      "Research",
		"design":        "Design",
		"plan":          "Plan",
		"ticket":        "Ticket",
		"ticket-index":  "Ticket Index",
		"structure":     "Structure",
		"propose":       "Proposal",
		"verify-report": "Verification Report",
		"spec":          "Spec",
	}
	ctx.TypeLabel = typeLabels[artifactType]

	// Resolve auto vars
	if err := tmpl.ResolveAutoVars(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Generate filename
	ctx.Filename = tmpl.GenerateFilename(artifactType, ctx)

	// Render template
	output, err := tmpl.RenderTemplate(artifactType, ctx, templatesDirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if writeFlag {
		return writeOutput(artifactType, ctx.Filename, output)
	}

	fmt.Print(output)
	return nil
}

func validateRequiredFlags(artifactType string) error {
	switch artifactType {
	case "research", "design", "plan", "structure", "propose", "verify-report", "spec":
		if topicFlag == "" {
			return fmt.Errorf("--topic is required for type %q", artifactType)
		}
	case "ticket":
		if topicFlag == "" {
			return fmt.Errorf("--topic is required for type %q", artifactType)
		}
		if ticketFlag == "" {
			return fmt.Errorf("--ticket is required for type %q", artifactType)
		}
	case "ticket-index":
		if topicFlag == "" {
			return fmt.Errorf("--topic is required for type %q", artifactType)
		}
		if prefixFlag == "" {
			return fmt.Errorf("--prefix is required for type %q", artifactType)
		}
	}
	return nil
}

func writeOutput(artifactType, filename, output string) error {
	subdir := typeDirs[artifactType]
	dir := filepath.Join(thoughtsDirFlag, subdir)

	// Create parent dirs if needed
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	outPath := filepath.Join(dir, filename)

	// Check for existing file
	if !forceFlag {
		if _, err := os.Stat(outPath); err == nil {
			fmt.Fprintf(os.Stderr, "error: file already exists: %s (use --force to overwrite)\n", outPath)
			os.Exit(3)
		}
	}

	if err := os.WriteFile(outPath, []byte(output), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(outPath)
	return nil
}
