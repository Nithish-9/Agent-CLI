package cmd

import (
	"bufio"
	"fmt"
	"os"
	"salesforce-ai-agent/configuration"
	"salesforce-ai-agent/daemon"
	resource "salesforce-ai-agent/resources"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show agent version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(" AGENT 0018 v1.0.0")
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the agent",
	Run: func(cmd *cobra.Command, args []string) {
		if IsDaemonRunning() {
			fmt.Println(" Agent 0018 is already running")
			return
		}
		fmt.Println(" Starting Agent 0018 ...")
		if err := daemon.Start(); err != nil {
			fmt.Println(" Failed to start agent:", err)
			return
		}
	},
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a yaml config file",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsDaemonRunning() {
			fmt.Println(" Agent is not started. Run 'agent start' first")
			return
		}

		file, _ := cmd.Flags().GetString("file")
		configYaml, err := configuration.LoadYamlFile(file)
		if err != nil {
			fmt.Println(" Error loading yaml file:", err)
			return
		}

		err = SendConfig(configYaml)
		if err != nil {
			fmt.Println(" Failed to send config to daemon:", err)
			return
		}

		fmt.Printf("  + Config applied from '%s'\n", file)
	},
}

var getStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get agent status",
	Run: func(cmd *cobra.Command, args []string) {
		if !IsDaemonRunning() {
			fmt.Println("Agent is not started. Run 'agent start' first")
			return
		}

		isReady, err := GetAgentStatus()
		if err != nil {
			fmt.Println("Error:", err)
		}

		fmt.Printf("Ready:  %v\n", isReady)
		if isReady {
			fmt.Printf("Status: %v\n", "Agent 0018 Active")
		} else {
			fmt.Printf("Status: %v\n", "Agent 0018 Inactive")
		}
	},
}

var runCmd = &cobra.Command{
	Use:    "run",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if !IsDaemonRunning() {
			fmt.Println(" Agent is not started. Run 'agent start' first")
			return
		}

		_, err := GetAgentStatus()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		wsconnection, _, err := websocket.DefaultDialer.Dial(WSURL+PORT+"/agent/chat", nil)
		if err != nil {
			fmt.Println("Agent is not running. Start it with 'agent start'")
			return
		}
		defer wsconnection.Close()

		resource.PrintBanner()
		reader := bufio.NewReader(os.Stdin)

		for {
			fmt.Print("\n\r\033[K") // Important to get you > tag on new line, otherwise you > tag disappears
			resource.YouPrompt()
			prompt, _ := reader.ReadString('\n')
			prompt = strings.TrimSpace(prompt)

			if prompt == "exit" {
				wsconnection.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				resource.AgentPrompt()
				fmt.Println()
				resource.GoodBye()
				break
			}

			if prompt == "" {
				continue
			}

			err := StreamChat(prompt, wsconnection, reader)
			if err != nil {
				fmt.Println(err)
			}
			resource.Separator()
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := StopDaemon()
		if err != nil {
			return err
		}
		fmt.Println("Stopping AGENT 0018...")
		return nil
	},
}
