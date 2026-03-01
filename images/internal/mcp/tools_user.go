package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/jjenkins/labnocturne/images/internal/service"
	"github.com/jjenkins/labnocturne/images/internal/store"
)

func registerUserTools(s *server.MCPServer, deps *Deps) {
	s.AddTool(listFilesTool(), handleListFiles(deps))
	s.AddTool(getFileTool(), handleGetFile(deps))
	s.AddTool(deleteFileTool(), handleDeleteFile(deps))
	s.AddTool(getStatsTool(), handleGetStats(deps))
	s.AddTool(generateTestKeyTool(), handleGenerateTestKey(deps))
}

// --- list_files ---

func listFilesTool() mcpgo.Tool {
	return mcpgo.NewTool("list_files",
		mcpgo.WithDescription("List your uploaded files with pagination and sorting"),
		mcpgo.WithNumber("limit", mcpgo.Description("Max files to return (1-1000, default 100)")),
		mcpgo.WithNumber("offset", mcpgo.Description("Number of files to skip (default 0)")),
		mcpgo.WithString("sort", mcpgo.Description("Sort order: uploaded_at_desc, uploaded_at_asc, size_desc, size_asc")),
	)
}

func handleListFiles(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		user, err := RequireAuth(ctx)
		if err != nil {
			return toolError(err.Error()), nil
		}

		userStore := store.NewUserStore(deps.DB)
		fileStore := store.NewFileStore(deps.DB)
		fileSvc := service.NewFileService(userStore, fileStore, deps.S3Client, deps.BaseURL, deps.S3Bucket)

		params := service.ListFilesParams{
			Limit:     req.GetInt("limit", 100),
			Offset:    req.GetInt("offset", 0),
			SortOrder: req.GetString("sort", "uploaded_at_desc"),
		}

		result, err := fileSvc.ListFiles(ctx, user.ID.String(), params)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to list files: %v", err)), nil
		}

		type fileItem struct {
			ID         string `json:"id"`
			ExternalID string `json:"external_id"`
			Filename   string `json:"filename"`
			SizeBytes  int64  `json:"size_bytes"`
			MimeType   string `json:"mime_type"`
			CDNURL     string `json:"cdn_url"`
			UploadedAt string `json:"uploaded_at"`
		}

		files := make([]fileItem, 0, len(result.Files))
		for _, f := range result.Files {
			files = append(files, fileItem{
				ID:         f.ID,
				ExternalID: f.ExternalID,
				Filename:   f.Filename,
				SizeBytes:  f.SizeBytes,
				MimeType:   f.MimeType,
				CDNURL:     f.CDNURL,
				UploadedAt: f.UploadedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		resp := map[string]any{
			"files":  files,
			"total":  result.Total,
			"limit":  result.Limit,
			"offset": result.Offset,
		}

		return jsonResult(resp)
	}
}

// --- get_file ---

func getFileTool() mcpgo.Tool {
	return mcpgo.NewTool("get_file",
		mcpgo.WithDescription("Get details of a specific file by its ID (ULID or external ID like img_...)"),
		mcpgo.WithString("id", mcpgo.Required(), mcpgo.Description("File ID (ULID or img_... external ID)")),
	)
}

func handleGetFile(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if _, err := RequireAuth(ctx); err != nil {
			return toolError(err.Error()), nil
		}

		id, err := req.RequireString("id")
		if err != nil {
			return toolError("id parameter is required"), nil
		}

		userStore := store.NewUserStore(deps.DB)
		fileStore := store.NewFileStore(deps.DB)
		fileSvc := service.NewFileService(userStore, fileStore, deps.S3Client, deps.BaseURL, deps.S3Bucket)

		file, err := fileSvc.GetByULID(ctx, id)
		if err != nil {
			return toolError(fmt.Sprintf("File not found: %v", err)), nil
		}

		resp := map[string]any{
			"id":          file.ID,
			"external_id": file.ExternalID,
			"filename":    file.Filename,
			"extension":   file.Extension,
			"size_bytes":  file.SizeBytes,
			"mime_type":   file.MimeType,
			"cdn_url":     file.CDNURL,
			"uploaded_at": file.UploadedAt.Format("2006-01-02T15:04:05Z"),
		}

		return jsonResult(resp)
	}
}

// --- delete_file ---

func deleteFileTool() mcpgo.Tool {
	return mcpgo.NewTool("delete_file",
		mcpgo.WithDescription("Soft-delete a file you own (sets deleted_at, recoverable for 30 days)"),
		mcpgo.WithString("id", mcpgo.Required(), mcpgo.Description("File ID (ULID or img_... external ID)")),
	)
}

func handleDeleteFile(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		user, err := RequireAuth(ctx)
		if err != nil {
			return toolError(err.Error()), nil
		}

		id, err := req.RequireString("id")
		if err != nil {
			return toolError("id parameter is required"), nil
		}

		userStore := store.NewUserStore(deps.DB)
		fileStore := store.NewFileStore(deps.DB)
		fileSvc := service.NewFileService(userStore, fileStore, deps.S3Client, deps.BaseURL, deps.S3Bucket)

		if err := fileSvc.DeleteFile(ctx, id, user.ID.String()); err != nil {
			return toolError(fmt.Sprintf("Failed to delete file: %v", err)), nil
		}

		return mcpgo.NewToolResultText("File deleted successfully"), nil
	}
}

// --- get_stats ---

func getStatsTool() mcpgo.Tool {
	return mcpgo.NewTool("get_stats",
		mcpgo.WithDescription("Get your usage statistics: storage, files, bandwidth, and quotas"),
	)
}

func handleGetStats(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		user, err := RequireAuth(ctx)
		if err != nil {
			return toolError(err.Error()), nil
		}

		userStore := store.NewUserStore(deps.DB)
		bandwidthStore := store.NewBandwidthStore(deps.DB)
		statsSvc := service.NewStatsService(userStore, bandwidthStore, deps.RateLimiter)

		stats, err := statsSvc.GetUsageStats(ctx, user)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to get stats: %v", err)), nil
		}

		return jsonResult(stats)
	}
}

// --- generate_test_key ---

func generateTestKeyTool() mcpgo.Tool {
	return mcpgo.NewTool("generate_test_key",
		mcpgo.WithDescription("Generate a new ln_test_* API key for testing (no signup required)"),
	)
}

func handleGenerateTestKey(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		// This tool doesn't require auth - anyone can generate a test key
		userStore := store.NewUserStore(deps.DB)
		keySvc := service.NewKeyService(userStore)

		user, err := keySvc.GenerateTestKey(ctx)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to generate test key: %v", err)), nil
		}

		resp := map[string]any{
			"api_key":    user.APIKey,
			"key_type":   user.KeyType,
			"plan":       user.Plan,
			"created_at": user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}

		return jsonResult(resp)
	}
}

// jsonResult marshals data to JSON and returns it as a text tool result.
func jsonResult(data any) (*mcpgo.CallToolResult, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return toolError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}
	return mcpgo.NewToolResultText(string(b)), nil
}
