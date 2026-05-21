package main

import (
	"encoding/json"
	"fmt"

	"github.com/A-NGJ/rpi/internal/specdrift"
	"github.com/spf13/cobra"
)

var (
	staleDaysFlag int
	ratioLowFlag  float64
	ratioHighFlag float64
	specsDirFlag  string
)

var specDriftCmd = &cobra.Command{
	Use:   "spec-drift <action>",
	Short: "Deterministic structural drift checks for .rpi/specs/",
	Long: `Compute deterministic structural drift signals over the specs directory.

Actions:
  scan  Walk .rpi/specs/*.md and emit a per-spec list of fired signals.

Signals reported:
  stale_last_updated      last_updated is older than --stale-days AND a
                          referenced file has git activity since.
  scenario_count_mismatch scenarios/impl-refs ratio is outside the
                          [--ratio-low, --ratio-high] window.
  broken_references       markdown link targets that no longer exist.
  naming_mismatch         filename slug differs from the slugified
                          'feature' frontmatter field.
  orphaned                no incoming references from any other .rpi/
                          artifact. Suppressed by frontmatter 'orphaned: false'.

The tool is read-only and deterministic — same inputs produce the same
JSON output across runs. It runs on a stock checkout with no external
services. When git is unavailable, stale_last_updated still fires with
details.git = "unavailable".`,
	Example: `  # Scan with default thresholds
  rpi spec-drift scan

  # Treat anything older than two weeks as stale
  rpi spec-drift scan --stale-days 14

  # Point at a non-default specs directory
  rpi spec-drift scan --specs-dir other/specs

  # Sample output:
  # [
  #   {
  #     "path": ".rpi/specs/foo.md",
  #     "signals": [
  #       {"name": "naming_mismatch", "details": {"expected_filename": "bar.md"}}
  #     ]
  #   }
  # ]`,
}

var specDriftScanCmd = &cobra.Command{
	Use:     "scan",
	Short:   "Emit per-spec drift signals as JSON",
	Long:    specDriftCmd.Long,
	Example: specDriftCmd.Example,
	RunE:    runSpecDriftScan,
}

func init() {
	specDriftScanCmd.Flags().IntVar(&staleDaysFlag, "stale-days", 30, "spec is stale when last_updated exceeds this many days")
	specDriftScanCmd.Flags().Float64Var(&ratioLowFlag, "ratio-low", 0.5, "scenarios/refs ratio below this fires scenario_count_mismatch")
	specDriftScanCmd.Flags().Float64Var(&ratioHighFlag, "ratio-high", 3.0, "scenarios/refs ratio above this fires scenario_count_mismatch")
	specDriftScanCmd.Flags().StringVar(&specsDirFlag, "specs-dir", ".rpi/specs", "directory containing spec files")
	addFormatFlag(specDriftScanCmd)
	specDriftCmd.AddCommand(specDriftScanCmd)
	rootCmd.AddCommand(specDriftCmd)
}

func runSpecDriftScan(cmd *cobra.Command, args []string) error {
	opts := specdrift.ScanOptions{
		StaleDays: staleDaysFlag,
		RatioLow:  ratioLowFlag,
		RatioHigh: ratioHighFlag,
		SpecsDir:  specsDirFlag,
	}
	records, err := specdrift.Scan(opts)
	if err != nil {
		return fmt.Errorf("spec-drift scan: %w", err)
	}
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal results: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}
