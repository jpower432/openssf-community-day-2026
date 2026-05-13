// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gemaraproj/go-gemara"
	"github.com/gemaraproj/go-gemara/fetcher"
	"github.com/gemaraproj/go-gemara/gemaraconv"
)

const (
	controlCatalogPath    = "governance/controls/controls.yaml"
	threatCatalogPath     = "governance/catalogs/threat-catalog.yaml"
	capabilityCatalogPath = "governance/catalogs/capabilities.yaml"
	evaluationLogPath     = "governance/evaluation-log.yaml"
	outputDir             = "output"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	ctx := context.Background()
	f := &fetcher.File{}

	fmt.Println("=== Gemara-OSCAL Interoperability Demo ===")
	fmt.Println()

	// --- Load ---

	fmt.Println("Loading Gemara artifacts...")

	catalog, err := gemara.Load[gemara.ControlCatalog](ctx, f, controlCatalogPath)
	if err != nil {
		return fmt.Errorf("loading control catalog: %w", err)
	}
	fmt.Printf("  Control Catalog: %s (%d controls)\n", catalog.Title, len(catalog.Controls))

	threats, err := gemara.Load[gemara.ThreatCatalog](ctx, f, threatCatalogPath)
	if err != nil {
		return fmt.Errorf("loading threat catalog: %w", err)
	}
	fmt.Printf("  Threat Catalog: %s (%d threats)\n", threats.Title, len(threats.Threats))

	capabilities, err := gemara.Load[gemara.CapabilityCatalog](ctx, f, capabilityCatalogPath)
	if err != nil {
		return fmt.Errorf("loading capability catalog: %w", err)
	}
	fmt.Printf("  Capability Catalog: %s (%d capabilities)\n", capabilities.Title, len(capabilities.Capabilities))

	evalLog, err := gemara.Load[gemara.EvaluationLog](ctx, f, evaluationLogPath)
	if err != nil {
		return fmt.Errorf("loading evaluation log: %w", err)
	}
	fmt.Printf("  Evaluation Log: %d evaluations\n", len(evalLog.Evaluations))
	fmt.Println()

	// --- Validate ---

	fmt.Println("Validating cross-references...")
	validateCrossReferences(catalog, threats, capabilities)
	fmt.Println()

	// --- Convert ---

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	fmt.Println("Converting to OSCAL Catalog...")
	oscalCatalog, err := gemaraconv.ControlCatalog(*catalog).ToOSCAL()
	if err != nil {
		return fmt.Errorf("converting to OSCAL catalog: %w", err)
	}
	if err := writeJSON(filepath.Join(outputDir, "oscal-catalog.json"), oscalCatalog); err != nil {
		return fmt.Errorf("writing OSCAL catalog: %w", err)
	}
	fmt.Println("  Wrote oscal-catalog.json")

	fmt.Println("Rendering Markdown...")
	md, err := gemaraconv.ControlCatalog(*catalog).ToMarkdown(ctx)
	if err != nil {
		return fmt.Errorf("rendering markdown: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "controls.md"), md, 0o644); err != nil {
		return fmt.Errorf("writing markdown: %w", err)
	}
	fmt.Println("  Wrote controls.md")

	fmt.Println("Converting EvaluationLog to OSCAL Assessment Results...")
	oscalAR, err := gemaraconv.EvaluationLog(*evalLog).ToOSCALAssessmentResults(gemaraconv.WithCatalog(catalog))
	if err != nil {
		return fmt.Errorf("converting to OSCAL assessment results: %w", err)
	}
	if err := writeJSON(filepath.Join(outputDir, "oscal-assessment-results.json"), oscalAR); err != nil {
		return fmt.Errorf("writing OSCAL assessment results: %w", err)
	}
	fmt.Println("  Wrote oscal-assessment-results.json")

	fmt.Println("Writing assessment summary...")
	if err := writeSummary(filepath.Join(outputDir, "summary.md"), catalog, evalLog); err != nil {
		return fmt.Errorf("writing summary: %w", err)
	}
	fmt.Println("  Wrote summary.md")

	// --- Summary ---

	fmt.Println()
	printSummary(catalog, evalLog)

	return nil
}

func validateCrossReferences(
	catalog *gemara.ControlCatalog,
	threats *gemara.ThreatCatalog,
	_ *gemara.CapabilityCatalog,
) {
	threatIDs := make(map[string]bool)
	for _, t := range threats.Threats {
		threatIDs[t.Id] = true
	}

	for _, control := range catalog.Controls {
		for _, threatMapping := range control.Threats {
			for _, entry := range threatMapping.Entries {
				if threatMapping.ReferenceId == threats.Metadata.Id && !threatIDs[entry.ReferenceId] {
					fmt.Fprintf(os.Stderr, "  WARNING: control %s references unknown threat %s\n",
						control.Id, entry.ReferenceId)
				}
			}
		}
	}

	fmt.Println("  Cross-reference validation complete")
}

func printSummary(catalog *gemara.ControlCatalog, evalLog *gemara.EvaluationLog) {
	fmt.Println("=== Summary ===")
	fmt.Printf("Controls: %d\n", len(catalog.Controls))
	fmt.Printf("Evaluations: %d\n", len(evalLog.Evaluations))

	counts := make(map[string]int)
	for _, eval := range evalLog.Evaluations {
		counts[eval.Result.String()]++
	}
	for result, count := range counts {
		fmt.Printf("  %s: %d\n", result, count)
	}

	fmt.Println()
	fmt.Println("Output files:")
	fmt.Println("  output/oscal-catalog.json")
	fmt.Println("  output/oscal-assessment-results.json")
	fmt.Println("  output/controls.md")
	fmt.Println("  output/summary.md")
}

func writeSummary(path string, catalog *gemara.ControlCatalog, evalLog *gemara.EvaluationLog) error {
	controlTitles := make(map[string]string, len(catalog.Controls))
	for _, c := range catalog.Controls {
		controlTitles[c.Id] = c.Title
	}

	counts := make(map[string]int)
	for _, eval := range evalLog.Evaluations {
		counts[eval.Result.String()]++
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# %s Assessment Summary\n\n", catalog.Title)
	fmt.Fprintf(&buf, "**Target:** %s (%s)  \n", evalLog.Target.Name, evalLog.Target.Environment)
	fmt.Fprintf(&buf, "**Result:** %s  \n", evalLog.Result.String())
	fmt.Fprintf(&buf, "**Catalog Version:** %s  \n", catalog.Metadata.Version)
	fmt.Fprintf(&buf, "\n## Evaluation Results\n\n")
	fmt.Fprintf(&buf, "| Control | Title | Result |\n")
	fmt.Fprintf(&buf, "|:--|:--|:--|\n")

	for _, eval := range evalLog.Evaluations {
		title := controlTitles[eval.Control.EntryId]
		fmt.Fprintf(&buf, "| %s | %s | %s |\n", eval.Control.EntryId, title, eval.Result.String())
	}

	fmt.Fprintf(&buf, "\n**Passed:** %d | **Failed:** %d | **Needs Review:** %d\n",
		counts["Passed"], counts["Failed"], counts["Needs Review"])

	fmt.Fprintf(&buf, "\n## Artifacts\n\n")
	fmt.Fprintf(&buf, "| File | Format | Detail |\n")
	fmt.Fprintf(&buf, "|:--|:--|:--|\n")
	fmt.Fprintf(&buf, "| oscal-catalog.json | OSCAL Catalog | %d controls, %d groups |\n",
		len(catalog.Controls), len(catalog.Groups))
	fmt.Fprintf(&buf, "| oscal-assessment-results.json | OSCAL Assessment Results | %d evaluations |\n",
		len(evalLog.Evaluations))
	fmt.Fprintf(&buf, "| controls.md | Markdown | Full control catalog |\n")

	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON for %s: %w", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}
