package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	mcpclient "salesforce-ai-agent/internal/mcpclient"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

func RunServerPlanner(appLogger *zap.Logger, ctx context.Context, model *LLMModel, userPrompt string, mcpServerPlannerInput *mcpclient.MCPServerPlannerInput) (*mcpclient.MCPServerPlannerOutput, error) {

	if mcpServerPlannerInput == nil {
		appLogger.Error("Server Planner input is nil")
		return nil, fmt.Errorf("Server Planner input is nil")
	}

	inputJSON, err := json.MarshalIndent(mcpServerPlannerInput, "", "  ")
	appLogger.Info("MCP Server Planner Input JSON", zap.String("schema", string(inputJSON)))
	if err != nil {
		appLogger.Error("failed to marshal input: ",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	var mcpServerPlannerOutput mcpclient.MCPServerPlannerOutput
	schemaBytes, err := generateSchema(&mcpclient.MCPServerPlannerOutput{})
	if err != nil {
		appLogger.Error("schema error:",
			zap.Error(err),
		)
		return nil, fmt.Errorf("schema error: %w", err)
	}

	appLogger.Info("generated schema", zap.String("schema", string(schemaBytes)))

	tool := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "select_mcp_servers",
			Description: "Select only the relevant MCP servers and tool categories based on the user prompt. Return empty mcp_servers array if the prompt is general and needs no tools.",
			Parameters:  json.RawMessage(schemaBytes),
			Strict:      true,
		},
	}

	resp, err := model.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf(`You are a tool routing assistant for a Salesforce AI Agent called APEX.
Your ONLY job is to analyse the user prompt and select the relevant MCP servers and tool categories needed to fulfil the request.

## Rules
- Return ONLY servers and tool categories that are directly relevant to the user prompt
- Use the EXACT server names and category names from the catalogue — do not rename or modify them
- If the prompt is a general question or conversation that needs no tools, return empty mcp_servers array
- Do NOT select tool categories that are not needed
- A single server can have multiple relevant categories
- Multiple servers can be selected if needed

## Available MCP Servers Catalogue
%s`, string(inputJSON)),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Tools: []openai.Tool{tool},
		ToolChoice: openai.ToolChoice{
			Type:     openai.ToolTypeFunction,
			Function: openai.ToolFunction{Name: "select_mcp_servers"},
		},
	})
	if err != nil {
		appLogger.Error("Planner model error ",
			zap.Error(err),
		)
		return nil, fmt.Errorf("Planner model error: %w", err)
	}

	if len(resp.Choices) == 0 {
		appLogger.Error("Planner model returned no choices",
			zap.String("model", model.Model),
			zap.String("prompt", userPrompt),
		)
		return nil, fmt.Errorf("Planner model returned no choices")
	}

	if len(resp.Choices[0].Message.ToolCalls) == 0 {
		appLogger.Error("Planner model no tool calls",
			zap.String("model", model.Model),
			zap.String("prompt", userPrompt),
		)
		return nil, fmt.Errorf("Planner model returned no tool calls")
	}

	args := resp.Choices[0].Message.ToolCalls[0].Function.Arguments
	appLogger.Info("model raw output", zap.String("args", args))

	if err := json.Unmarshal([]byte(args), &mcpServerPlannerOutput); err != nil {
		appLogger.Error("failed to unmarshal mcpServerPlannerOutput", zap.Error(err))
		return nil, fmt.Errorf("unmarshal mcpServerPlannerOutput: %w", err)
	}

	appLogger.Info("server planner completed",
		zap.String("Planner", model.Model),
		zap.Any("servers_selected", mcpServerPlannerOutput),
	)

	return &mcpServerPlannerOutput, nil
}

