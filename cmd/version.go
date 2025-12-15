package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const Version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gorcb",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gorcb v%s\n", Version)
		fmt.Println("Reinforced Concrete Beam Design Tool")
		fmt.Println("Based on NSCP 2015 (National Structural Code of the Philippines)")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

