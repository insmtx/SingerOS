// github 包提供 GitHub 平台的连接器实现
//
// 该包实现了与 GitHub 平台的集成，包括 Webhook 事件接收、
// OAuth 认证流程、用户信息同步等功能。
package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v78/github"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"

	auth "github.com/insmtx/SingerOS/backend/auth"
	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/internal/connectors"
	eventbus "github.com/insmtx/SingerOS/backend/internal/infra/mq"
	"github.com/insmtx/SingerOS/backend/types"
)

const (
	githubAPIBaseURL = "https://api.github.com"
)

var _ connectors.Connector = (*Connector)(nil)

// Connector implements the GitHub connector interface.
type Connector struct {
	cfg       config.GithubAppConfig
	client    *github.Client
	publisher eventbus.Publisher
	db        *gorm.DB
	authSvc   *auth.Service
}

// ChannelCode returns the channel identifier for GitHub.
func (Connector) ChannelCode() string {
	return "github"
}

// RegisterRoutes registers GitHub webhook and auth endpoints.
func (c *Connector) RegisterRoutes(r gin.IRouter) {
	r.POST("/github/webhook", c.handleWebhook)
	r.GET("/github/auth", c.oAuthRedirect)
	r.GET("/github/callback", c.oAuthCallback)
}

// NewConnector creates a new GitHub connector instance.
func NewConnector(cfg config.GithubAppConfig, publisher eventbus.Publisher, db *gorm.DB, authSvc *auth.Service) *Connector {
	logs.Infof("Creating new GitHub connector for app ID: %d", cfg.AppID)

	var githubClient *github.Client
	if cfg.AppID != 0 && cfg.PrivateKey != "" {
		logs.Debugf("GitHub connector initialized with app ID: %d", cfg.AppID)
	} else {
		logs.Warnf("GitHub connector initialized without authentication - limited functionality")
	}

	return &Connector{
		cfg:       cfg,
		client:    githubClient,
		publisher: publisher,
		db:        db,
		authSvc:   authSvc,
	}
}

// oAuthRedirect initiates the GitHub OAuth flow.
func (c *Connector) oAuthRedirect(ctx *gin.Context) {
	if c.authSvc == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "authorization service unavailable"})
		return
	}
	userID := ctx.Query("user_id")
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user_id parameter missing"})
		return
	}

	redirectURL, err := c.authSvc.StartAuthorization(ctx.Request.Context(), &auth.StartAuthorizationRequest{
		UserID:      userID,
		Provider:    auth.ProviderGitHub,
		RedirectURI: ctx.Query("redirect_uri"),
	})
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to start GitHub authorization: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// oAuthCallback handles the GitHub OAuth callback.
func (c *Connector) oAuthCallback(ctx *gin.Context) {
	if c.authSvc == nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"error": "authorization service unavailable"})
		return
	}

	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "code parameter missing"})
		return
	}

	state := ctx.Query("state")
	if state == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "state parameter missing"})
		return
	}

	result, err := c.authSvc.HandleAuthorizationCallback(ctx.Request.Context(), &auth.AuthorizationCallbackRequest{
		Provider: auth.ProviderGitHub,
		State:    state,
		Code:     code,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "Failed to complete GitHub authorization: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := c.buildOAuthResponse(result.Account)
	if err := c.saveUserIfNeeded(ctx, result.Account, response); err != nil {
		logs.ErrorContextf(ctx, "Failed to save user: %v", err)
	}

	ctx.JSON(http.StatusOK, response)
}

// buildOAuthResponse constructs the OAuth response.
func (c *Connector) buildOAuthResponse(account *auth.AuthorizedAccount) gin.H {
	user := gin.H{
		"github_id":    account.ExternalAccountID,
		"github_login": account.Metadata["github_login"],
		"name":         account.Metadata["name"],
		"email":        account.Metadata["email"],
	}

	return gin.H{
		"user":    user,
		"account": account,
	}
}

// saveUserIfNeeded saves user to database if available.
func (c *Connector) saveUserIfNeeded(ctx context.Context, account *auth.AuthorizedAccount, response gin.H) error {
	if c.db == nil {
		logs.WarnContext(ctx, "Database not available, user info will not be saved to DB")
		return nil
	}

	githubID, err := parseGithubID(account.ExternalAccountID)
	if err != nil {
		return err
	}

	newUser := &types.User{
		GithubID:    githubID,
		GithubLogin: account.Metadata["github_login"],
		Name:        account.Metadata["name"],
		Email:       account.Metadata["email"],
		AvatarURL:   account.Metadata["avatar_url"],
	}

	result := c.db.Where(types.User{GithubID: newUser.GithubID}).FirstOrCreate(newUser)
	if result.Error != nil {
		return result.Error
	}

	response["user"].(gin.H)["id"] = newUser.ID
	return nil
}

func parseGithubID(externalAccountID string) (int64, error) {
	var githubID int64
	_, err := fmt.Sscanf(externalAccountID, "%d", &githubID)
	if err != nil {
		return 0, fmt.Errorf("parse github id: %w", err)
	}
	return githubID, nil
}
