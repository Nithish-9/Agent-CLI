package client

import (
	"context"
	"salesforce-ai-agent/configuration"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

func InitializeMCPClient(ctx context.Context, config *configuration.Config, appLogger *zap.Logger) (*MCPClient, error) {

	mcpclient := MCPClient{
		Client: mcp.NewClient(&mcp.Implementation{
			Name:    "AGENT-0018",
			Version: "v1.0.0",
		}, nil),
		ServerInfo: make(map[string]*MCPServerInfo),
		MCPServerPlannerInput: &MCPServerPlannerInput{
			MCPServer: make([]*MCPServerPlanner, 0),
		},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var connErrors []string

	for _, server := range config.MCPServers.Servers {
		wg.Add(1)
		go func(server configuration.MCPServer) {
			defer wg.Done()
			switch server.Transport {
			case "streamable-http":
				mcpclient.StreamableClientTransport(ctx, server.Name, server.URL, &mu, &connErrors, appLogger)
			case "stdio":
				mcpclient.CommandTransport(ctx, server.Name, server.URL, &mu, &connErrors, appLogger)
			case "sse":
				mcpclient.SSEClientTransport(ctx, server.Name, server.URL, &mu, &connErrors, appLogger)
			default:
				appLogger.Warn("unknown transport type — skipping server",
					zap.String("transport", server.Transport),
					zap.String("server", server.Name),
				)
			}
		}(server)
	}

	wg.Wait()

	return &mcpclient, nil
}
