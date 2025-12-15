package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/alexiusacademia/gorcb/internal/beam"
	"github.com/alexiusacademia/gorcb/internal/diagram"
	"github.com/alexiusacademia/gorcb/internal/nscp"
	"github.com/spf13/cobra"
)

var (
	// Design inputs
	designWidth  float64
	designHeight float64
	designCover  float64
	designFc     float64
	designFy     float64
	designMu     float64

	// Diagram options
	designShowDiagram bool
	designExportFile  string
)

var beamDesignCmd = &cobra.Command{
	Use:   "design",
	Short: "Design reinforcement for a singly reinforced beam",
	Long: `Calculate the required tension reinforcement area (As) for a
singly reinforced rectangular beam given the factored moment (Mu).

The design follows NSCP 2015 provisions:
  - Section 409.3.2: Strength reduction factors
  - Section 409.6.1.2: Minimum reinforcement
  - Section 410.2.7.3: Equivalent rectangular stress block

Examples:
  # Design a 300x500mm beam with Mu=150 kN-m
  gorcb beam design --width 300 --height 500 --cover 65 --fc 28 --fy 415 --mu 150

  # Using short flags
  gorcb beam design -b 300 -h 500 -c 65 --fc 28 --fy 415 -m 150`,
	Run: runBeamDesign,
}

func init() {
	beamCmd.AddCommand(beamDesignCmd)

	// Geometry flags
	beamDesignCmd.Flags().Float64VarP(&designWidth, "width", "b", 0, "Beam width (mm) [required]")
	beamDesignCmd.Flags().Float64Var(&designHeight, "height", 0, "Beam total depth (mm) [required]")
	beamDesignCmd.Flags().Float64VarP(&designCover, "cover", "c", 65, "Effective cover to steel centroid (mm)")

	// Material flags
	beamDesignCmd.Flags().Float64Var(&designFc, "fc", 28, "Concrete compressive strength f'c (MPa)")
	beamDesignCmd.Flags().Float64Var(&designFy, "fy", 415, "Steel yield strength fy (MPa)")

	// Loading flag
	beamDesignCmd.Flags().Float64VarP(&designMu, "mu", "m", 0, "Factored moment Mu (kN-m) [required]")

	// Mark required flags
	beamDesignCmd.MarkFlagRequired("width")
	beamDesignCmd.MarkFlagRequired("height")
	beamDesignCmd.MarkFlagRequired("mu")

	// Diagram options
	beamDesignCmd.Flags().BoolVar(&designShowDiagram, "diagram", false, "Show ASCII stress-strain diagram")
	beamDesignCmd.Flags().StringVarP(&designExportFile, "output", "o", "", "Export diagram to file (png, svg, pdf)")
}

