package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/SingerOS/backend/tools"
	testtools "github.com/insmtx/SingerOS/backend/tools/test"
)

func TestNewServerRegistersOnlyEchoTool(t *testing.T) {
	srv := NewServer()

	if srv.GetTool(testtools.ToolNameEcho) == nil {
		t.Fatalf("expected %s to be registered", testtools.ToolNameEcho)
	}
	if srv.GetTool("node_shell") != nil {
		t.Fatalf("node tools must not be registered in the MCP server")
	}
	if srv.GetTool("skill_use") != nil {
		t.Fatalf("skill tools must not be registered in the MCP server")
	}
}

func TestHandleEcho(t *testing.T) {
	tool := testtools.NewEchoTool()

	output, err := tool.Execute(context.Background(), map[string]any{
		"message": " hello ",
	})
	if err != nil {
		t.Fatalf("execute echo: %v", err)
	}

	var structured struct {
		Message string `json:"message"`
		Server  string `json:"server"`
	}
	if err := json.Unmarshal([]byte(output), &structured); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if structured.Message != "hello" {
		t.Fatalf("expected trimmed message, got %q", structured.Message)
	}
	if structured.Server != serverName {
		t.Fatalf("expected server %q, got %q", serverName, structured.Server)
	}
}

func TestHandleEchoMissingMessageReturnsToolError(t *testing.T) {
	tool := testtools.NewEchoTool()

	if err := tool.Validate(map[string]any{}); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestNewServerWithToolsUsesBackendToolInterface(t *testing.T) {
	srv := NewServerWithTools(&testPublicTool{
		BaseTool: tools.NewBaseTool("public_test", "public test tool", tools.Schema{Type: "object"}),
	})

	if srv.GetTool("public_test") == nil {
		t.Fatalf("expected backend/tools.Tool to be adapted into MCP")
	}
}

type testPublicTool struct {
	tools.BaseTool
}

func (t *testPublicTool) Execute(ctx context.Context, input map[string]interface{}) (string, error) {
	return `{"ok":true}`, nil
}

func TestRegisterRoutesMountsMCPHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/v1")
	RegisterRoutes(group, NewServer())

	req := httptest.NewRequest(http.MethodPost, "/v1/mcp", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code == http.StatusNotFound {
		t.Fatalf("expected /v1/mcp to be mounted")
	}
}

func TestRegisterRoutesRequiresTokenWhenConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/v1")
	RegisterRoutes(group, NewServer())

	req := httptest.NewRequest(http.MethodPost, "/v1/mcp", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without token, got %d", recorder.Code)
	}
}

func TestRegisterRoutesAcceptsBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/v1")
	RegisterRoutes(group, NewServer())

	req := httptest.NewRequest(http.MethodPost, "/v1/mcp", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(authorizationHeader, "Bearer "+DefaultAuthToken())
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code == http.StatusUnauthorized {
		t.Fatalf("expected bearer token to pass auth")
	}
}

func TestRegisterRoutesAcceptsAPIKeyHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/v1")
	RegisterRoutes(group, NewServer())

	req := httptest.NewRequest(http.MethodPost, "/v1/mcp", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(apiKeyHeader, DefaultAuthToken())
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code == http.StatusUnauthorized {
		t.Fatalf("expected X-API-Key token to pass auth")
	}
}
