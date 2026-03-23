package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

func (c *MCPClient) StreamableClientTransport(ctx context.Context, serverName string, endpoint string, mu *sync.Mutex, connErrors *[]string, appLogger *zap.Logger) {
	transport := &mcp.StreamableClientTransport{
		Endpoint: endpoint,
	}
	session, err := c.Client.Connect(ctx, transport, nil)
	if err != nil {
		mu.Lock()
		c.ServerInfo[serverName] = &MCPServerInfo{
			URL:       endpoint,
			Transport: "streamable-http",
		}
		*connErrors = append(*connErrors, fmt.Sprintf("%s: %v", serverName, err))
		mu.Unlock()
		appLogger.Error("error connecting to MCP server",
			zap.String("server", serverName),
			zap.String("endpoint", endpoint),
			zap.Error(err),
		)
	} else {
		mu.Lock()
		c.ServerInfo[serverName] = &MCPServerInfo{
			URL:           endpoint,
			Transport:     "streamable-http",
			Session:       session,
			Tools:         make(map[string]*ToolSpecData),
			ToolConfigMap: make(map[string]*mcp.Tool),
		}
		listAllTools(c, ctx, serverName, appLogger, mu)
		mu.Unlock()
	}
}

func (c *MCPClient) CommandTransport(ctx context.Context, serverName string, endpoint string, mu *sync.Mutex, connErrors *[]string, appLogger *zap.Logger) {
	transport := &mcp.CommandTransport{
		Command: exec.Command(serverName),
	}
	session, err := c.Client.Connect(ctx, transport, nil)
	if err != nil {
		mu.Lock()
		c.ServerInfo[serverName] = &MCPServerInfo{
			URL:       endpoint,
			Transport: "stdio",
		}
		*connErrors = append(*connErrors, fmt.Sprintf("%s: %v", serverName, err))
		mu.Unlock()
		appLogger.Error("error connecting to MCP server",
			zap.String("server", serverName),
			zap.String("endpoint", endpoint),
			zap.Error(err),
		)
	} else {
		mu.Lock()
		c.ServerInfo[serverName] = &MCPServerInfo{
			URL:           endpoint,
			Transport:     "stdio",
			Session:       session,
			Tools:         make(map[string]*ToolSpecData),
			ToolConfigMap: make(map[string]*mcp.Tool),
		}
		listAllTools(c, ctx, serverName, appLogger, mu)
		mu.Unlock()
	}
}

func (c *MCPClient) SSEClientTransport(ctx context.Context, serverName string, endpoint string, mu *sync.Mutex, connErrors *[]string, appLogger *zap.Logger) {
	transport := &mcp.SSEClientTransport{
		Endpoint: endpoint,
	}
	session, err := c.Client.Connect(ctx, transport, nil)
	if err != nil {
		mu.Lock()
		c.ServerInfo[serverName] = &MCPServerInfo{
			URL:       endpoint,
			Transport: "sse",
		}
		*connErrors = append(*connErrors, fmt.Sprintf("%s: %v", serverName, err))
		mu.Unlock()
		appLogger.Error("error connecting to MCP server",
			zap.String("server", serverName),
			zap.String("endpoint", endpoint),
			zap.Error(err),
		)
	} else {
		mu.Lock()
		c.ServerInfo[serverName] = &MCPServerInfo{
			URL:           endpoint,
			Transport:     "sse",
			Session:       session,
			Tools:         make(map[string]*ToolSpecData),
			ToolConfigMap: make(map[string]*mcp.Tool),
		}
		listAllTools(c, ctx, serverName, appLogger, mu)
		mu.Unlock()
	}
}

