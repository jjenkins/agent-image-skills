package cmd

import (
	"log"
	"os"

	mcpinternal "github.com/jjenkins/labnocturne/images/internal/mcp"
	"github.com/jjenkins/labnocturne/images/internal/ratelimit"
	"github.com/jjenkins/labnocturne/images/internal/store"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

var mcpAPIKey string

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server over stdio for Claude Code / Claude Desktop",
	Long: `Start an MCP (Model Context Protocol) server that communicates over
stdin/stdout. This allows Claude Code and Claude Desktop to manage images
using your API key.

Example:
  ./bin/images mcp --api-key ln_test_abc123

Or set the LN_API_KEY environment variable:
  export LN_API_KEY=ln_test_abc123
  ./bin/images mcp`,
	Run: runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().StringVar(&mcpAPIKey, "api-key", "", "API key for authentication (or set LN_API_KEY env var)")
}

func runMCP(cmd *cobra.Command, args []string) {
	// Resolve API key from flag or env
	apiKey := mcpAPIKey
	if apiKey == "" {
		apiKey = os.Getenv("LN_API_KEY")
	}
	if apiKey == "" {
		log.Fatal("API key required: use --api-key flag or set LN_API_KEY environment variable")
	}

	// Database connection
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := store.NewDB(dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Validate API key
	userStore := store.NewUserStore(db)
	user, err := userStore.FindByAPIKey(cmd.Context(), apiKey)
	if err != nil {
		log.Fatalf("Invalid API key: %v", err)
	}

	// Build MCP server
	deps := &mcpinternal.Deps{
		DB:          db,
		BaseURL:     os.Getenv("BASE_URL"),
		S3Bucket:    os.Getenv("AWS_S3_BUCKET"),
		RateLimiter: ratelimit.NewLimiter(),
	}

	// S3 client is optional for stdio (admin tools like run_cleanup need it,
	// but most tools work without it)
	mcpSrv := mcpinternal.NewMCPServer(deps)

	// Serve over stdio with pre-authenticated user
	if err := server.ServeStdio(mcpSrv,
		server.WithStdioContextFunc(mcpinternal.StdioContextFunc(user)),
	); err != nil {
		log.Fatalf("MCP stdio server error: %v", err)
	}
}