func RunToolPlanner(appLogger *zap.Logger, ctx context.Context, model *LLMModel, userPrompt string, mcpToolPlannerInput *mcpclient.MCPToolPlannerInput) (*mcpclient.ToolPlannerOutput, error) {

	if mcpToolPlannerInput == nil {
		appLogger.Error("Tool Planner input is nil")
		return nil, fmt.Errorf("Tool Planner input is nil")
	}

	inputJSON, err := json.MarshalIndent(mcpToolPlannerInput, "", "  ")
	if err != nil {
		appLogger.Error("failed to marshal tool planner input", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal tool planner input: %w", err)
	}

	var toolPlannerOutput mcpclient.ToolPlannerOutput
	schemaBytes, err := generateSchema(&mcpclient.ToolPlannerOutput{})
	if err != nil {
		appLogger.Error("tool planner schema marshal error", zap.Error(err))
		return nil, fmt.Errorf("tool planner schema marshal error: %w", err)
	}

	tool := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "select_mcp_tools",
			Description: "Select only the relevant tools from the given MCP servers needed to fulfil the user request. Return empty mcp_tools array if no tools are needed.",
			Parameters:  json.RawMessage(schemaBytes),
		},
	}

	resp, err := model.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: model.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf(`You are a tool selection assistant for a Salesforce AI Agent called APEX.
Your ONLY job is to analyse the user prompt and select the exact tools needed to fulfil the request.

## Rules
- Return ONLY the tools that are directly needed to fulfil the user prompt
- Use the EXACT tool names and server names from the catalogue — do not rename or modify them
- If no tools are needed return empty mcp_tools array
- Select the minimum number of tools required

## Available Tools Catalogue
%s`, string(inputJSON)),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		Tools: []openai.Tool{tool},
		ToolChoice: openai.ToolChoice{
			Type:     openai.ToolTypeFunction,
			Function: openai.ToolFunction{Name: "select_mcp_tools"},
		},
	})
	if err != nil {
		appLogger.Error("tool planner model error", zap.Error(err))
		return nil, fmt.Errorf("tool planner model error: %w", err)
	}

	if len(resp.Choices) == 0 {
		appLogger.Error("tool planner returned no choices", zap.Error(err))
		return nil, fmt.Errorf("tool planner returned no choices")
	}

	if len(resp.Choices[0].Message.ToolCalls) == 0 {
		appLogger.Error("tool planner returned no tool calls", zap.Error(err))
		return nil, fmt.Errorf("tool planner returned no tool calls")
	}

	args := resp.Choices[0].Message.ToolCalls[0].Function.Arguments

	if err := json.Unmarshal([]byte(args), &toolPlannerOutput); err != nil {
		appLogger.Error("failed to unmarshall tool planner output", zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshall tool planner output: %w", err)
	}

	llmOutput, err := json.MarshalIndent(toolPlannerOutput, "", "  ")
	if err != nil {
		appLogger.Error("failed to marshal tool planner output", zap.Error(err))
	} else {
		appLogger.Info("tool planner output", zap.String("output", string(llmOutput)))
	}

	return &toolPlannerOutput, nil
}

func ConvertToOpenAITools(appLogger *zap.Logger, toolPlannerOutput *mcpclient.ToolPlannerOutput, mcpClient *mcpclient.MCPClient) []openai.Tool {
	var openAITools []openai.Tool

	for _, mcpServer := range toolPlannerOutput.MCPToolPlanner {
		serverInfo, ok := mcpClient.ServerInfo[mcpServer.Name]
		if !ok {
			fmt.Printf("Server '%s' not found in MCP client\n", mcpServer.Name)
			continue
		}

		for _, tool := range mcpServer.ToolLst {
			params := buildSafeParams(nil)
			found := false
			for _, toolSpecData := range serverInfo.Tools {
				if mcpTool, exists := toolSpecData.ToolMap[tool.Name]; exists {
					params = buildSafeParams(mcpTool.InputSchema)
					found = true
					break
				}
			}

			if !found {
				appLogger.Warn("tool not found in ToolMap, using empty schema",
					zap.String("tool", tool.Name),
				)
			}

			openAITools = append(openAITools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Name,
					Description: serverInfo.ToolConfigMap[tool.Name].Description,
					Parameters:  params,
				},
			})
		}
	}

	return openAITools
}
func buildSafeParams(inputSchema any) json.RawMessage {
	if inputSchema == nil {
		return json.RawMessage(`{"type":"object","properties":{}}`)
	}

	schemaBytes, err := json.Marshal(inputSchema)
	if err != nil {
		return json.RawMessage(`{"type":"object","properties":{}}`)
	}

	var raw map[string]any
	if err := json.Unmarshal(schemaBytes, &raw); err != nil {
		return json.RawMessage(`{"type":"object","properties":{}}`)
	}

	delete(raw, "$schema")
	delete(raw, "$ref")
	delete(raw, "$defs")
	delete(raw, "additionalProperties")

	if _, ok := raw["type"]; !ok {
		raw["type"] = "object"
	}
	if _, ok := raw["properties"]; !ok {
		raw["properties"] = map[string]any{}
	}

	cleanBytes, err := json.Marshal(raw)
	if err != nil {
		return json.RawMessage(`{"type":"object","properties":{}}`)
	}

	return json.RawMessage(cleanBytes)
}

