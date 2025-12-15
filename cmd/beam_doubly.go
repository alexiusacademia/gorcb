package cmd

import (
	"github.com/spf13/cobra"
)

var beamDoublyCmd = &cobra.Command{
	Use:   "doubly",
	Short: "Doubly reinforced rectangular beam design and analysis",
	Long: `Design and analyze doubly reinforced rectangular concrete beams
based on NSCP 2015 provisions.

Doubly reinforced beams have both tension and compression steel.
Use when the required moment exceeds the capacity of a singly reinforced section.

Subcommands:
  design   - Calculate required tension and compression reinforcement
  analyze  - Calculate moment capacity for given reinforcement`,
}

func init() {
	beamCmd.AddCommand(beamDoublyCmd)
}

