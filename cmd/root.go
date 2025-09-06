package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bai",
	Short: "Bai is a CLI tool",
	Long: `Bai is a command-line interface tool that provides various utilities.
This application is built with love using Cobra in Go.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from bai! Use --help to see available commands.")
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
