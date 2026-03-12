package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/index"
	"github.com/spf13/cobra"
)

var (
	indexLangFlag     string
	indexPathFlag     string
	indexForceFlag    bool
	indexKindFlag     string
	indexExportedFlag bool
	indexFormatFlag   string
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Build and query a codebase symbol index",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var indexBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a symbol index of the codebase",
	RunE:  runIndexBuild,
}

var indexQueryCmd = &cobra.Command{
	Use:   "query <pattern>",
	Short: "Search for symbols in the index",
	Args:  cobra.ExactArgs(1),
	RunE:  runIndexQuery,
}

var indexFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "List all indexed files",
	RunE:  runIndexFiles,
}

var indexStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show index metadata and freshness",
	RunE:  runIndexStatus,
}

func init() {
	indexBuildCmd.Flags().StringVar(&indexLangFlag, "lang", "", "Comma-separated languages to index (e.g., go,py,ts)")
	indexBuildCmd.Flags().StringVar(&indexPathFlag, "path", ".", "Root path to index")
	indexBuildCmd.Flags().BoolVar(&indexForceFlag, "force", false, "Force full rebuild")

	indexQueryCmd.Flags().StringVar(&indexKindFlag, "kind", "", "Filter by symbol kind (function, method, class, struct, interface, type_alias)")
	indexQueryCmd.Flags().BoolVar(&indexExportedFlag, "exported", false, "Show only exported symbols")
	indexQueryCmd.Flags().StringVar(&indexFormatFlag, "format", "json", "Output format: json, md")

	indexFilesCmd.Flags().StringVar(&indexLangFlag, "lang", "", "Filter by language")
	indexFilesCmd.Flags().StringVar(&indexFormatFlag, "format", "json", "Output format: json, md")

	indexStatusCmd.Flags().StringVar(&indexFormatFlag, "format", "text", "Output format: json, text")

	indexCmd.AddCommand(indexBuildCmd)
	indexCmd.AddCommand(indexQueryCmd)
	indexCmd.AddCommand(indexFilesCmd)
	indexCmd.AddCommand(indexStatusCmd)
	rootCmd.AddCommand(indexCmd)
}

func runIndexBuild(cmd *cobra.Command, args []string) error {
	start := time.Now()

	opts := index.BuildOptions{
		ForceRebuild: indexForceFlag,
	}
	if indexLangFlag != "" {
		opts.Languages = strings.Split(indexLangFlag, ",")
	}

	idx, err := index.Build(indexPathFlag, opts)
	if err != nil {
		return fmt.Errorf("build index: %w", err)
	}

	absPath, _ := filepath.Abs(indexPathFlag)
	indexPath := filepath.Join(absPath, index.DefaultIndexPath)
	if err := index.Save(idx, indexPath); err != nil {
		return fmt.Errorf("save index: %w", err)
	}

	if !index.IsGitignored(absPath) {
		fmt.Fprintln(os.Stderr, "warning: .rpi/ is not in .gitignore — index may be accidentally committed")
	}

	elapsed := time.Since(start).Seconds()
	fmt.Fprintf(cmd.OutOrStdout(), "Indexed %d files (%d symbols) in %.1fs. Written to %s\n",
		idx.Metadata.FileCount, idx.Metadata.SymbolCount, elapsed, index.DefaultIndexPath)
	return nil
}

func runIndexQuery(cmd *cobra.Command, args []string) error {
	idx, err := loadIndex()
	if err != nil {
		return err
	}

	results := index.QuerySymbols(idx, index.QueryOptions{
		Pattern:      args[0],
		Kind:         indexKindFlag,
		ExportedOnly: indexExportedFlag,
	})

	if results == nil {
		results = []index.Symbol{}
	}

	switch indexFormatFlag {
	case "md":
		printSymbolsMarkdown(cmd, results)
	default:
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	}
	return nil
}

func runIndexFiles(cmd *cobra.Command, args []string) error {
	idx, err := loadIndex()
	if err != nil {
		return err
	}

	results := index.QueryFiles(idx, indexLangFlag)
	if results == nil {
		results = []index.FileEntry{}
	}

	type fileOutput struct {
		Path     string `json:"path"`
		Language string `json:"language"`
		Symbols  int    `json:"symbols"`
		Size     int64  `json:"size"`
	}

	// Count symbols per file.
	symCounts := make(map[string]int)
	for _, s := range idx.Symbols {
		symCounts[s.File]++
	}

	var out []fileOutput
	for _, f := range results {
		out = append(out, fileOutput{
			Path:     f.Path,
			Language: f.Language,
			Symbols:  symCounts[f.Path],
			Size:     f.Size,
		})
	}
	if out == nil {
		out = []fileOutput{}
	}

	switch indexFormatFlag {
	case "md":
		fmt.Fprintln(cmd.OutOrStdout(), "| Path | Language | Symbols | Size |")
		fmt.Fprintln(cmd.OutOrStdout(), "|---|---|---|---|")
		for _, f := range out {
			fmt.Fprintf(cmd.OutOrStdout(), "| %s | %s | %d | %d |\n", f.Path, f.Language, f.Symbols, f.Size)
		}
	default:
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	}
	return nil
}

func runIndexStatus(cmd *cobra.Command, args []string) error {
	indexPath := index.DefaultIndexPath

	idx, err := index.Load(indexPath)
	if err != nil {
		// Index doesn't exist or is unreadable — not an error for status.
		switch indexFormatFlag {
		case "json":
			fmt.Fprintln(cmd.OutOrStdout(), `{"exists": false}`)
		default:
			fmt.Fprintln(cmd.OutOrStdout(), "No index found. Run 'rpi index build' to create one.")
		}
		return nil
	}

	result := index.Status(idx, idx.Metadata.RootPath)
	result.IndexPath = indexPath
	if info, err := os.Stat(indexPath); err == nil {
		result.IndexSizeBytes = info.Size()
	}

	switch indexFormatFlag {
	case "json":
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "Index: %s\n", result.IndexPath)
		fmt.Fprintf(cmd.OutOrStdout(), "Built: %s (%.0fs ago)\n", result.BuiltAt.Format(time.RFC3339), float64(result.AgeSeconds))
		fmt.Fprintf(cmd.OutOrStdout(), "Files: %d (%d stale)\n", result.FileCount, result.StaleFiles)
		fmt.Fprintf(cmd.OutOrStdout(), "Symbols: %d\n", result.SymbolCount)
		langs := make([]string, 0, len(result.Languages))
		for l, c := range result.Languages {
			langs = append(langs, fmt.Sprintf("%s: %d", l, c))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Languages: %s\n", strings.Join(langs, ", "))
	}
	return nil
}

func loadIndex() (*index.Index, error) {
	idx, err := index.Load(index.DefaultIndexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: no index found — run 'rpi index build' first\n")
		os.Exit(1)
	}
	return idx, nil
}

func printSymbolsMarkdown(cmd *cobra.Command, syms []index.Symbol) {
	fmt.Fprintln(cmd.OutOrStdout(), "| Name | Kind | File | Line | Package | Signature |")
	fmt.Fprintln(cmd.OutOrStdout(), "|---|---|---|---|---|---|")
	for _, s := range syms {
		fmt.Fprintf(cmd.OutOrStdout(), "| %s | %s | %s | %d | %s | %s |\n",
			s.Name, s.Kind, s.File, s.Line, s.Package, s.Signature)
	}
}
