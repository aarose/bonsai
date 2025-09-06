package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of bai",
	Long:  `All software has versions. This is bai's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("bai v0.1.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}