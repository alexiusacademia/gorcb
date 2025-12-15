package cmd

import (
	"fmt"

	"github.com/alexiusacademia/gorcb/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gorcb",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		fmt.Printf("  gorcb v%s\n", version.Version)
		fmt.Println("  ─────────────────────────────────────────")
		fmt.Println("  Reinforced Concrete Beam Design Tool")
		fmt.Println("  Based on NSCP 2015 (National Structural Code of the Philippines)")
		fmt.Println()

		if version.GitCommit != "unknown" {
			fmt.Printf("  Commit:  %s\n", version.GitCommit)
		}
		if version.BuildTime != "unknown" {
			fmt.Printf("  Built:   %s\n", version.BuildTime)
		}

		fmt.Printf("  Author:  %s\n", version.Author)
		fmt.Printf("  © %s %s. All rights reserved.\n", version.Year, version.Author)
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