func listAllTools(c *MCPClient, ctx context.Context, serverName string, appLogger *zap.Logger, mu *sync.Mutex) {

	params := &mcp.CallToolParams{
		Name:      "list_all_tools",
		Arguments: map[string]any{},
	}

	mu.Unlock()
	res, err := c.ServerInfo[serverName].Session.CallTool(ctx, params)
	mu.Lock()
	if err != nil {
		appLogger.Error("error calling list_all_tools",
			zap.String("server", serverName),
			zap.Error(err),
		)
		return
	}

	text := res.Content[0].(*mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &c.ServerInfo[serverName].Metadata); err != nil {
		appLogger.Warn("failed to unmarshal metadata",
			zap.String("server", serverName),
			zap.Error(err),
		)
		return
	}

	toolCatLst := make([]*PlainToolCatData, 0)

	for _, category := range c.ServerInfo[serverName].Metadata.ToolCategories {
		toolSpecData := &ToolSpecData{
			ToolMap:  make(map[string]*mcp.Tool),
			ToolList: make([]*PlainToolData, 0),
		}

		for _, tool := range category.Tools {
			c.ServerInfo[serverName].ToolConfigMap[tool.Tool.Name] = tool.Tool
			toolSpecData.ToolList = append(toolSpecData.ToolList, &PlainToolData{
				Name:        tool.Tool.Name,
				Description: tool.Tool.Description,
			})
			toolSpecData.ToolMap[tool.Tool.Name] = tool.Tool
		}

		c.ServerInfo[serverName].Tools[category.Name] = toolSpecData
		toolCatLst = append(toolCatLst, &PlainToolCatData{
			Name:        category.Name,
			Description: category.Description,
		})
	}

	appLogger.Info("server loaded",
		zap.String("server", serverName),
		zap.Int("categories", len(c.ServerInfo[serverName].Metadata.ToolCategories)),
	)
	for category, tools := range c.ServerInfo[serverName].Tools {
		fmt.Printf("    %s:\n", category)
		for _, toolSpec := range tools.ToolList {
			fmt.Printf("      - %s\n", toolSpec.Name)
		}
	}

	c.MCPServerPlannerInput.MCPServer = append(c.MCPServerPlannerInput.MCPServer, &MCPServerPlanner{
		Name:            serverName,
		Description:     c.ServerInfo[serverName].Metadata.Description,
		ToolCategoryLst: toolCatLst,
	})
}

func ProcessMCPServerPlannerOutput(appLogger *zap.Logger, mcpClientResult *MCPClient, mcpServerPlannerOutput *MCPServerPlannerOutput) (*MCPToolPlannerInput, error) {
	if mcpServerPlannerOutput == nil {
		appLogger.Error("Server Planner output is nil")
		return nil, fmt.Errorf("Server Planner output is nil")
	}

	mcpToolPlannerInput := &MCPToolPlannerInput{
		MCPToolPlanner: make([]*MCPToolPlanner, 0),
	}

	for _, mcpServer := range mcpServerPlannerOutput.MCPServer {
		serverInfo, ok := mcpClientResult.ServerInfo[mcpServer.Name]
		if !ok || serverInfo.Session == nil {
			appLogger.Warn("server inactive — skipping", zap.String("server", mcpServer.Name))
			continue
		}

		toolPlanner := MCPToolPlanner{
			Name:        serverInfo.Metadata.Name,
			Description: serverInfo.Metadata.Description,
			ToolLst:     make([]*PlainToolData, 0),
		}

		for _, toolCategory := range mcpServer.ToolCategoryLst {
			toolSpecData, ok := serverInfo.Tools[toolCategory.Name]
			if !ok {
				appLogger.Warn("category not found in server — skipping",
					zap.String("category", toolCategory.Name),
					zap.String("server", mcpServer.Name),
				)
				continue
			}
			toolPlanner.ToolLst = append(toolPlanner.ToolLst, toolSpecData.ToolList...)
		}

		mcpToolPlannerInput.MCPToolPlanner = append(mcpToolPlannerInput.MCPToolPlanner, &toolPlanner)
	}

	return mcpToolPlannerInput, nil
}