func runBeamDesign(cmd *cobra.Command, args []string) {
	// Create beam
	b := beam.NewSinglyReinforced(designWidth, designHeight, designCover, designFc, designFy)

	// Run design
	result, err := b.Design(designMu)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print results
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("     SINGLY REINFORCED BEAM DESIGN - NSCP 2015")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Input summary
	fmt.Println("INPUT DATA:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Beam Width (b):\t%.0f mm\n", b.Width)
	fmt.Fprintf(w, "  Beam Depth (h):\t%.0f mm\n", b.Height)
	fmt.Fprintf(w, "  Effective Depth (d):\t%.0f mm\n", b.EffectiveDepth)
	fmt.Fprintf(w, "  Concrete Cover:\t%.0f mm\n", b.Cover)
	fmt.Fprintf(w, "  f'c:\t%.1f MPa\n", b.Fc)
	fmt.Fprintf(w, "  fy:\t%.1f MPa\n", b.Fy)
	fmt.Fprintf(w, "  Factored Moment (Mu):\t%.2f kN-m\n", designMu)
	w.Flush()
	fmt.Println()

	// Reinforcement ratios
	fmt.Println("REINFORCEMENT RATIOS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  ρ_min:\t%.6f\n", result.RhoMin)
	fmt.Fprintf(w, "  ρ_max (tension-controlled):\t%.6f\n", result.RhoMax)
	fmt.Fprintf(w, "  ρ_bal:\t%.6f\n", result.RhoBalanced)
	fmt.Fprintf(w, "  ρ_required:\t%.6f\n", result.RhoRequired)
	w.Flush()
	fmt.Println()

	// Steel area limits
	fmt.Println("STEEL AREA LIMITS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  As,min:\t%.2f mm²\n", result.AsMin)
	fmt.Fprintf(w, "  As,max:\t%.2f mm²\n", result.AsMax)
	w.Flush()
	fmt.Println()

	// Section analysis
	fmt.Println("SECTION ANALYSIS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Compression block depth (a):\t%.2f mm\n", result.A)
	fmt.Fprintf(w, "  Neutral axis depth (c):\t%.2f mm\n", result.C)
	fmt.Fprintf(w, "  Tensile strain (εt):\t%.6f\n", result.EpsilonT)
	fmt.Fprintf(w, "  Strength reduction factor (φ):\t%.2f\n", result.Phi)
	controlStatus := "Tension-controlled"
	if !result.IsTensionControlled {
		controlStatus = "Transition zone"
	}
	fmt.Fprintf(w, "  Section status:\t%s\n", controlStatus)
	w.Flush()
	fmt.Println()

	// Design result
	fmt.Println("DESIGN RESULT:")
	fmt.Println("───────────────────────────────────────────────────────────────")

	if result.IsAdequate {
		fmt.Printf("  ╔═════════════════════════════════════════╗\n")
		fmt.Printf("  ║  REQUIRED As = %.2f mm²              \n", result.AsRequired)
		fmt.Printf("  ╚═════════════════════════════════════════╝\n")
		fmt.Println()
		fmt.Printf("  φMn = %.2f kN-m ≥ Mu = %.2f kN-m ✓\n", result.PhiMn, designMu)
		fmt.Println()
		fmt.Printf("  Status: %s\n", result.Message)
	} else {
		fmt.Println("  ╔═════════════════════════════════════════╗")
		fmt.Println("  ║  DESIGN NOT ADEQUATE                    ║")
		fmt.Println("  ╚═════════════════════════════════════════╝")
		fmt.Println()
		fmt.Printf("  %s\n", result.Message)
	}
	fmt.Println()

	// Suggested bar combinations
	if result.IsAdequate {
		printBarSuggestions(result.AsRequired)
	}

	// Show diagram if requested
	if designShowDiagram && result.IsAdequate {
		epsilonY := designFy / nscp.Es
		tensionYields := result.EpsilonT >= epsilonY

		diagramData := diagram.SectionDiagramData{
			Width:            designWidth,
			Height:           designHeight,
			NeutralAxisDepth: result.C,
			StressBlockDepth: result.A,
			TensionSteelY:    designCover,
			TensionSteelArea: result.AsRequired,
			EpsilonCU:        nscp.EpsilonCU,
			EpsilonT:         result.EpsilonT,
			EpsilonY:         epsilonY,
			Fc:               0.85 * designFc,
			FsTension:        designFy,
			TensionYields:    tensionYields,
			IsDoubly:         false,
		}

		fmt.Println(diagram.DrawASCIISectionDiagram(diagramData))
		fmt.Println(diagram.DrawStrainDiagram(diagramData))
	}

	// Export diagram if requested
	if designExportFile != "" && result.IsAdequate {
		epsilonY := designFy / nscp.Es
		tensionYields := result.EpsilonT >= epsilonY

		diagramData := diagram.SectionDiagramData{
			Width:            designWidth,
			Height:           designHeight,
			NeutralAxisDepth: result.C,
			StressBlockDepth: result.A,
			TensionSteelY:    designCover,
			TensionSteelArea: result.AsRequired,
			EpsilonCU:        nscp.EpsilonCU,
			EpsilonT:         result.EpsilonT,
			EpsilonY:         epsilonY,
			Fc:               0.85 * designFc,
			FsTension:        designFy,
			TensionYields:    tensionYields,
			IsDoubly:         false,
		}

		err := diagram.ExportSectionDiagram(diagramData, designExportFile)
		if err != nil {
			fmt.Printf("Error exporting diagram: %v\n", err)
		} else {
			fmt.Printf("Diagram exported to: %s\n", designExportFile)
		}
	}
}

// Common rebar areas in mm²
var rebarAreas = map[int]float64{
	10: 78.54,   // 10mm diameter
	12: 113.10,  // 12mm diameter
	16: 201.06,  // 16mm diameter
	20: 314.16,  // 20mm diameter
	25: 490.87,  // 25mm diameter
	28: 615.75,  // 28mm diameter
	32: 804.25,  // 32mm diameter
	36: 1017.88, // 36mm diameter
}

func printBarSuggestions(asRequired float64) {
	fmt.Println("SUGGESTED BAR COMBINATIONS:")
	fmt.Println("───────────────────────────────────────────────────────────────")

	suggestions := []struct {
		dia   int
		count int
		area  float64
	}{}

	// Find suitable combinations
	for _, dia := range []int{16, 20, 25, 28, 32} {
		area := rebarAreas[dia]
		count := int(asRequired/area) + 1
		if count >= 2 && count <= 8 {
			totalArea := float64(count) * area
			if totalArea >= asRequired {
				suggestions = append(suggestions, struct {
					dia   int
					count int
					area  float64
				}{dia, count, totalArea})
			}
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Bars\tAs Provided\tRatio\n")
	fmt.Fprintf(w, "  ────\t───────────\t─────\n")

	for _, s := range suggestions {
		ratio := s.area / asRequired
		fmt.Fprintf(w, "  %d - φ%dmm\t%.2f mm²\t%.2f\n", s.count, s.dia, s.area, ratio)
	}
	w.Flush()
	fmt.Println()
}

