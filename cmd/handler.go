package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"salesforce-ai-agent/configuration"
	llm "salesforce-ai-agent/internal/llm"
	client "salesforce-ai-agent/internal/mcpclient"
	resource "salesforce-ai-agent/resources"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/gorilla/websocket"
)

const PORT = ":9999"
const URL = "http://localhost"
const WSURL = "ws://localhost"

func ListAllConnectedComponents(mcpClientResult *client.MCPClient, llmModelsResult *llm.LLMModels) {

	width := 60
	border := strings.Repeat("─", width)

	fmt.Printf("\n%s%s%s\n", resource.BlueBright, border, resource.Reset)
	fmt.Printf("%s   System Status%s\n", resource.BlueBright+resource.Bold, resource.Reset)
	fmt.Printf("%s%s%s\n", resource.BlueBright, border, resource.Reset)

	//listActiveModels(llmModelsResult)
	listActiveMCPServers(mcpClientResult)

	fmt.Printf("%s%s%s\n\n", resource.BlueBright, border, resource.Reset)
}

/*func listActiveModels(llmModelsResult *llm.LLMModels) {
	fmt.Printf("\n%s   Models%s\n", resource.WhiteBright+resource.Bold, resource.Reset)

	if llmModelsResult == nil || len(llmModelsResult.Models) == 0 {
		fmt.Printf("%s     No models connected%s\n", resource.DimWhite, resource.Reset)
	} else {
		for name, model := range llmModelsResult.Models {
			tierLabel := "[ " + fmt.Sprintf("%d", model.Tier) + " ]"
			fmt.Printf("%s     %-20s%s  %s%s%s%s\n",
				resource.CyanBright, name, resource.Reset,
				resource.DimWhite, model.Model,
				tierLabel,
				resource.Reset,
			)
		}
	}
}*/

func listActiveMCPServers(mcpClientResult *client.MCPClient) {
	fmt.Printf("\n%s  * MCP Servers%s\n", resource.WhiteBright+resource.Bold, resource.Reset)

	if mcpClientResult == nil || len(mcpClientResult.ServerInfo) == 0 {
		fmt.Printf("%s    x No servers connected%s\n", resource.DimWhite, resource.Reset)
	} else {
		for serverName, serverInfo := range mcpClientResult.ServerInfo {
			fmt.Printf("%s    + %s%s\n", resource.CyanBright, serverName, resource.Reset)

			for categoryName, toolSpecData := range serverInfo.Tools {
				fmt.Printf("%s      |-- %s%s\n", resource.BlueBright, categoryName, resource.Reset)
				for i, tool := range toolSpecData.ToolList {
					prefix := "|   |--"
					if i == len(toolSpecData.ToolList)-1 {
						prefix = "|   \\--"
					}
					fmt.Printf("%s      %s %s%s\n", resource.DimWhite, prefix, tool.Name, resource.Reset)
				}
			}
			fmt.Println()
		}
	}
}

func IsDaemonRunning() bool {
	conn, err := net.Listen("tcp", PORT)
	if err != nil {
		return true
	}
	conn.Close()
	return false
}

