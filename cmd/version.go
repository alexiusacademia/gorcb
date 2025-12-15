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
		fmt.Printf("gorcb v%s\n", version.Version)
		fmt.Println("Reinforced Concrete Beam Design Tool")
		fmt.Println("Based on NSCP 2015 (National Structural Code of the Philippines)")

		if version.GitCommit != "unknown" {
			fmt.Printf("Commit: %s\n", version.GitCommit)
		}
		if version.BuildTime != "unknown" {
			fmt.Printf("Built:  %s\n", version.BuildTime)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

