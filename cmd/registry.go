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

/*var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View agent logs",
	Run: func(cmd *cobra.Command, args []string) {
		lines, _ := cmd.Flags().GetInt("lines")
		level, _ := cmd.Flags().GetString("level")
		readLogs("agent0018.log", lines, level)
	},
}*/

/*func readLogs(path string, lines int, level string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("  Could not read log file: %v\n", err)
		return
	}

	allLines := strings.Split(string(data), "\n")
	filtered := []string{}
	for _, line := range allLines {
		if line == "" {
			continue
		}
		if level == "" || strings.Contains(strings.ToLower(line), `"level":"`+level+`"`) {
			filtered = append(filtered, line)
		}
	}

	start := 0
	if len(filtered) > lines {
		start = len(filtered) - lines
	}

	fmt.Printf("\n  Showing last %d log entries:\n\n", lines)
	for _, line := range filtered[start:] {
		if strings.Contains(line, `"level":"error"`) {
			fmt.Printf("\033[91m%s\033[0m\n", line)
		} else if strings.Contains(line, `"level":"warn"`) {
			fmt.Printf("\033[93m%s\033[0m\n", line)
		} else {
			fmt.Printf("\033[37m%s\033[0m\n", line)
		}
	}
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources",
}

var getAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Get all models and MCP servers",
	Run: func(cmd *cobra.Command, args []string) {
		config := loadConfig()
		printModelsTable(config, "all")
		printMCPServersTable(config, "all")
	},
}

var getModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List all models",
	Run: func(cmd *cobra.Command, args []string) {
		active, _ := cmd.Flags().GetBool("active")
		inactive, _ := cmd.Flags().GetBool("inactive")

		filter := "all"
		if active {
			filter = "active"
		}
		if inactive {
			filter = "inactive"
		}*/

/*config := loadConfig()
		printModelsTable(config, filter)
	},
}

var getMCPServersCmd = &cobra.Command{
	Use:   "mcpservers",
	Short: "List all MCP servers",
	Run: func(cmd *cobra.Command, args []string) {
		active, _ := cmd.Flags().GetBool("active")
		inactive, _ := cmd.Flags().GetBool("inactive")

		filter := "all"
		if active {
			filter = "active"
		}
		if inactive {
			filter = "inactive"
		}

		config := loadConfig()
		printMCPServersTable(config, filter)
	},
}

var getModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Get model details by reference name",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		config := loadConfig()

		for _, m := range config.Models.Models {
			if m.Name == name {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Reference", "Model Name", "Endpoint", "API Key", "Tier", "Status"})
				table.SetBorder(true)
				status := checkModelConnectivity(m)
				apiKey := maskAPIKey(m.APIKey)
				table.Append([]string{m.Name, m.Model, m.BaseURL, apiKey, strconv.Itoa(m.Tier), status})
				table.Render()
				return
			}
		}
		fmt.Printf("  x Model '%s' not found\n", name)
	},
}

var getMCPServerCmd = &cobra.Command{
	Use:   "mcpserver",
	Short: "Get MCP server details by name",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		config := loadConfig()

		for _, s := range config.MCPServers {
			if s.Name == name {
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"Name", "Transport", "Endpoint", "Status"})
				table.SetBorder(true)
				status := checkMCPConnectivity(s)
				table.Append([]string{s.Name, s.Transport, s.Endpoint, status})
				table.Render()
				return
			}
		}
		fmt.Printf("  x MCP server '%s' not found\n", name)
	},
}*/