func StreamChat(prompt string, wsconnection *websocket.Conn, reader *bufio.Reader) error {

	if err := wsconnection.WriteMessage(websocket.TextMessage, []byte(prompt)); err != nil {
		return fmt.Errorf("failed to send prompt: %w", err)
	}

	var fullResponse strings.Builder
	inThinkBlock := false
	var buffer strings.Builder

	var spinnerFrames []string
	if runtime.GOOS == "windows" {
		spinnerFrames = []string{"-", "\\", "|", "/"}
	} else {
		spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	}

	var wg sync.WaitGroup
	spinnerDone := make(chan struct{})

	startSpinner := func(done chan struct{}) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			i := 0
			for {
				select {
				case <-done:
					fmt.Print("\r\033[K")
					return
				default:
					fmt.Printf("\r%s   %s thinking...%s", resource.CyanBright, spinnerFrames[i%len(spinnerFrames)], resource.Reset)
					i++
					time.Sleep(80 * time.Millisecond)
				}
			}
		}()
	}

	closeSpinner := func() {
		close(spinnerDone)
		wg.Wait()
	}

	fmt.Println()
	startSpinner(spinnerDone)

	for {
		_, msg, err := wsconnection.ReadMessage()
		if err != nil {
			closeSpinner()
			break
		}

		text := string(msg)

		if text == "[DONE]" {
			closeSpinner()
			break
		}

		if strings.HasPrefix(text, "[ERROR]") {
			closeSpinner()
			fmt.Println("\nError:", strings.TrimPrefix(text, "[ERROR] "))
			break
		}

		if strings.HasPrefix(text, "[ASK]") {
			closeSpinner()
			question := strings.TrimPrefix(text, "[ASK] ")
			fmt.Print(resource.Reset)
			resource.AgentPrompt()
			fmt.Println(question)
			fmt.Print("\n\r\033[K")
			resource.YouPrompt()

			answer, _ := reader.ReadString('\n')
			answer = strings.TrimSpace(answer)

			if err := wsconnection.WriteMessage(websocket.TextMessage, []byte(answer)); err != nil {
				return fmt.Errorf("failed to send answer: %w", err)
			}

			wg = sync.WaitGroup{}
			spinnerDone = make(chan struct{})
			closeSpinner = func() {
				close(spinnerDone)
				wg.Wait()
			}
			fmt.Println()
			startSpinner(spinnerDone)
			continue
		}

		token := strings.ReplaceAll(text, "\\n", "\n")
		buffer.WriteString(token)
		current := buffer.String()

		if !inThinkBlock && strings.Contains(current, "<think>") {
			inThinkBlock = true
			buffer.Reset()
			continue
		}

		if inThinkBlock {
			if strings.Contains(current, "</think>") {
				inThinkBlock = false
				after := current[strings.Index(current, "</think>")+len("</think>"):]
				buffer.Reset()
				buffer.WriteString(after)
			} else {
				buffer.Reset()
			}
			continue
		}

		fullResponse.WriteString(token)
		buffer.Reset()
	}

	fmt.Print(resource.Reset)
	fmt.Print("\r\033[K")

	resource.AgentPrompt()
	fmt.Println()

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStylePath(resource.GLAMOUR_STYLE),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		fmt.Println(fullResponse.String())
		return nil
	}

	out, err := renderer.Render(fullResponse.String())
	if err != nil {
		fmt.Println(fullResponse.String())
		return nil
	}

	fmt.Print(out)
	return nil
}

func CheckReady() error {
	resp, err := http.Get(URL + PORT + "/agent/status")
	if err != nil {
		return fmt.Errorf("daemon not running")
	}
	defer resp.Body.Close()

	var status struct {
		Ready   bool   `json:"ready"`
		Message string `json:"message"`
	}
	json.NewDecoder(resp.Body).Decode(&status)

	if !status.Ready {
		return fmt.Errorf("%s", status.Message)
	}
	return nil
}

func StopDaemon() error {
	resp, err := http.Post(URL+PORT+"/agent/stop", "", nil)
	if err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func SendConfig(config *configuration.Config) error {
	body, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	resp, err := http.Post(
		URL+PORT+"/agent/init",
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return fmt.Errorf("failed to send config to daemon: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func GetAgentStatus() (bool, error) {
	resp, err := http.Get(URL + PORT + "/agent/status")
	if err != nil {
		return false, fmt.Errorf("agent 0018 not running: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	isReady, ok := result["ready"].(bool)
	if !ok {
		return false, fmt.Errorf("invalid response format")
	}
	if !isReady {
		return false, fmt.Errorf("agent not initialized, run 'agent apply -f' first")
	}

	return true, nil
}
