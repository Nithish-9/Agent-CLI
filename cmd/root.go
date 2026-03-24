package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "AGENT 0018 — AI Agent CLI",
	Long:  `AGENT 0018 is a AI Agent powered by MCP and LLM models.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringP("file", "f", "", "Path to yaml config file (required)")
	applyCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getStatusCmd)
}
