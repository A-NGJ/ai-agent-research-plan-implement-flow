package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/frontmatter"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/scanner"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive operations for .thoughts/ artifacts",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var archiveScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Discover archivable artifacts with reference counts",
	RunE:  runArchiveScan,
}

var archiveCheckRefsCmd = &cobra.Command{
	Use:   "check-refs <path>",
	Short: "Find all files referencing a given path",
	Args:  cobra.ExactArgs(1),
	RunE:  runArchiveCheckRefs,
}

var archiveMoveCmd = &cobra.Command{
	Use:   "move <path>",
	Short: "Archive an artifact: update frontmatter and move to archive/",
	Args:  cobra.ExactArgs(1),
	RunE:  runArchiveMove,
}

var archiveMoveForce bool

type archiveMoveResult struct {
	From               string `json:"from"`
	To                 string `json:"to"`
	FrontmatterUpdated bool   `json:"frontmatter_updated"`
}

// errHasReferences is returned when a file has references and --force is not set.
var errHasReferences = fmt.Errorf("file has active references")

type archiveScanResult struct {
	Path     string  `json:"path"`
	Type     string  `json:"type"`
	Status   *string `json:"status"`
	Title    *string `json:"title"`
	RefCount int     `json:"ref_count"`
}

func init() {
	archiveMoveCmd.Flags().BoolVar(&archiveMoveForce, "force", false, "Skip ref check warning")
	archiveCmd.AddCommand(archiveScanCmd)
	archiveCmd.AddCommand(archiveCheckRefsCmd)
	archiveCmd.AddCommand(archiveMoveCmd)
	rootCmd.AddCommand(archiveCmd)
}

func runArchiveScan(cmd *cobra.Command, args []string) error {
	results, err := scanner.Scan(thoughtsDirFlag, scanner.Filters{Archivable: true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var output []archiveScanResult
	for _, r := range results {
		// Use relative path for ref counting since references are stored as relative paths
		refPath := r.Path
		if rel, err := filepath.Rel(thoughtsDirFlag, r.Path); err == nil {
			refPath = rel
		}
		refCount, err := scanner.CountReferences(thoughtsDirFlag, refPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: counting refs for %s: %v\n", r.Path, err)
			refCount = 0
		}
		output = append(output, archiveScanResult{
			Path:     r.Path,
			Type:     r.Type,
			Status:   r.Status,
			Title:    r.Title,
			RefCount: refCount,
		})
	}

	if output == nil {
		output = []archiveScanResult{}
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

func runArchiveCheckRefs(cmd *cobra.Command, args []string) error {
	refs, err := scanner.FindReferences(thoughtsDirFlag, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(refs, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

func runArchiveMove(cmd *cobra.Command, args []string) error {
	result, err := doArchiveMove(args[0], thoughtsDirFlag, archiveMoveForce, time.Now())
	if err == errHasReferences {
		fmt.Fprintln(os.Stderr, "error: file has active references (use --force to override)")
		os.Exit(3)
	}
	if err != nil {
		return err
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

func doArchiveMove(targetPath, thoughtsDir string, force bool, now time.Time) (*archiveMoveResult, error) {
	doc, err := frontmatter.Parse(targetPath)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", targetPath, err)
	}

	// Check references
	refPath := targetPath
	if rel, relErr := filepath.Rel(thoughtsDir, targetPath); relErr == nil {
		refPath = rel
	}
	refCount, err := scanner.CountReferences(thoughtsDir, refPath)
	if err != nil {
		return nil, fmt.Errorf("count references: %w", err)
	}
	if refCount > 0 && !force {
		return nil, errHasReferences
	}

	// Update frontmatter
	doc.Frontmatter["status"] = "archived"
	doc.Frontmatter["archived_date"] = now.Format("2006-01-02")
	if err := frontmatter.Write(doc); err != nil {
		return nil, fmt.Errorf("write frontmatter: %w", err)
	}

	// Compute destination
	artifactType := scanner.InferType(targetPath)
	yearMonth := now.Format("2006-01")
	filename := filepath.Base(targetPath)
	destDir := filepath.Join(thoughtsDir, "archive", yearMonth, artifactType)
	destPath := filepath.Join(destDir, filename)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("create archive dir: %w", err)
	}
	if err := os.Rename(targetPath, destPath); err != nil {
		return nil, fmt.Errorf("move file: %w", err)
	}

	return &archiveMoveResult{
		From:               targetPath,
		To:                 destPath,
		FrontmatterUpdated: true,
	}, nil
}
