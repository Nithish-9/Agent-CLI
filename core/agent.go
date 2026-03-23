package core

import (
	"context"
	"salesforce-ai-agent/configuration"
	llm "salesforce-ai-agent/internal/llm"
	client "salesforce-ai-agent/internal/mcpclient"
	resource "salesforce-ai-agent/resources"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

func ExecuteAgent(appLogger *zap.Logger, ctx context.Context, llmModelsResult *llm.LLMModels, prompt string, history *[]openai.ChatCompletionMessage, mcpClientResult *client.MCPClient, configYaml *configuration.Config, tokenChan chan<- string, errChan chan<- error, askChan chan<- string, replyChan <-chan string) {

	fallback := func() {
		messages := buildMessages(history, prompt)
		response, err := llm.StreamFinalResponse(appLogger, ctx, llmModelsResult.Executor, messages, tokenChan, errChan)
		if err == nil && response != "" {
			*history = append(*history, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: response,
			})
		}
	}

	if mcpClientResult == nil || mcpClientResult.MCPServerPlannerInput == nil {
		fallback()
		return
	}

	mcpServerPlannerOutput, err := llm.RunServerPlanner(appLogger, ctx, llmModelsResult.Planner, prompt, mcpClientResult.MCPServerPlannerInput)
	if err != nil {
		appLogger.Warn("Planner failed, retrying once", zap.Error(err))
		mcpServerPlannerOutput, err = llm.RunServerPlanner(appLogger, ctx, llmModelsResult.Planner, prompt, mcpClientResult.MCPServerPlannerInput)
		if err != nil {
			appLogger.Warn("Planner failed again, falling back to direct response", zap.Error(err))
			fallback()
			return
		}
	}

	if mcpServerPlannerOutput == nil || len(mcpServerPlannerOutput.MCPServer) == 0 {
		fallback()
		return
	}

	mcpToolPlannerInput, err := client.ProcessMCPServerPlannerOutput(appLogger, mcpClientResult, mcpServerPlannerOutput)
	if err != nil {
		appLogger.Error("Agent failed to process the MCP Server Planner Output", zap.Error(err))
		fallback()
		return
	}

	mcpToolPlannerOutput, err := llm.RunToolPlanner(appLogger, ctx, llmModelsResult.Planner, prompt, mcpToolPlannerInput)
	if err != nil {
		appLogger.Error("Agent failed to run the MCP Tool Planner", zap.Error(err))
		fallback()
		return
	}

	if mcpToolPlannerOutput == nil {
		appLogger.Error("Tool Planner Output is nil, falling back to direct response")
		fallback()
		return
	}

	response, err := llm.RunToolExecutor(appLogger, ctx, llmModelsResult.Executor, prompt, mcpToolPlannerOutput, mcpClientResult, tokenChan, errChan, askChan, replyChan)
	if err == nil && response != "" {
		*history = append(*history, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: response,
		})
	}
}

func buildMessages(history *[]openai.ChatCompletionMessage, prompt string) []openai.ChatCompletionMessage {
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: resource.ChatMessageRoleSystem},
	}

	start := 0
	if len(*history) > 6 {
		start = len(*history) - 6
	}
	messages = append(messages, (*history)[start:]...)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})
	return messages
}
