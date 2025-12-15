package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/alexiusacademia/gorcb/internal/beam"
	"github.com/spf13/cobra"
)

var (
	// Doubly design inputs
	doublyDesignWidth     float64
	doublyDesignHeight    float64
	doublyDesignCover     float64
	doublyDesignCoverComp float64
	doublyDesignFc        float64
	doublyDesignFy        float64
	doublyDesignMu        float64
)

var beamDoublyDesignCmd = &cobra.Command{
	Use:   "design",
	Short: "Design reinforcement for a doubly reinforced beam",
	Long: `Calculate the required tension (As) and compression (A'sc) reinforcement
for a doubly reinforced rectangular beam given the factored moment (Mu).

The design follows NSCP 2015 provisions. If the moment can be resisted by
a singly reinforced section, no compression steel will be required.

Examples:
  # Design a 300x500mm beam with Mu=250 kN-m
  gorcb beam doubly design -b 300 --height 500 -c 65 --cover-comp 65 --fc 28 --fy 415 -m 250

  # Using short flags
  gorcb beam doubly design -b 300 --height 500 -c 65 -d 65 --fc 28 --fy 415 -m 250`,
	Run: runDoublyDesign,
}

func init() {
	beamDoublyCmd.AddCommand(beamDoublyDesignCmd)

	// Geometry flags
	beamDoublyDesignCmd.Flags().Float64VarP(&doublyDesignWidth, "width", "b", 0, "Beam width (mm) [required]")
	beamDoublyDesignCmd.Flags().Float64Var(&doublyDesignHeight, "height", 0, "Beam total depth (mm) [required]")
	beamDoublyDesignCmd.Flags().Float64VarP(&doublyDesignCover, "cover", "c", 65, "Effective cover to tension steel centroid (mm)")
	beamDoublyDesignCmd.Flags().Float64VarP(&doublyDesignCoverComp, "cover-comp", "d", 65, "Cover to compression steel centroid d' (mm)")

	// Material flags
	beamDoublyDesignCmd.Flags().Float64Var(&doublyDesignFc, "fc", 28, "Concrete compressive strength f'c (MPa)")
	beamDoublyDesignCmd.Flags().Float64Var(&doublyDesignFy, "fy", 415, "Steel yield strength fy (MPa)")

	// Loading flag
	beamDoublyDesignCmd.Flags().Float64VarP(&doublyDesignMu, "mu", "m", 0, "Factored moment Mu (kN-m) [required]")

	// Mark required flags
	beamDoublyDesignCmd.MarkFlagRequired("width")
	beamDoublyDesignCmd.MarkFlagRequired("height")
	beamDoublyDesignCmd.MarkFlagRequired("mu")
}

