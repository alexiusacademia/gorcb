package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/alexiusacademia/gorcb/internal/beam"
	"github.com/spf13/cobra"
)

var (
	// Analysis inputs
	analyzeWidth  float64
	analyzeHeight float64
	analyzeCover  float64
	analyzeFc     float64
	analyzeFy     float64
	analyzeAs     float64
)

var beamAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze moment capacity of a singly reinforced beam",
	Long: `Calculate the moment capacity (φMn) of a singly reinforced 
rectangular beam given the tension reinforcement area (As).

The analysis follows NSCP 2015 provisions:
  - Section 409.3.2: Strength reduction factors
  - Section 409.6.1.2: Minimum reinforcement check
  - Section 410.2.7.3: Equivalent rectangular stress block

Examples:
  # Analyze a 300x500mm beam with 3-20mm bars (As = 942 mm²)
  gorcb beam analyze --width 300 --height 500 --cover 65 --fc 28 --fy 415 --as 942

  # Using short flags
  gorcb beam analyze -b 300 -h 500 -c 65 --fc 28 --fy 415 -a 942`,
	Run: runBeamAnalyze,
}

func init() {
	beamCmd.AddCommand(beamAnalyzeCmd)

	// Geometry flags
	beamAnalyzeCmd.Flags().Float64VarP(&analyzeWidth, "width", "b", 0, "Beam width (mm) [required]")
	beamAnalyzeCmd.Flags().Float64Var(&analyzeHeight, "height", 0, "Beam total depth (mm) [required]")
	beamAnalyzeCmd.Flags().Float64VarP(&analyzeCover, "cover", "c", 65, "Effective cover to steel centroid (mm)")

	// Material flags
	beamAnalyzeCmd.Flags().Float64Var(&analyzeFc, "fc", 28, "Concrete compressive strength f'c (MPa)")
	beamAnalyzeCmd.Flags().Float64Var(&analyzeFy, "fy", 415, "Steel yield strength fy (MPa)")

	// Reinforcement flag
	beamAnalyzeCmd.Flags().Float64VarP(&analyzeAs, "as", "a", 0, "Tension reinforcement area As (mm²) [required]")

	// Mark required flags
	beamAnalyzeCmd.MarkFlagRequired("width")
	beamAnalyzeCmd.MarkFlagRequired("height")
	beamAnalyzeCmd.MarkFlagRequired("as")
}

func runBeamAnalyze(cmd *cobra.Command, args []string) {
	// Create beam
	b := beam.NewSinglyReinforced(analyzeWidth, analyzeHeight, analyzeCover, analyzeFc, analyzeFy)

	// Run analysis
	result, err := b.Analyze(analyzeAs)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print results
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("     SINGLY REINFORCED BEAM ANALYSIS - NSCP 2015")
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
	fmt.Fprintf(w, "  Reinforcement (As):\t%.2f mm²\n", analyzeAs)
	w.Flush()
	fmt.Println()

	// Reinforcement ratios
	fmt.Println("REINFORCEMENT RATIOS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  ρ_min:\t%.6f\n", result.RhoMin)
	fmt.Fprintf(w, "  ρ_max (tension-controlled):\t%.6f\n", result.RhoMax)
	fmt.Fprintf(w, "  ρ_bal:\t%.6f\n", result.RhoBalanced)
	fmt.Fprintf(w, "  ρ_actual:\t%.6f", result.Rho)

	// Add status indicator for rho
	if !result.MeetsMinReinf {
		fmt.Fprintf(w, " ⚠ (< ρ_min)")
	} else if !result.MeetsMaxReinf {
		fmt.Fprintf(w, " ⚠ (> ρ_max)")
	} else {
		fmt.Fprintf(w, " ✓")
	}
	fmt.Fprintln(w)
	w.Flush()
	fmt.Println()

	// Steel area limits
	fmt.Println("STEEL AREA LIMITS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	asMin := result.RhoMin * analyzeWidth * (analyzeHeight - analyzeCover)
	asMax := result.RhoMax * analyzeWidth * (analyzeHeight - analyzeCover)
	fmt.Fprintf(w, "  As,min:\t%.2f mm²\n", asMin)
	fmt.Fprintf(w, "  As,max:\t%.2f mm²\n", asMax)
	fmt.Fprintf(w, "  As,provided:\t%.2f mm²\n", analyzeAs)
	w.Flush()
	fmt.Println()

	// Section analysis
	fmt.Println("SECTION PROPERTIES:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  β₁:\t%.4f\n", result.Beta1)
	fmt.Fprintf(w, "  Compression block depth (a):\t%.2f mm\n", result.A)
	fmt.Fprintf(w, "  Neutral axis depth (c):\t%.2f mm\n", result.C)
	fmt.Fprintf(w, "  c/d ratio:\t%.4f\n", result.C/(analyzeHeight-analyzeCover))
	fmt.Fprintf(w, "  Tensile strain (εt):\t%.6f\n", result.EpsilonT)
	fmt.Fprintf(w, "  Strength reduction factor (φ):\t%.2f\n", result.Phi)
	w.Flush()
	fmt.Println()

	// Moment capacity
	fmt.Println("MOMENT CAPACITY:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Nominal Moment (Mn):\t%.2f kN-m\n", result.Mn)
	w.Flush()
	fmt.Println()

	fmt.Printf("  ╔═════════════════════════════════════════╗\n")
	fmt.Printf("  ║  DESIGN CAPACITY φMn = %.2f kN-m     \n", result.PhiMn)
	fmt.Printf("  ╚═════════════════════════════════════════╝\n")
	fmt.Println()

	// Status
	fmt.Println("STATUS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	controlStatus := "Tension-controlled (φ = 0.90)"
	if !result.IsTensionControlled {
		if result.EpsilonT >= analyzeFy/200000 {
			controlStatus = fmt.Sprintf("Transition zone (φ = %.2f)", result.Phi)
		} else {
			controlStatus = "Compression-controlled (φ = 0.65)"
		}
	}
	fmt.Printf("  Section: %s\n", controlStatus)
	fmt.Printf("  %s\n", result.Message)
	fmt.Println()
}

