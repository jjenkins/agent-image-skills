package mcp

import (
	"database/sql"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jjenkins/labnocturne/images/internal/ratelimit"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Deps holds shared dependencies for MCP tool handlers.
type Deps struct {
	DB          *sql.DB
	S3Client    *s3.Client
	S3Bucket    string
	BaseURL     string
	RateLimiter *ratelimit.Limiter
}

// NewMCPServer creates a configured MCP server with all tools registered.
func NewMCPServer(deps *Deps) *server.MCPServer {
	s := server.NewMCPServer(
		"Lab Nocturne Images",
		"0.1.0",
		server.WithToolCapabilities(false),
		server.WithInstructions("Lab Nocturne Images API - manage images and storage via MCP. Authenticate with your API key (ln_test_*, ln_live_*, or ln_admin_*)."),
	)

	registerUserTools(s, deps)
	registerAdminTools(s, deps)

	return s
}

// toolError returns a CallToolResult representing an error.
func toolError(msg string) *mcpgo.CallToolResult {
	return mcpgo.NewToolResultError(msg)
}
