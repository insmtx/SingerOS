package github

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v78/github"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"

	"github.com/insmtx/SingerOS/backend/config"
	"github.com/insmtx/SingerOS/backend/interaction/connectors"
	"github.com/insmtx/SingerOS/backend/interaction/eventbus"
	"github.com/insmtx/SingerOS/backend/types"
)

const (
	githubAuthURL    = "https://github.com/login/oauth/authorize"
	githubTokenURL   = "https://github.com/login/oauth/access_token"
	githubAPIBaseURL = "https://api.github.com"
	defaultScope     = "user:email"
	stateLength      = 16
)

var _ connectors.Connector = (*Connector)(nil)

// Connector implements the GitHub connector interface.
type Connector struct {
	cfg       config.GithubAppConfig
	client    *github.Client
	publisher eventbus.Publisher
	db        *gorm.DB
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
func NewConnector(cfg config.GithubAppConfig, publisher eventbus.Publisher, db *gorm.DB) *Connector {
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
	}
}

// oAuthRedirect initiates the GitHub OAuth flow.
func (c *Connector) oAuthRedirect(ctx *gin.Context) {
	state, err := generateState()
	if err != nil {
		logs.Errorf("Failed to generate state: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if c.cfg.ClientID == "" {
		logs.Errorf("Missing GitHub OAuth Client ID")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "GitHub OAuth not properly configured"})
		return
	}

	redirectURL := fmt.Sprintf(
		"%s?client_id=%s&state=%s&scope=%s&redirect_uri=%s",
		githubAuthURL,
		c.cfg.ClientID,
		state,
		defaultScope,
		"", // Redirect URI should be configured
	)
	ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// oAuthCallback handles the GitHub OAuth callback.
func (c *Connector) oAuthCallback(ctx *gin.Context) {
	code := ctx.Query("code")
	if code == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "code parameter missing"})
		return
	}

	if c.cfg.ClientID == "" || c.cfg.ClientSecret == "" {
		logs.Errorf("Missing GitHub OAuth client credentials")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "GitHub OAuth not properly configured"})
		return
	}

	accessToken, err := c.exchangeCodeForToken(code)
	if err != nil {
		logs.Errorf("Failed to exchange code for token: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get access token"})
		return
	}

	ghClient := github.NewTokenClient(context.Background(), accessToken)
	user, _, err := ghClient.Users.Get(context.Background(), "")
	if err != nil {
		logs.Errorf("Failed to get user profile: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user profile"})
		return
	}

	response := c.buildOAuthResponse(user, accessToken)
	if err := c.saveUserIfNeeded(user, response); err != nil {
		logs.Errorf("Failed to save user: %v", err)
	}

	ctx.JSON(http.StatusOK, response)
}

// exchangeCodeForToken exchanges OAuth code for access token.
func (c *Connector) exchangeCodeForToken(code string) (string, error) {
	requestData := fmt.Sprintf(
		"code=%s&client_id=%s&client_secret=%s&redirect_uri=%s",
		code,
		c.cfg.ClientID,
		c.cfg.ClientSecret,
		"",
	)

	req, err := http.NewRequest("POST", githubTokenURL, strings.NewReader(requestData))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("access token not found in response: %s", string(body))
	}

	return tokenResp.AccessToken, nil
}

// buildOAuthResponse constructs the OAuth response.
func (c *Connector) buildOAuthResponse(user *github.User, accessToken string) gin.H {
	return gin.H{
		"user": gin.H{
			"github_id":    user.GetID(),
			"github_login": user.GetLogin(),
			"name":         user.GetName(),
			"email":        user.GetEmail(),
		},
		"access_token": accessToken,
	}
}

// saveUserIfNeeded saves user to database if available.
func (c *Connector) saveUserIfNeeded(user *github.User, response gin.H) error {
	if c.db == nil {
		logs.Warn("Database not available, user info will not be saved to DB")
		return nil
	}

	newUser := &types.User{
		GithubID:    user.GetID(),
		GithubLogin: user.GetLogin(),
		Name:        user.GetName(),
		Email:       user.GetEmail(),
		AvatarURL:   user.GetAvatarURL(),
		Bio:         user.GetBio(),
		Company:     user.GetCompany(),
		Location:    user.GetLocation(),
		PublicRepos: user.GetPublicRepos(),
		Followers:   user.GetFollowers(),
	}

	result := c.db.Where(types.User{GithubID: newUser.GithubID}).FirstOrCreate(newUser)
	if result.Error != nil {
		return result.Error
	}

	response["user"].(gin.H)["id"] = newUser.ID
	return nil
}

// generateState creates a random state string for OAuth.
func generateState() (string, error) {
	b := make([]byte, stateLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
