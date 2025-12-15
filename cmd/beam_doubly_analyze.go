package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/alexiusacademia/gorcb/internal/beam"
	"github.com/spf13/cobra"
)

var (
	// Doubly analysis inputs
	doublyAnalyzeWidth     float64
	doublyAnalyzeHeight    float64
	doublyAnalyzeCover     float64
	doublyAnalyzeCoverComp float64
	doublyAnalyzeFc        float64
	doublyAnalyzeFy        float64
	doublyAnalyzeAs        float64
	doublyAnalyzeAsc       float64
)

var beamDoublyAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze moment capacity of a doubly reinforced beam",
	Long: `Calculate the moment capacity (φMn) of a doubly reinforced 
rectangular beam given the tension (As) and compression (A'sc) reinforcement.

The analysis follows NSCP 2015 provisions and accounts for:
  - Compression steel stress (yield or elastic)
  - Displaced concrete by compression steel
  - Strain compatibility

Examples:
  # Analyze a 300x500mm beam with As=1500 mm² and A'sc=600 mm²
  gorcb beam doubly analyze -b 300 --height 500 -c 65 -d 65 --fc 28 --fy 415 --as 1500 --asc 600`,
	Run: runDoublyAnalyze,
}

func init() {
	beamDoublyCmd.AddCommand(beamDoublyAnalyzeCmd)

	// Geometry flags
	beamDoublyAnalyzeCmd.Flags().Float64VarP(&doublyAnalyzeWidth, "width", "b", 0, "Beam width (mm) [required]")
	beamDoublyAnalyzeCmd.Flags().Float64Var(&doublyAnalyzeHeight, "height", 0, "Beam total depth (mm) [required]")
	beamDoublyAnalyzeCmd.Flags().Float64VarP(&doublyAnalyzeCover, "cover", "c", 65, "Effective cover to tension steel centroid (mm)")
	beamDoublyAnalyzeCmd.Flags().Float64VarP(&doublyAnalyzeCoverComp, "cover-comp", "d", 65, "Cover to compression steel centroid d' (mm)")

	// Material flags
	beamDoublyAnalyzeCmd.Flags().Float64Var(&doublyAnalyzeFc, "fc", 28, "Concrete compressive strength f'c (MPa)")
	beamDoublyAnalyzeCmd.Flags().Float64Var(&doublyAnalyzeFy, "fy", 415, "Steel yield strength fy (MPa)")

	// Reinforcement flags
	beamDoublyAnalyzeCmd.Flags().Float64Var(&doublyAnalyzeAs, "as", 0, "Tension reinforcement area As (mm²) [required]")
	beamDoublyAnalyzeCmd.Flags().Float64Var(&doublyAnalyzeAsc, "asc", 0, "Compression reinforcement area A'sc (mm²) [required]")

	// Mark required flags
	beamDoublyAnalyzeCmd.MarkFlagRequired("width")
	beamDoublyAnalyzeCmd.MarkFlagRequired("height")
	beamDoublyAnalyzeCmd.MarkFlagRequired("as")
	beamDoublyAnalyzeCmd.MarkFlagRequired("asc")
}

