package daemon

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

func getBasePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/.agent"
	}
	return filepath.Join(home, ".agent")
}

func getPidFilePath() string {
	return filepath.Join(getBasePath(), "agent.pid")
}

func closeMCPSessions() {
	if agentDaemon.MCPClient == nil {
		return
	}
	for _, serverInfo := range agentDaemon.MCPClient.ServerInfo {
		if serverInfo.Session != nil {
			serverInfo.Session.Close()
			agentDaemon.Logger.Info("closed MCP session",
				zap.String("server", serverInfo.Metadata.Name),
			)
		}
	}
}
