package cmd

import (
	"fmt"
	"os"

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

All calculations follow NSCP 2015 (Volume 1) provisions.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		fmt.Println("  ╔═══════════════════════════════════════════════════════════╗")
		fmt.Println("  ║                                                           ║")
		fmt.Println("  ║   gorcb - Go Reinforced Concrete Beam Designer            ║")
		fmt.Println("  ║                                                           ║")
		fmt.Println("  ╚═══════════════════════════════════════════════════════════╝")
		fmt.Println()
		fmt.Println("  A CLI tool for the design of reinforced concrete beams")
		fmt.Println("  based on the National Structural Code of the Philippines (NSCP).")
		fmt.Println()
		fmt.Println("  Features:")
		fmt.Println("    • Factored moment calculation using NSCP load combinations")
		fmt.Println("    • Singly reinforced beam design and analysis")
		fmt.Println("    • Flexural reinforcement calculation")
		fmt.Println()
		fmt.Println("  Use 'gorcb --help' to see available commands.")
		fmt.Println()
		fmt.Println("  ─────────────────────────────────────────────────────────────")
		fmt.Println("  Copyright © 2025 Alexius Academia. All rights reserved.")
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

