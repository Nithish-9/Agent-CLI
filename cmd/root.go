package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"golang.org/x/sys/windows"
)

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "AGENT 0018 — AI Agent CLI",
	Long:  `AGENT 0018 is a AI Agent powered by MCP and LLM models.`,
}

func Execute() {
	if runtime.GOOS == "windows" {
		windows.SetConsoleOutputCP(65001)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	// START Completed Commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringP("file", "f", "", "Path to yaml config file (required)")
	applyCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(stopCmd)
	// END Completed Commands

	/*rootCmd.AddCommand(logsCmd)
	logsCmd.Flags().IntP("lines", "n", 50, "Number of lines to show")
	logsCmd.Flags().StringP("level", "l", "", "Filter by log level (info/warn/error)")

	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getAllCmd)
	getCmd.AddCommand(getModelsCmd)
	getCmd.AddCommand(getMCPServersCmd)
	getCmd.AddCommand(getModelCmd)
	getCmd.AddCommand(getMCPServerCmd)
	getCmd.AddCommand(getStatusCmd)

	getModelsCmd.Flags().BoolP("active", "a", false, "List active models only")
	getModelsCmd.Flags().BoolP("inactive", "i", false, "List inactive models only")

	getMCPServersCmd.Flags().BoolP("active", "a", false, "List active MCP servers only")
	getMCPServersCmd.Flags().BoolP("inactive", "i", false, "List inactive MCP servers only")

	getModelCmd.Flags().StringP("name", "n", "", "Model reference name (required)")
	getModelCmd.MarkFlagRequired("name")

	getMCPServerCmd.Flags().StringP("name", "n", "", "MCP server name (required)")
	getMCPServerCmd.MarkFlagRequired("name")

	/*rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addModelCmd)
	addCmd.AddCommand(addMCPServerCmd)

	addModelCmd.Flags().StringP("ref", "r", "", "Reference name for this model (required)")
	addModelCmd.Flags().StringP("modelname", "m", "", "Actual provider model name (required)")
	addModelCmd.Flags().StringP("endpoint", "e", "", "API endpoint (required)")
	addModelCmd.Flags().StringP("apikey", "k", "", "API key (required)")
	addModelCmd.Flags().IntP("tier", "t", 0, "Model tier (1=planner, 2=executor)")
	addModelCmd.MarkFlagRequired("ref")
	addModelCmd.MarkFlagRequired("modelname")
	addModelCmd.MarkFlagRequired("endpoint")
	addModelCmd.MarkFlagRequired("apikey")

	addMCPServerCmd.Flags().StringP("name", "n", "", "MCP server name (required)")
	addMCPServerCmd.Flags().StringP("transport", "t", "", "Transport type: streamable-http/sse/stdio (required)")
	addMCPServerCmd.Flags().StringP("endpoint", "e", "", "Server endpoint (for http/sse)")
	addMCPServerCmd.MarkFlagRequired("name")
	addMCPServerCmd.MarkFlagRequired("transport")

	rootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(updateModelCmd)
	updateCmd.AddCommand(updateMCPServerCmd)

	updateModelCmd.Flags().StringP("name", "n", "", "Current reference name (required)")
	updateModelCmd.Flags().StringP("ref", "r", "", "New reference name")
	updateModelCmd.Flags().StringP("modelname", "m", "", "New provider model name")
	updateModelCmd.Flags().StringP("endpoint", "e", "", "New API endpoint")
	updateModelCmd.Flags().StringP("apikey", "k", "", "New API key")
	updateModelCmd.Flags().IntP("tier", "t", 0, "New tier (1=planner, 2=executor)")
	updateModelCmd.MarkFlagRequired("name")

	updateMCPServerCmd.Flags().StringP("name", "n", "", "Server name (required)")
	updateMCPServerCmd.Flags().StringP("endpoint", "e", "", "New endpoint")
	updateMCPServerCmd.Flags().StringP("transport", "t", "", "New transport: streamable-http/sse/stdio")
	updateMCPServerCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteModelCmd)
	deleteCmd.AddCommand(deleteMCPServerCmd)

	deleteModelCmd.Flags().StringP("name", "n", "", "Model reference name (required)")
	deleteModelCmd.MarkFlagRequired("name")

	deleteMCPServerCmd.Flags().StringP("name", "n", "", "MCP server name (required)")
	deleteMCPServerCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkModelCmd)
	checkCmd.AddCommand(checkMCPServerCmd)

	checkModelCmd.Flags().StringP("name", "n", "", "Model reference name (required)")
	checkModelCmd.MarkFlagRequired("name")

	checkMCPServerCmd.Flags().StringP("name", "n", "", "MCP server name (required)")
	checkMCPServerCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(setCmd)
	setCmd.AddCommand(setPlannerExecutorCmd)
	setPlannerExecutorCmd.Flags().StringP("plan", "p", "", "Planner model reference name (required)")
	setPlannerExecutorCmd.Flags().StringP("exec", "e", "", "Executor model reference name (required)")
	setPlannerExecutorCmd.MarkFlagRequired("plan")
	setPlannerExecutorCmd.MarkFlagRequired("exec")*/
}