/*
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a model or MCP server",
}

var addModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Add a new model",
	Run: func(cmd *cobra.Command, args []string) {
		ref, _ := cmd.Flags().GetString("ref")
		modelname, _ := cmd.Flags().GetString("modelname")
		endpoint, _ := cmd.Flags().GetString("endpoint")
		apikey, _ := cmd.Flags().GetString("apikey")
		tier, _ := cmd.Flags().GetInt("tier")

		config := loadConfig()
		config.Models.Models = append(config.Models.Models, configuration.Model{
			Name:    ref,
			Model:   modelname,
			BaseURL: endpoint,
			APIKey:  apikey,
			Tier:    tier,
		})
		saveConfig(config)
		fmt.Printf("  + Model '%s' added successfully\n", ref)
	},
}

var addMCPServerCmd = &cobra.Command{
	Use:   "mcpserver",
	Short: "Add a new MCP server",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		transport, _ := cmd.Flags().GetString("transport")
		endpoint, _ := cmd.Flags().GetString("endpoint")

		config := loadConfig()
		config.MCPServers = append(config.MCPServers, configuration.MCPServer{
			Name:      name,
			Transport: transport,
			Endpoint:  endpoint,
		})
		saveConfig(config)
		fmt.Printf("  + MCP server '%s' added successfully\n", name)
	},
}

// ── update ────────────────────────────────────────────────────────────────────
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a model or MCP server",
}

var updateModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Update an existing model",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		ref, _ := cmd.Flags().GetString("ref")
		modelname, _ := cmd.Flags().GetString("modelname")
		endpoint, _ := cmd.Flags().GetString("endpoint")
		apikey, _ := cmd.Flags().GetString("apikey")
		tier, _ := cmd.Flags().GetInt("tier")

		config := loadConfig()
		for i, m := range config.Models.Models {
			if m.Name == name {
				if ref != "" {
					config.Models.Models[i].Name = ref
				}
				if modelname != "" {
					config.Models.Models[i].Model = modelname
				}
				if endpoint != "" {
					config.Models.Models[i].BaseURL = endpoint
				}
				if apikey != "" {
					config.Models.Models[i].APIKey = apikey
				}
				if tier != 0 {
					config.Models.Models[i].Tier = tier
				}
				saveConfig(config)
				fmt.Printf("  + Model '%s' updated successfully\n", name)
				return
			}
		}
		fmt.Printf("  x Model '%s' not found\n", name)
	},
}

var updateMCPServerCmd = &cobra.Command{
	Use:   "mcpserver",
	Short: "Update an existing MCP server",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		endpoint, _ := cmd.Flags().GetString("endpoint")
		transport, _ := cmd.Flags().GetString("transport")

		config := loadConfig()
		for i, s := range config.MCPServers {
			if s.Name == name {
				if endpoint != "" {
					config.MCPServers[i].Endpoint = endpoint
				}
				if transport != "" {
					config.MCPServers[i].Transport = transport
				}
				saveConfig(config)
				fmt.Printf("  + MCP server '%s' updated successfully\n", name)
				return
			}
		}
		fmt.Printf("  x MCP server '%s' not found\n", name)
	},
}

// ── delete ────────────────────────────────────────────────────────────────────
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a model or MCP server",
}

var deleteModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Delete a model",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		config := loadConfig()

		for i, m := range config.Models.Models {
			if m.Name == name {
				config.Models.Models = append(
					config.Models.Models[:i],
					config.Models.Models[i+1:]...,
				)
				saveConfig(config)
				fmt.Printf("  + Model '%s' deleted successfully\n", name)
				return
			}
		}
		fmt.Printf("  x Model '%s' not found\n", name)
	},
}

var deleteMCPServerCmd = &cobra.Command{
	Use:   "mcpserver",
	Short: "Delete an MCP server",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		config := loadConfig()

		for i, s := range config.MCPServers {
			if s.Name == name {
				config.MCPServers = append(
					config.MCPServers[:i],
					config.MCPServers[i+1:]...,
				)
				saveConfig(config)
				fmt.Printf("  + MCP server '%s' deleted successfully\n", name)
				return
			}
		}
		fmt.Printf("  x MCP server '%s' not found\n", name)
	},
}

// ── check ─────────────────────────────────────────────────────────────────────
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check connectivity of a model or MCP server",
}

var checkModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Check model connectivity",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		config := loadConfig()

		for _, m := range config.Models.Models {
			if m.Name == name {
				fmt.Printf("  Checking '%s'...\n", name)
				status := checkModelConnectivity(m)
				fmt.Printf("  %s\n", status)
				return
			}
		}
		fmt.Printf("  x Model '%s' not found\n", name)
	},
}

var checkMCPServerCmd = &cobra.Command{
	Use:   "mcpserver",
	Short: "Check MCP server connectivity",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		config := loadConfig()

		for _, s := range config.MCPServers {
			if s.Name == name {
				fmt.Printf("  Checking '%s'...\n", name)
				status := checkMCPConnectivity(s)
				fmt.Printf("  %s\n", status)
				return
			}
		}
		fmt.Printf("  x MCP server '%s' not found\n", name)
	},
}

// ── set ───────────────────────────────────────────────────────────────────────
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set agent configuration",
}

var setPlannerExecutorCmd = &cobra.Command{
	Use:   "roles",
	Short: "Set planner and executor models",
	Run: func(cmd *cobra.Command, args []string) {
		plan, _ := cmd.Flags().GetString("plan")
		exec, _ := cmd.Flags().GetString("exec")
		config := loadConfig()

		planFound, execFound := false, false
		for i, m := range config.Models.Models {
			if m.Name == plan {
				config.Models.Models[i].Tier = 1
				planFound = true
			}
			if m.Name == exec {
				config.Models.Models[i].Tier = 2
				execFound = true
			}
		}

		if !planFound {
			fmt.Printf("  x Planner model '%s' not found\n", plan)
			return
		}
		if !execFound {
			fmt.Printf("  x Executor model '%s' not found\n", exec)
			return
		}

		saveConfig(config)
		fmt.Printf("  + Planner set to '%s'\n", plan)
		fmt.Printf("  + Executor set to '%s'\n", exec)
	},
}


// ── helpers ───────────────────────────────────────────────────────────────────
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

func printModelsTable(config *configuration.Config, filter string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Reference", "Model Name", "Endpoint", "Tier", "Status"})
	table.SetBorder(true)

	for _, m := range config.Models.Models {
		status := checkModelConnectivity(m)
		if filter == "active" && status != "active" {
			continue
		}
		if filter == "inactive" && status != "inactive" {
			continue
		}

		tierLabel := strconv.Itoa(m.Tier)
		if m.Tier == 0 {
			tierLabel = "-"
		}

		row := []string{m.Name, m.Model, m.BaseURL, tierLabel, status}
		if status == "active" {
			table.Rich(row, []tablewriter.Colors{
				{}, {}, {}, {},
				{tablewriter.FgGreenColor},
			})
		} else {
			table.Rich(row, []tablewriter.Colors{
				{}, {}, {}, {},
				{tablewriter.FgRedColor},
			})
		}
	}
	table.Render()
}

func printMCPServersTable(config *configuration.Config, filter string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Transport", "Endpoint", "Status"})
	table.SetBorder(true)

	for _, s := range config.MCPServers {
		status := checkMCPConnectivity(s)
		if filter == "active" && status != "active" {
			continue
		}
		if filter == "inactive" && status != "inactive" {
			continue
		}

		row := []string{s.Name, s.Transport, s.Endpoint, status}
		if status == "active" {
			table.Rich(row, []tablewriter.Colors{
				{}, {}, {},
				{tablewriter.FgGreenColor},
			})
		} else {
			table.Rich(row, []tablewriter.Colors{
				{}, {}, {},
				{tablewriter.FgRedColor},
			})
		}
	}
	table.Render()
}
*/
