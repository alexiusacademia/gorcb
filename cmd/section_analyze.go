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
	sectionAnalyzeFile       string
	sectionAnalyzeShowDiagram bool
	sectionAnalyzeExportFile string
)

var sectionAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze moment capacity of a non-rectangular section",
	Long: `Calculate the moment capacity (φMn) of a non-rectangular section
defined in a JSON file.

The analysis uses strain compatibility and force equilibrium to find
the neutral axis position and calculate the moment capacity.

Examples:
  gorcb section analyze --file t-beam.json
  gorcb section analyze -f my-section.json`,
	Run: runSectionAnalyze,
}

func init() {
	sectionCmd.AddCommand(sectionAnalyzeCmd)

	sectionAnalyzeCmd.Flags().StringVarP(&sectionAnalyzeFile, "file", "f", "", "Path to section JSON file [required]")
	sectionAnalyzeCmd.MarkFlagRequired("file")

	// Diagram options
	sectionAnalyzeCmd.Flags().BoolVar(&sectionAnalyzeShowDiagram, "diagram", false, "Show ASCII stress-strain diagram")
	sectionAnalyzeCmd.Flags().StringVarP(&sectionAnalyzeExportFile, "output", "o", "", "Export diagram to file (png, svg, pdf)")
}

func runSectionAnalyze(cmd *cobra.Command, args []string) {
	// Load section from file
	sec, err := section.LoadFromFile(sectionAnalyzeFile)
	if err != nil {
		fmt.Printf("Error loading section: %v\n", err)
		return
	}

	// Run analysis
	result, err := sec.Analyze()
	if err != nil {
		fmt.Printf("Error analyzing section: %v\n", err)
		return
	}

	// Print results
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("     NON-RECTANGULAR SECTION ANALYSIS - NSCP 2015")
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
	fmt.Fprintf(w, "  Vertices:\t%d points\n", len(sec.Vertices))
	w.Flush()
	fmt.Println()

	// Reinforcement
	fmt.Println("REINFORCEMENT:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Layer\tY (mm)\tArea (mm²)\tDescription\n")
	fmt.Fprintf(w, "  ─────\t──────\t──────────\t───────────\n")
	for i, layer := range sec.Reinforcement {
		fmt.Fprintf(w, "  %d\t%.0f\t%.2f\t%s\n", i+1, layer.Y, layer.Area, layer.Description)
	}
	w.Flush()
	fmt.Println()
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Total Tension Steel:\t%.2f mm²\n", result.Properties.TotalTensionSteel)
	if result.Properties.TotalCompressionSteel > 0 {
		fmt.Fprintf(w, "  Total Compression Steel:\t%.2f mm²\n", result.Properties.TotalCompressionSteel)
	}
	w.Flush()
	fmt.Println()

	// Neutral axis analysis
	fmt.Println("NEUTRAL AXIS ANALYSIS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Neutral axis depth (c):\t%.2f mm\n", result.C)
	fmt.Fprintf(w, "  Compression block depth (a):\t%.2f mm\n", result.A)
	fmt.Fprintf(w, "  c/d ratio:\t%.4f\n", result.C/result.Properties.EffectiveDepth)
	fmt.Fprintf(w, "  Compression zone area:\t%.0f mm²\n", result.CompressionArea)
	w.Flush()
	fmt.Println()

	// Steel layer results
	fmt.Println("STEEL LAYER ANALYSIS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Layer\tStrain\tStress (MPa)\tForce (kN)\tStatus\n")
	fmt.Fprintf(w, "  ─────\t──────\t────────────\t──────────\t──────\n")
	for i, layer := range result.SteelLayers {
		status := "Tension"
		if !layer.IsTension {
			status = "Compression"
		}
		if layer.HasYielded {
			status += " (yields)"
		}
		fmt.Fprintf(w, "  %d\t%.6f\t%.2f\t%.2f\t%s\n", 
			i+1, layer.Strain, layer.Stress, layer.Force, status)
	}
	w.Flush()
	fmt.Println()

	// Internal forces
	fmt.Println("INTERNAL FORCES:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Cc (concrete compression):\t%.2f kN\n", result.Cc)
	if result.Cs != 0 {
		fmt.Fprintf(w, "  Cs (compression steel):\t%.2f kN\n", result.Cs)
	}
	fmt.Fprintf(w, "  T (tension steel):\t%.2f kN\n", result.T)
	equilibrium := "✓"
	if absFloat(result.T-(result.Cc+result.Cs)) > 1 {
		equilibrium = "⚠"
	}
	fmt.Fprintf(w, "  Force equilibrium:\t%s\n", equilibrium)
	w.Flush()
	fmt.Println()

	// Capacity
	fmt.Println("MOMENT CAPACITY:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Maximum tensile strain (εt):\t%.6f\n", result.EpsilonT)
	fmt.Fprintf(w, "  Strength reduction factor (φ):\t%.2f\n", result.Phi)
	fmt.Fprintf(w, "  Nominal Moment (Mn):\t%.2f kN-m\n", result.Mn)
	w.Flush()
	fmt.Println()

	fmt.Printf("  ╔═════════════════════════════════════════════════╗\n")
	fmt.Printf("  ║  DESIGN CAPACITY φMn = %.2f kN-m            \n", result.PhiMn)
	fmt.Printf("  ╚═════════════════════════════════════════════════╝\n")
	fmt.Println()

	// Status
	fmt.Println("STATUS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Printf("  %s\n", result.Message)
	fmt.Println()

	// Find tension steel info for diagram
	var tensionSteelY, tensionSteelArea float64
	var compSteelY, compSteelArea float64
	var tensionYields, compYields bool
	epsilonY := sec.Fy / nscp.Es

	for _, layer := range result.SteelLayers {
		if layer.IsTension {
			tensionSteelY = layer.Y
			tensionSteelArea += layer.Area
			if layer.HasYielded {
				tensionYields = true
			}
		} else {
			compSteelY = layer.Y
			compSteelArea += layer.Area
			if layer.HasYielded {
				compYields = true
			}
		}
	}

	// Convert section vertices to diagram points
	var diagramVertices []diagram.Point
	for _, v := range sec.Vertices {
		diagramVertices = append(diagramVertices, diagram.Point{X: v.X, Y: v.Y})
	}

	// Show diagram if requested
	if sectionAnalyzeShowDiagram {
		diagramData := diagram.SectionDiagramData{
			Width:            result.Properties.Width,
			Height:           result.Properties.Height,
			Vertices:         diagramVertices,
			NeutralAxisDepth: result.C,
			StressBlockDepth: result.A,
			TensionSteelY:    tensionSteelY,
			TensionSteelArea: tensionSteelArea,
			CompSteelY:       result.Properties.Height - compSteelY,
			CompSteelArea:    compSteelArea,
			EpsilonCU:        nscp.EpsilonCU,
			EpsilonT:         result.EpsilonT,
			EpsilonY:         epsilonY,
			Fc:               0.85 * sec.Fc,
			FsTension:        sec.Fy,
			FsComp:           sec.Fy,
			TensionYields:    tensionYields,
			CompYields:       compYields,
			IsDoubly:         compSteelArea > 0,
		}

		fmt.Println(diagram.DrawASCIISectionDiagram(diagramData))
		fmt.Println(diagram.DrawStrainDiagram(diagramData))
	}

	// Export diagram if requested
	if sectionAnalyzeExportFile != "" {
		diagramData := diagram.SectionDiagramData{
			Width:            result.Properties.Width,
			Height:           result.Properties.Height,
			Vertices:         diagramVertices,
			NeutralAxisDepth: result.C,
			StressBlockDepth: result.A,
			TensionSteelY:    tensionSteelY,
			TensionSteelArea: tensionSteelArea,
			CompSteelY:       result.Properties.Height - compSteelY,
			CompSteelArea:    compSteelArea,
			EpsilonCU:        nscp.EpsilonCU,
			EpsilonT:         result.EpsilonT,
			EpsilonY:         epsilonY,
			Fc:               0.85 * sec.Fc,
			FsTension:        sec.Fy,
			FsComp:           sec.Fy,
			TensionYields:    tensionYields,
			CompYields:       compYields,
			IsDoubly:         compSteelArea > 0,
		}

		err := diagram.ExportSectionDiagram(diagramData, sectionAnalyzeExportFile)
		if err != nil {
			fmt.Printf("Error exporting diagram: %v\n", err)
		} else {
			fmt.Printf("Diagram exported to: %s\n", sectionAnalyzeExportFile)
		}
	}
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