func runDoublyAnalyze(cmd *cobra.Command, args []string) {
	// Create beam
	b := beam.NewDoublyReinforced(
		doublyAnalyzeWidth,
		doublyAnalyzeHeight,
		doublyAnalyzeCover,
		doublyAnalyzeCoverComp,
		doublyAnalyzeFc,
		doublyAnalyzeFy,
	)

	// Run analysis
	result, err := b.Analyze(doublyAnalyzeAs, doublyAnalyzeAsc)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print results
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("     DOUBLY REINFORCED BEAM ANALYSIS - NSCP 2015")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Input summary
	fmt.Println("INPUT DATA:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Beam Width (b):\t%.0f mm\n", b.Width)
	fmt.Fprintf(w, "  Beam Depth (h):\t%.0f mm\n", b.Height)
	fmt.Fprintf(w, "  Effective Depth (d):\t%.0f mm\n", b.EffectiveDepth)
	fmt.Fprintf(w, "  Tension Cover:\t%.0f mm\n", b.Cover)
	fmt.Fprintf(w, "  Compression Cover (d'):\t%.0f mm\n", b.CoverComp)
	fmt.Fprintf(w, "  f'c:\t%.1f MPa\n", b.Fc)
	fmt.Fprintf(w, "  fy:\t%.1f MPa\n", b.Fy)
	fmt.Fprintf(w, "  Tension Steel (As):\t%.2f mm²\n", doublyAnalyzeAs)
	fmt.Fprintf(w, "  Compression Steel (A'sc):\t%.2f mm²\n", doublyAnalyzeAsc)
	w.Flush()
	fmt.Println()

	// Reinforcement ratios
	fmt.Println("REINFORCEMENT RATIOS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  ρ_min:\t%.6f\n", result.RhoMin)
	fmt.Fprintf(w, "  ρ_max (tension-controlled):\t%.6f\n", result.RhoMax)
	fmt.Fprintf(w, "  ρ_bal:\t%.6f\n", result.RhoBalanced)
	fmt.Fprintf(w, "  ρ_tension (As/bd):\t%.6f", result.Rho)
	if !result.MeetsMinReinf {
		fmt.Fprintf(w, " ⚠ (< ρ_min)")
	} else {
		fmt.Fprintf(w, " ✓")
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  ρ_compression (A'sc/bd):\t%.6f\n", result.RhoComp)
	w.Flush()
	fmt.Println()

	// Section properties
	fmt.Println("SECTION PROPERTIES:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  β₁:\t%.4f\n", result.Beta1)
	fmt.Fprintf(w, "  Neutral axis depth (c):\t%.2f mm\n", result.C)
	fmt.Fprintf(w, "  Compression block depth (a):\t%.2f mm\n", result.A)
	fmt.Fprintf(w, "  c/d ratio:\t%.4f\n", result.C/b.EffectiveDepth)
	w.Flush()
	fmt.Println()

	// Strain analysis
	fmt.Println("STRAIN ANALYSIS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  εcu (concrete):\t0.003000\n")
	fmt.Fprintf(w, "  εy (steel yield):\t%.6f\n", doublyAnalyzeFy/200000)
	fmt.Fprintf(w, "  εt (tension steel):\t%.6f", result.EpsilonT)
	if result.TensionYielded {
		fmt.Fprintf(w, " → YIELDS")
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  ε'sc (compression steel):\t%.6f", result.EpsilonSc)
	if result.CompYielded {
		fmt.Fprintf(w, " → YIELDS")
	}
	fmt.Fprintln(w)
	w.Flush()
	fmt.Println()

	// Steel stresses
	fmt.Println("STEEL STRESSES:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  fs (tension):\t%.2f MPa\n", result.FsStress)
	fmt.Fprintf(w, "  f'sc (compression):\t%.2f MPa\n", result.FscStress)
	w.Flush()
	fmt.Println()

	// Internal forces
	fmt.Println("INTERNAL FORCES:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Cc (concrete compression):\t%.2f kN\n", result.Cc)
	fmt.Fprintf(w, "  Cs (compression steel):\t%.2f kN\n", result.Cs)
	fmt.Fprintf(w, "  T (tension steel):\t%.2f kN\n", result.T)
	fmt.Fprintf(w, "  ΣC = Cc + Cs:\t%.2f kN\n", result.Cc+result.Cs)
	equilibrium := "✓"
	if abs(result.T-(result.Cc+result.Cs)) > 1 {
		equilibrium = "⚠"
	}
	fmt.Fprintf(w, "  Force equilibrium:\t%s\n", equilibrium)
	w.Flush()
	fmt.Println()

	// Moment capacity
	fmt.Println("MOMENT CAPACITY:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  Nominal Moment (Mn):\t%.2f kN-m\n", result.Mn)
	fmt.Fprintf(w, "  Strength reduction factor (φ):\t%.2f\n", result.Phi)
	w.Flush()
	fmt.Println()

	fmt.Printf("  ╔═════════════════════════════════════════════════╗\n")
	fmt.Printf("  ║  DESIGN CAPACITY φMn = %.2f kN-m            \n", result.PhiMn)
	fmt.Printf("  ╚═════════════════════════════════════════════════╝\n")
	fmt.Println()

	// Status
	fmt.Println("STATUS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	controlStatus := "Tension-controlled (φ = 0.90)"
	if !result.IsTensionControlled {
		if result.EpsilonT >= doublyAnalyzeFy/200000 {
			controlStatus = fmt.Sprintf("Transition zone (φ = %.2f)", result.Phi)
		} else {
			controlStatus = "Compression-controlled (φ = 0.65)"
		}
	}
	fmt.Printf("  Section: %s\n", controlStatus)
	fmt.Printf("  %s\n", result.Message)
	fmt.Println()
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

