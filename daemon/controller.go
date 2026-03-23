package daemon

import (
	"context"
	"fmt"
	"net/http"
	"salesforce-ai-agent/configuration"
	"salesforce-ai-agent/core"
	"salesforce-ai-agent/internal/llm"
	client "salesforce-ai-agent/internal/mcpclient"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

func handlePostInit(c *gin.Context) {

	agentDaemon.mu.Lock()
	defer agentDaemon.mu.Unlock()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	agentDaemon.IsInitialized = false
	agentDaemon.LLM = nil
	if agentDaemon.MCPClient != nil {
		closeMCPSessions()
	}

	var config configuration.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		agentDaemon.Logger.Error("invalid config",
			zap.Error(err),
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(400, gin.H{"error": "invalid config"})
		return
	}
	agentDaemon.Config = &config

	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	wg.Add(2)

	go func() {
		defer wg.Done()
		var err error
		agentDaemon.LLM, err = llm.InitializeLLM(ctx, agentDaemon.Config, agentDaemon.Logger)
		if err != nil {
			agentDaemon.Logger.Error("LLMs init failed:",
				zap.Error(err),
			)
			errCh <- err
			return
		}
		agentDaemon.LLM.Planner, agentDaemon.LLM.Executor, err = llm.SetPlannerExecutor(agentDaemon.LLM, agentDaemon.Config, agentDaemon.Logger)
		if err != nil {
			agentDaemon.Logger.Error("Failed to set Planner and Executor",
				zap.Error(err),
			)
			errCh <- err
			return
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		agentDaemon.MCPClient, err = client.InitializeMCPClient(ctx, agentDaemon.Config, agentDaemon.Logger)
		if err != nil {
			agentDaemon.Logger.Error("MCP Servers init failed",
				zap.Error(err),
			)
			return
		}
	}()

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			c.JSON(500, gin.H{"error": "initialization failed"})
			return
		}
	}

	agentDaemon.Logger.Info("All system components initialized!")
	agentDaemon.IsInitialized = true
	c.JSON(200, gin.H{"status": "config applied, initializing..."})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func handleChat(c *gin.Context) {
	agentDaemon.mu.RLock()
	defer agentDaemon.mu.RUnlock()

	if agentDaemon.LLM == nil {
		c.JSON(404, gin.H{"error": "LLM not initialized, run 'agent apply -f' first"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		agentDaemon.Logger.Error("websocket upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	ctx := c.Request.Context()
	history := make([]openai.ChatCompletionMessage, 0)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				agentDaemon.Logger.Error("websocket read error", zap.Error(err))
			}
			return
		}

		prompt := strings.TrimSpace(string(msg))
		if prompt == "" {
			continue
		}

		agentDaemon.Logger.Info("websocket prompt received", zap.String("prompt", prompt))

		tokenChan := make(chan string)
		errChan := make(chan error, 1)
		askChan := make(chan string, 1)
		replyChan := make(chan string, 1)

		go func() {
			core.ExecuteAgent(
				agentDaemon.Logger,
				ctx,
				agentDaemon.LLM,
				prompt,
				&history,
				agentDaemon.MCPClient,
				agentDaemon.Config,
				tokenChan,
				errChan,
				askChan,
				replyChan,
			)
			close(tokenChan)
		}()

	streamLoop:
		for {
			select {
			case <-ctx.Done():
				conn.WriteMessage(websocket.TextMessage, []byte("[DONE]"))
				return

			case err := <-errChan:
				if err != nil {
					conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("[ERROR] %s", err.Error())))
					conn.WriteMessage(websocket.TextMessage, []byte("[DONE]"))
					break streamLoop
				}
				for token := range tokenChan {
					conn.WriteMessage(websocket.TextMessage, []byte(token))
				}
				conn.WriteMessage(websocket.TextMessage, []byte("[DONE]"))
				break streamLoop

			case question, ok := <-askChan:
				if !ok {
					continue
				}
				conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("[ASK] %s", question)))

				_, answer, err := conn.ReadMessage()
				if err != nil {
					agentDaemon.Logger.Error("failed to read answer from client", zap.Error(err))
					replyChan <- ""
					break streamLoop
				}
				replyChan <- strings.TrimSpace(string(answer))

			case token, ok := <-tokenChan:
				if !ok {
					conn.WriteMessage(websocket.TextMessage, []byte("[DONE]"))
					break streamLoop
				}
				conn.WriteMessage(websocket.TextMessage, []byte(token))
			}
		}
	}
}

func handleStop(c *gin.Context) {
	agentDaemon.mu.Lock()
	defer agentDaemon.mu.Unlock()

	agentDaemon.Logger.Info("agent 0018 stopping...")
	c.JSON(200, gin.H{"status": "stopping..."})

	go func() {
		time.Sleep(100 * time.Millisecond)
		closeMCPSessions()
		agentDaemon.Cancel()

	}()
}

func handleAgentStatus(c *gin.Context) {
	agentDaemon.mu.RLock()
	defer agentDaemon.mu.RUnlock()

	if !agentDaemon.IsInitialized {
		c.JSON(200, gin.H{
			"ready":   false,
			"message": "agent not initialized, run 'agent apply -f' first",
		})
		return
	}
	c.JSON(200, gin.H{
		"ready":   true,
		"message": "agent ready",
	})
}

func handleAgentAll(c *gin.Context) {
	agentDaemon.mu.RLock()
	defer agentDaemon.mu.RUnlock()

	if !agentDaemon.IsInitialized {
		c.JSON(200, gin.H{
			"ready":   false,
			"message": "agent not initialized, run 'agent apply -f' first",
		})
		return
	}
}
