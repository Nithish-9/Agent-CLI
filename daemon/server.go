package daemon

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"salesforce-ai-agent/logger"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const PORT = ":9999"
const URL = "http://localhost"

var agentDaemon *AgentDaemon

func Start() error {
	agentDaemon = &AgentDaemon{}
	agentDaemon.Ctx, agentDaemon.Cancel = context.WithCancel(context.Background())
	defer agentDaemon.Cancel()

	agentDaemon.Logger, agentDaemon.LoggerCleanup = logger.NewLogger("agent-0018")
	defer agentDaemon.LoggerCleanup()

	agentDaemon.Logger.Info("agent starting")

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		agentDaemon.Logger.Info("shutting down agent-0018")
		agentDaemon.Cancel()
	}()

	pid := os.Getpid()
	os.WriteFile(getPidFilePath(), []byte(strconv.Itoa(pid)), 0644)
	defer os.Remove(getPidFilePath())

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	registerRoutes(r)
	srv := &http.Server{Addr: PORT, Handler: r}

	go func() {
		<-agentDaemon.Ctx.Done()
		agentDaemon.Logger.Info("server shutting down gracefully")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		srv.Shutdown(shutdownCtx)
		agentDaemon.Logger.Info("server stopped")
	}()

	agentDaemon.Logger.Info("Agent 0018 server running", zap.String("port", PORT))
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
