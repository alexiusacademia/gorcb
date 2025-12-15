package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/alexiusacademia/gorcb/internal/nscp"
	"github.com/spf13/cobra"
)

var (
	// Unfactored moments (kN-m)
	momentDead       float64
	momentLive       float64
	momentRoof       float64
	momentWind       float64
	momentEarthquake float64
	momentRain       float64

	// Options
	showAll      bool
	useSimplified bool
)

var momentCmd = &cobra.Command{
	Use:   "moment",
	Short: "Calculate factored moment using NSCP load combinations",
	Long: `Calculate the factored moment (Mu) based on NSCP 2015 load combinations.

Provide unfactored moments from different load types and this command will
compute the factored moments for all applicable NSCP load combinations.

Load Types:
  D  - Dead load
  L  - Live load  
  Lr - Roof live load
  W  - Wind load
  E  - Earthquake load
  R  - Rain load

Examples:
  # Simple gravity loads (dead + live)
  gorcb moment --dead 50 --live 30

  # With wind load
  gorcb moment --dead 50 --live 30 --wind 20

  # Show all combinations
  gorcb moment --dead 50 --live 30 --all`,
	Run: runMoment,
}

func init() {
	rootCmd.AddCommand(momentCmd)

	// Load moment flags
	momentCmd.Flags().Float64VarP(&momentDead, "dead", "d", 0, "Moment due to dead load (kN-m)")
	momentCmd.Flags().Float64VarP(&momentLive, "live", "l", 0, "Moment due to live load (kN-m)")
	momentCmd.Flags().Float64VarP(&momentRoof, "roof", "r", 0, "Moment due to roof live load (kN-m)")
	momentCmd.Flags().Float64VarP(&momentWind, "wind", "w", 0, "Moment due to wind load (kN-m)")
	momentCmd.Flags().Float64VarP(&momentEarthquake, "earthquake", "e", 0, "Moment due to earthquake load (kN-m)")
	momentCmd.Flags().Float64VarP(&momentRain, "rain", "R", 0, "Moment due to rain load (kN-m)")

	// Options
	momentCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all load combination results")
	momentCmd.Flags().BoolVarP(&useSimplified, "simplified", "s", false, "Use simplified combinations (gravity only: 1.4D and 1.2D+1.6L)")
}

func runMoment(cmd *cobra.Command, args []string) {
	moments := nscp.LoadMoments{
		Dead:       momentDead,
		Live:       momentLive,
		Roof:       momentRoof,
		Wind:       momentWind,
		Earthquake: momentEarthquake,
		Rain:       momentRain,
	}

	// Check if any moment is provided
	if moments.Dead == 0 && moments.Live == 0 && moments.Roof == 0 &&
		moments.Wind == 0 && moments.Earthquake == 0 && moments.Rain == 0 {
		fmt.Println("Error: Please provide at least one unfactored moment.")
		fmt.Println("Use 'gorcb moment --help' for usage information.")
		return
	}

	// Select which combinations to use
	combinations := nscp.LoadCombinations
	if useSimplified {
		combinations = nscp.SimplifiedCombinations
	}

	// Print header
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("          NSCP 2015 FACTORED MOMENT CALCULATION")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Print input moments
	fmt.Println("UNFACTORED MOMENTS (kN-m):")
	fmt.Println("───────────────────────────────────────────────────────────────")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if moments.Dead != 0 {
		fmt.Fprintf(w, "  Dead Load (D):\t%.2f\n", moments.Dead)
	}
	if moments.Live != 0 {
		fmt.Fprintf(w, "  Live Load (L):\t%.2f\n", moments.Live)
	}
	if moments.Roof != 0 {
		fmt.Fprintf(w, "  Roof Live Load (Lr):\t%.2f\n", moments.Roof)
	}
	if moments.Wind != 0 {
		fmt.Fprintf(w, "  Wind Load (W):\t%.2f\n", moments.Wind)
	}
	if moments.Earthquake != 0 {
		fmt.Fprintf(w, "  Earthquake Load (E):\t%.2f\n", moments.Earthquake)
	}
	if moments.Rain != 0 {
		fmt.Fprintf(w, "  Rain Load (R):\t%.2f\n", moments.Rain)
	}
	w.Flush()
	fmt.Println()

	// Calculate governing moment
	maxMu, governingCombo := nscp.CalculateGoverningMoment(moments, combinations)

	if showAll {
		// Show all combinations
		fmt.Println("LOAD COMBINATIONS (NSCP 2015 Section 203.3):")
		fmt.Println("───────────────────────────────────────────────────────────────")
		w = tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "  #\tCombination\tMu (kN-m)\n")
		fmt.Fprintf(w, "  ─\t───────────\t─────────\n")

		for _, combo := range combinations {
			mu := combo.CalculateFactoredMoment(moments)
			marker := ""
			if combo.ID == governingCombo.ID {
				marker = " ← GOVERNS"
			}
			fmt.Fprintf(w, "  %s\t%s\t%.2f%s\n", combo.ID, combo.Description, mu, marker)
		}
		w.Flush()
		fmt.Println()
	}

	// Print result
	fmt.Println("RESULT:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	fmt.Printf("  Governing Combination: %s (%s)\n", governingCombo.ID, governingCombo.Description)
	fmt.Println()
	fmt.Printf("  ╔═══════════════════════════════════╗\n")
	fmt.Printf("  ║  FACTORED MOMENT (Mu) = %.2f kN-m  \n", maxMu)
	fmt.Printf("  ╚═══════════════════════════════════╝\n")
	fmt.Println()
}

