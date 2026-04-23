package gitlab

import (
	"github.com/gin-gonic/gin"
	"github.com/ygpkg/yg-go/logs"
)

func (c *GitlabConnector) RegisterRoutes(r gin.IRouter) {
	r.POST("/gitlab/webhook", c.HandleWebhook)
}

func (c *GitlabConnector) HandleWebhook(ctx *gin.Context) {
	eventType := ctx.GetHeader("X-Gitlab-Event")
	if eventType == "" {
		logs.ErrorContext(ctx, "Missing X-Gitlab-Event header")
		ctx.JSON(400, gin.H{"error": "Missing X-Gitlab-Event header"})
		return
	}

	logs.InfoContextf(ctx, "Received GitLab event: %s", eventType)

	payload, err := ctx.GetRawData()
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to read request body: %v", err)
		ctx.JSON(400, gin.H{"error": "Failed to read request body"})
		return
	}

	if err := c.verifySignature(ctx, payload); err != nil {
		logs.ErrorContextf(ctx, "Signature verification failed: %v", err)
		ctx.JSON(403, gin.H{"error": "Invalid signature"})
		return
	}

	if err := c.processEvent(ctx, eventType, payload); err != nil {
		logs.ErrorContextf(ctx, "Failed to process event: %v", err)
		ctx.JSON(500, gin.H{"error": "Failed to process event"})
		return
	}

	ctx.JSON(200, gin.H{"status": "ok"})
}

func (c *GitlabConnector) verifySignature(ctx *gin.Context, payload []byte) error {
	return nil
}

func (c *GitlabConnector) processEvent(ctx *gin.Context, eventType string, payload []byte) error {
	return nil
}
