package cmd

import (
	"fmt"
	"os"

	"github.com/alexiusacademia/gorcb/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gorcb",
	Short: "Reinforced Concrete Beam Design Tool",
	Long: `gorcb - Go Reinforced Concrete Beam Designer

A CLI tool for the design of reinforced concrete beams
based on the National Structural Code of the Philippines (NSCP).

This tool helps structural engineers perform:
  - Flexural design (singly and doubly reinforced beams)
  - Shear design and stirrup spacing
  - Serviceability checks (deflection, crack width)
  - Reinforcement detailing
  - Non-rectangular section analysis

All calculations follow NSCP 2015 (Volume 1) provisions.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		fmt.Println("  ╔═══════════════════════════════════════════════════════════╗")
		fmt.Println("  ║                                                           ║")
		fmt.Printf("  ║   gorcb v%-49s║\n", version.Version)
		fmt.Println("  ║   Go Reinforced Concrete Beam Designer                    ║")
		fmt.Println("  ║   Alexius S. Academia ©  2025                             ║")
		fmt.Println("  ║                                                           ║")
		fmt.Println("  ╚═══════════════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Println("  A CLI tool for the design of reinforced concrete beams")
		fmt.Println("  based on the National Structural Code of the Philippines (NSCP).")
		fmt.Println()
		fmt.Println("  Features:")
		fmt.Println("    • Factored moment calculation using NSCP load combinations")
		fmt.Println("    • Singly reinforced beam design and analysis")
		fmt.Println("    • Doubly reinforced beam design and analysis")
		fmt.Println("    • Non-rectangular section design and analysis")
		fmt.Println()
		fmt.Println("  Use 'gorcb --help' to see available commands.")
		fmt.Println()
		fmt.Println("  ─────────────────────────────────────────────────────────────")
		fmt.Printf("  Copyright © %s %s. All rights reserved.\n", version.Year, version.Author)
		fmt.Println()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

