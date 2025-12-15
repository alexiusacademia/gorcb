package cmd

import (
	"github.com/spf13/cobra"
)

var beamCmd = &cobra.Command{
	Use:   "beam",
	Short: "Singly reinforced rectangular beam design and analysis",
	Long: `Design and analyze singly reinforced rectangular concrete beams
based on NSCP 2015 provisions.

Subcommands:
  design   - Calculate required reinforcement for a given moment
  analyze  - Calculate moment capacity for a given reinforcement

All calculations follow NSCP 2015 strength design method.`,
}

func init() {
	rootCmd.AddCommand(beamCmd)
}

