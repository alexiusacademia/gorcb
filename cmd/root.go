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
		fmt.Println("Welcome to gorcb - Reinforced Concrete Beam Designer")
		fmt.Println()
		fmt.Println("Use 'gorcb --help' to see available commands.")
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

