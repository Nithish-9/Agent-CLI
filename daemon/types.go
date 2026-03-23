package daemon

import (
	"context"
	"salesforce-ai-agent/configuration"
	"salesforce-ai-agent/internal/llm"
	client "salesforce-ai-agent/internal/mcpclient"
	"sync"

	"go.uber.org/zap"
)

type AgentDaemon struct {
	mu            sync.RWMutex
	Config        *configuration.Config
	LLM           *llm.LLMModels
	MCPClient     *client.MCPClient
	Ctx           context.Context
	Cancel        context.CancelFunc
	Logger        *zap.Logger
	IsInitialized bool
	LoggerCleanup func()
}