func RunToolExecutor(appLogger *zap.Logger, ctx context.Context, model *LLMModel, userPrompt string,
	toolPlannerOutput *mcpclient.ToolPlannerOutput, mcpClient *mcpclient.MCPClient,
	tokenChan chan<- string, errChan chan<- error, askChan chan<- string, replyChan <-chan string) (string, error) {

	openAITools := ConvertToOpenAITools(appLogger, toolPlannerOutput, mcpClient)

	openAITools = append(openAITools, buildAskUserTool())

	if len(openAITools) == 1 {
		return StreamFinalResponse(appLogger, ctx, model, []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "You are AGENT 0018, a Salesforce AI Agent."},
			{Role: openai.ChatMessageRoleUser, Content: userPrompt},
		}, tokenChan, errChan)
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are AGENT 0018, a Salesforce AI Agent. Use the provided tools to fulfil the user request. Call tools in the correct order. If you need clarification or a choice from the user before proceeding, use the ask_user tool.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}

	const maxIterations = 7
	iteration := 0

	for {
		if iteration >= maxIterations {
			appLogger.Info("Executor Model Max iterations reached")
			return StreamFinalResponse(appLogger, ctx, model, messages, tokenChan, errChan)
		}
		iteration++

		resp, err := model.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    model.Model,
			Messages: messages,
			Tools:    openAITools,
		})
		if err != nil {
			if strings.Contains(err.Error(), "400") {
				appLogger.Warn("executor LLM 400 error, retrying with final response",
					zap.Error(err),
				)
				return StreamFinalResponse(appLogger, ctx, model, messages, tokenChan, errChan)
			}
			appLogger.Error("executor LLM error",
				zap.Error(err),
			)
			return "", fmt.Errorf("executor LLM error: %w", err)
		}

		if len(resp.Choices) == 0 {
			appLogger.Error("LLM returned no choices",
				zap.String("model", model.Model),
			)
			return "", fmt.Errorf("no choices returned")
		}

		choice := resp.Choices[0]
		messages = append(messages, choice.Message)

		switch choice.FinishReason {

		case openai.FinishReasonToolCalls:
			for _, toolCall := range choice.Message.ToolCalls {
				if toolCall.Function.Name == "ask_user" {
					var askArgs struct {
						Question string   `json:"question"`
						Options  []string `json:"options"`
					}
					json.Unmarshal([]byte(toolCall.Function.Arguments), &askArgs)
					question := askArgs.Question
					for i, opt := range askArgs.Options {
						question += fmt.Sprintf("\n  %d. %s", i+1, opt)
					}

					askChan <- question

					var userAnswer string
					select {
					case userAnswer = <-replyChan:
					case <-time.After(5 * time.Minute):
						userAnswer = "user did not respond"
					case <-ctx.Done():
						return "", nil
					}

					messages = append(messages, openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleTool,
						ToolCallID: toolCall.ID,
						Content:    userAnswer,
					})
					continue
				}

				appLogger.Info("calling tool",
					zap.String("tool", toolCall.Function.Name),
					zap.String("args", toolCall.Function.Arguments),
				)

				var args map[string]any
				argStr := toolCall.Function.Arguments
				if argStr == "" || argStr == "null" || argStr == "{}" {
					args = map[string]any{}
				} else {
					if err := json.Unmarshal([]byte(argStr), &args); err != nil {
						appLogger.Error("failed to parse tool args",
							zap.String("tool", toolCall.Function.Name),
							zap.String("args", argStr),
							zap.Error(err),
						)
						messages = append(messages, openai.ChatCompletionMessage{
							Role:       openai.ChatMessageRoleTool,
							ToolCallID: toolCall.ID,
							Content:    fmt.Sprintf("error parsing args: %v", err),
						})
						continue
					}
				}

				_, session, err := findToolSession(appLogger, toolCall.Function.Name, toolPlannerOutput, mcpClient)
				if err != nil {
					appLogger.Error("tool not found",
						zap.String("tool", toolCall.Function.Name),
						zap.Error(err),
					)
					messages = append(messages, openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleTool,
						ToolCallID: toolCall.ID,
						Content:    fmt.Sprintf("tool not found: %v", err),
					})
					continue
				}

				toolCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				mcpResult, err := session.CallTool(toolCtx, &mcp.CallToolParams{
					Name:      toolCall.Function.Name,
					Arguments: args,
				})
				cancel()

				if err != nil {
					appLogger.Error("tool execution failed",
						zap.String("tool", toolCall.Function.Name),
						zap.Error(err),
					)
					messages = append(messages, openai.ChatCompletionMessage{
						Role:       openai.ChatMessageRoleTool,
						ToolCallID: toolCall.ID,
						Content:    fmt.Sprintf("tool failed: %v", err),
					})
					continue
				}

				resultText := truncateResult(extractMCPResult(mcpResult), 4000)
				appLogger.Info("tool result",
					zap.String("tool", toolCall.Function.Name),
					zap.String("result", resultText),
				)

				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					ToolCallID: toolCall.ID,
					Content:    resultText,
				})
			}

		case openai.FinishReasonStop:
			return StreamFinalResponse(appLogger, ctx, model, messages, tokenChan, errChan)

		default:
			return StreamFinalResponse(appLogger, ctx, model, messages, tokenChan, errChan)
		}
	}
}