func runDoublyDesign(cmd *cobra.Command, args []string) {
	// Create beam
	b := beam.NewDoublyReinforced(
		doublyDesignWidth,
		doublyDesignHeight,
		doublyDesignCover,
		doublyDesignCoverComp,
		doublyDesignFc,
		doublyDesignFy,
	)

	// Run design
	result, err := b.Design(doublyDesignMu)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Print results
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("     DOUBLY REINFORCED BEAM DESIGN - NSCP 2015")
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
	fmt.Fprintf(w, "  Factored Moment (Mu):\t%.2f kN-m\n", doublyDesignMu)
	w.Flush()
	fmt.Println()

	// Reinforcement ratios
	fmt.Println("REINFORCEMENT LIMITS (Singly Reinforced):")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  ρ_min:\t%.6f\n", result.RhoMin)
	fmt.Fprintf(w, "  ρ_max (tension-controlled):\t%.6f\n", result.RhoMax)
	fmt.Fprintf(w, "  ρ_bal:\t%.6f\n", result.RhoBalanced)
	fmt.Fprintf(w, "  As,min:\t%.2f mm²\n", result.AsMin)
	fmt.Fprintf(w, "  As,max (singly):\t%.2f mm²\n", result.AsMax)
	w.Flush()
	fmt.Println()

	// Design type determination
	fmt.Println("DESIGN DETERMINATION:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	Mu1Max := result.Mu1
	if result.RequiresCompSteel {
		// Mu1Max is already set correctly
	} else {
		// For singly reinforced, calculate what Mu1Max would be
		phi := 0.90
		Mu1Max = phi * 0.85 * doublyDesignFc * doublyDesignWidth * result.AMax * (b.EffectiveDepth - result.AMax/2) / 1e6
	}
	fmt.Fprintf(w, "  Max φMn (singly reinforced):\t%.2f kN-m\n", Mu1Max)
	fmt.Fprintf(w, "  Required Mu:\t%.2f kN-m\n", doublyDesignMu)
	if result.RequiresCompSteel {
		fmt.Fprintf(w, "  Design Type:\tDOUBLY REINFORCED REQUIRED\n")
	} else {
		fmt.Fprintf(w, "  Design Type:\tSingly Reinforced Adequate\n")
	}
	w.Flush()
	fmt.Println()

	if result.RequiresCompSteel {
		// Doubly reinforced details
		fmt.Println("MOMENT DISTRIBUTION:")
		fmt.Println("───────────────────────────────────────────────────────────────")
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "  Mu1 (concrete couple):\t%.2f kN-m\n", result.Mu1)
		fmt.Fprintf(w, "  Mu2 (steel couple):\t%.2f kN-m\n", result.Mu2)
		fmt.Fprintf(w, "  Total Mu:\t%.2f kN-m\n", result.Mu1+result.Mu2)
		w.Flush()
		fmt.Println()

		fmt.Println("COMPRESSION STEEL CHECK:")
		fmt.Println("───────────────────────────────────────────────────────────────")
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "  c (at ρmax):\t%.2f mm\n", result.CMax)
		fmt.Fprintf(w, "  d':\t%.2f mm\n", b.CoverComp)
		fmt.Fprintf(w, "  ε'sc:\t%.6f\n", result.EpsilonSc)
		fmt.Fprintf(w, "  εy:\t%.6f\n", doublyDesignFy/200000)
		if result.CompYielded {
			fmt.Fprintf(w, "  Compression steel:\tYIELDS (f'sc = fy = %.1f MPa)\n", doublyDesignFy)
		} else {
			fmt.Fprintf(w, "  Compression steel:\tDOES NOT YIELD (f'sc = %.1f MPa)\n", result.FscStress)
		}
		w.Flush()
		fmt.Println()

		fmt.Println("TENSION STEEL CALCULATION:")
		fmt.Println("───────────────────────────────────────────────────────────────")
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "  As1 (for Mu1):\t%.2f mm²\n", result.As1)
		fmt.Fprintf(w, "  As2 (for Mu2):\t%.2f mm²\n", result.As2)
		w.Flush()
		fmt.Println()
	}

	// Section analysis
	fmt.Println("SECTION STATUS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
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
		fmt.Printf("  ╔═════════════════════════════════════════════════╗\n")
		fmt.Printf("  ║  TENSION STEEL     As  = %.2f mm²           \n", result.AsTotal)
		if result.RequiresCompSteel {
			fmt.Printf("  ║  COMPRESSION STEEL A'sc = %.2f mm²           \n", result.AscRequired)
		}
		fmt.Printf("  ╚═════════════════════════════════════════════════╝\n")
		fmt.Println()
		fmt.Printf("  φMn = %.2f kN-m ≥ Mu = %.2f kN-m ✓\n", result.PhiMn, doublyDesignMu)
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
		fmt.Println("  Tension Steel:")
		printBarSuggestionsFor(result.AsTotal, "    ")
		if result.RequiresCompSteel && result.AscRequired > 0 {
			fmt.Println()
			fmt.Println("  Compression Steel:")
			printBarSuggestionsFor(result.AscRequired, "    ")
		}
	}
}

func printBarSuggestionsFor(asRequired float64, indent string) {
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
	fmt.Fprintf(w, "%sBars\tAs Provided\tRatio\n", indent)
	fmt.Fprintf(w, "%s────\t───────────\t─────\n", indent)

	for _, s := range suggestions {
		ratio := s.area / asRequired
		fmt.Fprintf(w, "%s%d - φ%dmm\t%.2f mm²\t%.2f\n", indent, s.count, s.dia, s.area, ratio)
	}
	w.Flush()
}

