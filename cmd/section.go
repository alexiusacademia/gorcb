package cmd

import (
	"github.com/spf13/cobra"
)

var sectionCmd = &cobra.Command{
	Use:   "section",
	Short: "Non-rectangular section design and analysis",
	Long: `Design and analyze non-rectangular concrete sections
defined in JSON files.

This allows analysis of complex shapes like T-beams, L-beams,
or any arbitrary polygonal section.

The section is defined in a JSON file with vertices and reinforcement.

Subcommands:
  analyze  - Calculate moment capacity for a defined section
  design   - Calculate required reinforcement for a given moment

Example JSON file structure:
{
  "name": "T-Beam Section",
  "fc": 28,
  "fy": 415,
  "vertices": [
    {"x": 0, "y": 0},
    {"x": 300, "y": 0},
    {"x": 300, "y": 400},
    {"x": 600, "y": 400},
    {"x": 600, "y": 500},
    {"x": 0, "y": 500}
  ],
  "reinforcement": [
    {"y": 65, "area": 1256.64, "description": "4-20mm"}
  ]
}`,
}

func init() {
	rootCmd.AddCommand(sectionCmd)
}