func buildAskUserTool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "ask_user",
			Description: "Ask the user for clarification, confirmation, or a choice before proceeding. Use this when you need more information from the user.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"question": {
						"type": "string",
						"description": "The question to ask the user"
					},
					"options": {
						"type": "array",
						"items": {"type": "string"},
						"description": "Optional list of choices to show the user"
					}
				},
				"required": ["question"]
			}`),
		},
	}
}

func truncateResult(result string, maxLen int) string {
	if len(result) <= maxLen {
		return result
	}
	return result[:maxLen] + "... [truncated]"
}

func findToolSession(appLogger *zap.Logger, toolName string, toolPlannerOutput *mcpclient.ToolPlannerOutput, mcpClient *mcpclient.MCPClient) (string, *mcp.ClientSession, error) {
	seen := make(map[string]bool)
	for _, mcpServer := range toolPlannerOutput.MCPToolPlanner {
		if seen[mcpServer.Name] {
			continue
		}
		seen[mcpServer.Name] = true
		serverInfo, ok := mcpClient.ServerInfo[mcpServer.Name]
		if !ok || serverInfo.Session == nil {
			appLogger.Warn("server not found or inactive — skipping",
				zap.String("server", mcpServer.Name),
			)
			continue
		}

		for _, toolSpecData := range serverInfo.Tools {
			if _, exists := toolSpecData.ToolMap[toolName]; exists {
				return mcpServer.Name, serverInfo.Session, nil
			}
		}
	}

	return "", nil, fmt.Errorf("tool '%s' not found in any server", toolName)
}

func extractMCPResult(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return "no result"
	}
	var parts []string
	for _, content := range result.Content {
		if textContent, ok := content.(*mcp.TextContent); ok {
			parts = append(parts, textContent.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func StreamFinalResponse(appLogger *zap.Logger, ctx context.Context, model *LLMModel,
	messages []openai.ChatCompletionMessage, tokenChan chan<- string, errChan chan<- error) (string, error) {
	stream, err := model.Client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    model.Model,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		appLogger.Error("stream error: ",
			zap.Error(err),
		)
		errChan <- fmt.Errorf("stream error: %w", err)
		return "", fmt.Errorf("stream error: %w", err)
	}
	defer stream.Close()

	var buffer strings.Builder
	inThinkBlock := false

	var fullResponse strings.Builder

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			appLogger.Error("stream recv error: ",
				zap.Error(err),
			)
			errChan <- fmt.Errorf("stream recv error: %w", err)
			return "", fmt.Errorf("stream recv error: %w", err)
		}

		token := response.Choices[0].Delta.Content
		buffer.WriteString(token)
		current := buffer.String()

		if !inThinkBlock && strings.Contains(current, "<think>") {
			inThinkBlock = true
			before := current[:strings.Index(current, "<think>")]
			if before != "" {
				tokenChan <- before
			}
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

		tokenChan <- token
		fullResponse.WriteString(token)
		buffer.Reset()
	}

	return fullResponse.String(), nil
}
