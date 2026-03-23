package daemon

import "github.com/gin-gonic/gin"

func registerRoutes(r *gin.Engine) {
	r.Use(gin.Logger())

	agent := r.Group("/agent")
	{
		agent.POST("/init", handlePostInit)
		agent.GET("/chat", handleChat)
		agent.POST("/stop", handleStop)
		agent.GET("/status", handleAgentStatus)
		agent.GET("/all", handleAgentAll)
	}
}
