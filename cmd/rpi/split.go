package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/A-NGJ/rpi/internal/split"
	"github.com/spf13/cobra"
)

var splitScoreThresholdFlag int

var splitScoreCmd = &cobra.Command{
	Use:   "split-score <design-path>",
	Short: "Compute the rpi-plan-split complexity score for a design",
	Long: `Compute the deterministic complexity score used by the rpi-plan skill
to decide whether to propose splitting a design into multiple sibling plans.

The score combines three signals from the design's markdown structure:

  components = number of ` + "`### N. ...`" + ` numbered subheadings under
               ` + "`## Components`" + ` (falls back to ` + "`## File Structure`" + `).
  dirs       = distinct top-level directories referenced in
               ` + "`## File Structure`" + `; root files (CLAUDE.md, README.md, etc.)
               count as one bucket.
  multi_spec = the design references >1 file under ` + "`.rpi/specs/`" + `.

  score = max(0, components-1) + max(0, dirs-1) + (multi_spec ? 2 : 0)

A split is proposed when score ≥ threshold (default 4).

Output is JSON with the per-signal breakdown so the caller can see why the
heuristic landed where it did.`,
	Example: `  # Score a design (default threshold 4)
  rpi split-score .rpi/designs/2026-04-30-rpi-wiki.md

  # Score with a custom threshold
  rpi split-score .rpi/designs/2026-04-30-rpi-wiki.md --threshold 5

  # Sample output:
  # {
  #   "path": ".rpi/designs/2026-04-30-rpi-wiki.md",
  #   "components": {"count": 5, "headings": [...], "source": "components"},
  #   "dirs": {"count": 4, "top_level": [".rpi/", "cmd/", "internal/", "<root>"]},
  #   "specs": {"count": 1, "paths": [".rpi/specs/rpi-wiki.md"]},
  #   "score": {"components_contrib": 4, "dirs_contrib": 3, "multi_spec_contrib": 0, "total": 7},
  #   "threshold": 4,
  #   "should_split": true
  # }`,
	Args: cobra.ExactArgs(1),
	RunE: runSplitScore,
}

func init() {
	splitScoreCmd.Flags().IntVar(&splitScoreThresholdFlag, "threshold", split.DefaultThreshold,
		"Score threshold; should_split is true when score ≥ threshold")
	rootCmd.AddCommand(splitScoreCmd)
}

func runSplitScore(_ *cobra.Command, args []string) error {
	result, err := split.ComputeBytes(args[0], readDesignOrExit(args[0]), splitScoreThresholdFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// readDesignOrExit reads the design file or terminates with a clear error.
// Inlined here (rather than re-using split.Compute) so the file path stays
// captured on the Result for stable JSON output.
func readDesignOrExit(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read design: %v\n", err)
		os.Exit(1)
	}
	return data
}
