package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/alexiusacademia/gorcb/internal/diagram"
	"github.com/alexiusacademia/gorcb/internal/nscp"
	"github.com/alexiusacademia/gorcb/internal/section"
	"github.com/spf13/cobra"
)

var (
	sectionDesignFile       string
	sectionDesignMu         float64
	sectionDesignShowDiagram bool
	sectionDesignExportFile string
)

var sectionDesignCmd = &cobra.Command{
	Use:   "design",
	Short: "Design reinforcement for a non-rectangular section",
	Long: `Calculate the required tension reinforcement for a non-rectangular
section to resist a given factored moment (Mu).

The section geometry and compression reinforcement (if any) are defined
in the JSON file. The design will calculate the required tension steel.

Examples:
  gorcb section design --file t-beam.json --mu 200
  gorcb section design -f my-section.json -m 150`,
	Run: runSectionDesign,
}

func init() {
	sectionCmd.AddCommand(sectionDesignCmd)

	sectionDesignCmd.Flags().StringVarP(&sectionDesignFile, "file", "f", "", "Path to section JSON file [required]")
	sectionDesignCmd.Flags().Float64VarP(&sectionDesignMu, "mu", "m", 0, "Factored moment Mu (kN-m) [required]")

	sectionDesignCmd.MarkFlagRequired("file")
	sectionDesignCmd.MarkFlagRequired("mu")

	// Diagram options
	sectionDesignCmd.Flags().BoolVar(&sectionDesignShowDiagram, "diagram", false, "Show ASCII stress-strain diagram")
	sectionDesignCmd.Flags().StringVarP(&sectionDesignExportFile, "output", "o", "", "Export diagram to file (png, svg, pdf)")
}

func runSectionDesign(cmd *cobra.Command, args []string) {
	// Load section from file
	sec, err := section.LoadFromFile(sectionDesignFile)
	if err != nil {
		fmt.Printf("Error loading section: %v\n", err)
		return
	}

	// Run design
	result, err := sec.Design(sectionDesignMu)
	if err != nil {
		fmt.Printf("Error designing section: %v\n", err)
		return
	}

	// Print results
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("     NON-RECTANGULAR SECTION DESIGN - NSCP 2015")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Section info
	if sec.Name != "" {
		fmt.Printf("  Section: %s\n", sec.Name)
	}
	if sec.Description != "" {
		fmt.Printf("  Description: %s\n", sec.Description)
	}
	fmt.Println()

	// Material properties
	fmt.Println("MATERIAL PROPERTIES:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  f'c:\t%.1f MPa\n", sec.Fc)
	fmt.Fprintf(w, "  fy:\t%.1f MPa\n", sec.Fy)
	fmt.Fprintf(w, "  β₁:\t%.4f\n", result.Beta1)
	w.Flush()
	fmt.Println()

	// Geometric properties
	fmt.Println("SECTION GEOMETRY:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Width (max):\t%.0f mm\n", result.Properties.Width)
	fmt.Fprintf(w, "  Height:\t%.0f mm\n", result.Properties.Height)
	fmt.Fprintf(w, "  Gross Area:\t%.0f mm²\n", result.Properties.Area)
	fmt.Fprintf(w, "  Effective Depth (d):\t%.0f mm\n", result.Properties.EffectiveDepth)
	w.Flush()
	fmt.Println()

	// Design input
	fmt.Println("DESIGN REQUIREMENT:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Factored Moment (Mu):\t%.2f kN-m\n", sectionDesignMu)
	w.Flush()
	fmt.Println()

	// Section analysis at design
	fmt.Println("SECTION AT DESIGN CAPACITY:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Neutral axis depth (c):\t%.2f mm\n", result.C)
	fmt.Fprintf(w, "  Compression block depth (a):\t%.2f mm\n", result.A)
	fmt.Fprintf(w, "  Strength reduction factor (φ):\t%.2f\n", result.Phi)
	w.Flush()
	fmt.Println()

	// Steel area limits
	fmt.Println("REINFORCEMENT LIMITS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  As,min:\t%.2f mm²\n", result.AsMin)
	w.Flush()
	fmt.Println()

	// Design result
	fmt.Println("DESIGN RESULT:")
	fmt.Println("───────────────────────────────────────────────────────────────")

	if result.IsAdequate {
		fmt.Printf("  ╔═════════════════════════════════════════════════╗\n")
		fmt.Printf("  ║  REQUIRED TENSION STEEL As = %.2f mm²       \n", result.AsRequired)
		fmt.Printf("  ╚═════════════════════════════════════════════════╝\n")
		fmt.Println()
		fmt.Printf("  φMn = %.2f kN-m ≥ Mu = %.2f kN-m ✓\n", result.PhiMn, sectionDesignMu)
		fmt.Println()
		fmt.Printf("  Status: %s\n", result.Message)
	} else {
		fmt.Println("  ╔═════════════════════════════════════════════════╗")
		fmt.Println("  ║  DESIGN NOT ADEQUATE                            ║")
		fmt.Println("  ╚═════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Printf("  %s\n", result.Message)
	}
	fmt.Println()

	// Suggested bar combinations
	if result.IsAdequate {
		fmt.Println("SUGGESTED BAR COMBINATIONS:")
		fmt.Println("───────────────────────────────────────────────────────────────")
		printBarSuggestionsFor(result.AsRequired, "  ")
	}

	// Convert section vertices to diagram points
	var diagramVertices []diagram.Point
	for _, v := range sec.Vertices {
		diagramVertices = append(diagramVertices, diagram.Point{X: v.X, Y: v.Y})
	}

	// Show diagram if requested
	if sectionDesignShowDiagram && result.IsAdequate {
		epsilonY := sec.Fy / nscp.Es

		// Get tension steel position from original section
		var tensionSteelY float64
		for _, layer := range sec.Reinforcement {
			if layer.Type == "tension" || layer.Y < result.Properties.Height/2 {
				tensionSteelY = layer.Y
				break
			}
		}

		diagramData := diagram.SectionDiagramData{
			Width:            result.Properties.Width,
			Height:           result.Properties.Height,
			Vertices:         diagramVertices,
			NeutralAxisDepth: result.C,
			StressBlockDepth: result.A,
			TensionSteelY:    tensionSteelY,
			TensionSteelArea: result.AsRequired,
			EpsilonCU:        nscp.EpsilonCU,
			EpsilonT:         0.005, // Tension-controlled by design
			EpsilonY:         epsilonY,
			Fc:               0.85 * sec.Fc,
			FsTension:        sec.Fy,
			TensionYields:    true, // By design
			IsDoubly:         false,
		}

		fmt.Println(diagram.DrawASCIISectionDiagram(diagramData))
		fmt.Println(diagram.DrawStrainDiagram(diagramData))
	}

	// Export diagram if requested
	if sectionDesignExportFile != "" && result.IsAdequate {
		epsilonY := sec.Fy / nscp.Es

		var tensionSteelY float64
		for _, layer := range sec.Reinforcement {
			if layer.Type == "tension" || layer.Y < result.Properties.Height/2 {
				tensionSteelY = layer.Y
				break
			}
		}

		diagramData := diagram.SectionDiagramData{
			Width:            result.Properties.Width,
			Height:           result.Properties.Height,
			Vertices:         diagramVertices,
			NeutralAxisDepth: result.C,
			StressBlockDepth: result.A,
			TensionSteelY:    tensionSteelY,
			TensionSteelArea: result.AsRequired,
			EpsilonCU:        nscp.EpsilonCU,
			EpsilonT:         0.005,
			EpsilonY:         epsilonY,
			Fc:               0.85 * sec.Fc,
			FsTension:        sec.Fy,
			TensionYields:    true,
			IsDoubly:         false,
		}

		err := diagram.ExportSectionDiagram(diagramData, sectionDesignExportFile)
		if err != nil {
			fmt.Printf("Error exporting diagram: %v\n", err)
		} else {
			fmt.Printf("Diagram exported to: %s\n", sectionDesignExportFile)
		}
	}
}

