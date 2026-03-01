package mcp

import (
	"context"
	"fmt"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/jjenkins/labnocturne/images/internal/service"
	"github.com/jjenkins/labnocturne/images/internal/store"
)

func registerAdminTools(s *server.MCPServer, deps *Deps) {
	s.AddTool(listUsersTool(), handleListUsers(deps))
	s.AddTool(getUserTool(), handleGetUser(deps))
	s.AddTool(deleteUserFilesTool(), handleDeleteUserFiles(deps))
	s.AddTool(getSystemStatsTool(), handleGetSystemStats(deps))
	s.AddTool(runCleanupTool(), handleRunCleanup(deps))
}

// --- list_users ---

func listUsersTool() mcpgo.Tool {
	return mcpgo.NewTool("list_users",
		mcpgo.WithDescription("[Admin] List all users with pagination"),
		mcpgo.WithNumber("limit", mcpgo.Description("Max users to return (1-100, default 50)")),
		mcpgo.WithNumber("offset", mcpgo.Description("Number of users to skip (default 0)")),
	)
}

func handleListUsers(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if _, err := RequireAdmin(ctx); err != nil {
			return toolError(err.Error()), nil
		}

		limit := req.GetInt("limit", 50)
		if limit > 100 {
			limit = 100
		}
		offset := req.GetInt("offset", 0)

		userStore := store.NewUserStore(deps.DB)
		users, total, err := userStore.ListAll(ctx, limit, offset)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to list users: %v", err)), nil
		}

		type userItem struct {
			ID        string  `json:"id"`
			KeyType   string  `json:"key_type"`
			Plan      string  `json:"plan"`
			Email     *string `json:"email,omitempty"`
			CreatedAt string  `json:"created_at"`
		}

		items := make([]userItem, 0, len(users))
		for _, u := range users {
			items = append(items, userItem{
				ID:        u.ID.String(),
				KeyType:   u.KeyType,
				Plan:      u.Plan,
				Email:     u.Email,
				CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		resp := map[string]any{
			"users":  items,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		}

		return jsonResult(resp)
	}
}

// --- get_user ---

func getUserTool() mcpgo.Tool {
	return mcpgo.NewTool("get_user",
		mcpgo.WithDescription("[Admin] Get detailed information about a specific user"),
		mcpgo.WithString("user_id", mcpgo.Required(), mcpgo.Description("User UUID")),
	)
}

func handleGetUser(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if _, err := RequireAdmin(ctx); err != nil {
			return toolError(err.Error()), nil
		}

		userID, err := req.RequireString("user_id")
		if err != nil {
			return toolError("user_id parameter is required"), nil
		}

		userStore := store.NewUserStore(deps.DB)
		user, err := userStore.FindByID(ctx, userID)
		if err != nil {
			return toolError(fmt.Sprintf("User not found: %v", err)), nil
		}

		// Get storage usage
		totalBytes, fileCount, err := userStore.GetStorageUsage(ctx, user.ID.String())
		if err != nil {
			return toolError(fmt.Sprintf("Failed to get storage usage: %v", err)), nil
		}

		resp := map[string]any{
			"id":         user.ID.String(),
			"key_type":   user.KeyType,
			"plan":       user.Plan,
			"email":      user.Email,
			"created_at": user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updated_at": user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			"storage": map[string]any{
				"used_bytes": totalBytes,
				"used_mb":    float64(totalBytes) / 1024 / 1024,
				"file_count": fileCount,
			},
		}

		return jsonResult(resp)
	}
}

// --- delete_user_files ---

func deleteUserFilesTool() mcpgo.Tool {
	return mcpgo.NewTool("delete_user_files",
		mcpgo.WithDescription("[Admin] Soft-delete ALL files belonging to a user"),
		mcpgo.WithString("user_id", mcpgo.Required(), mcpgo.Description("User UUID whose files to delete")),
	)
}

func handleDeleteUserFiles(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if _, err := RequireAdmin(ctx); err != nil {
			return toolError(err.Error()), nil
		}

		userID, err := req.RequireString("user_id")
		if err != nil {
			return toolError("user_id parameter is required"), nil
		}

		fileStore := store.NewFileStore(deps.DB)
		count, err := fileStore.SoftDeleteAllByUserID(ctx, userID)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to delete user files: %v", err)), nil
		}

		resp := map[string]any{
			"deleted_count": count,
			"user_id":       userID,
			"message":       fmt.Sprintf("Soft-deleted %d files for user %s", count, userID),
		}

		return jsonResult(resp)
	}
}

// --- get_system_stats ---

func getSystemStatsTool() mcpgo.Tool {
	return mcpgo.NewTool("get_system_stats",
		mcpgo.WithDescription("[Admin] Get system-wide statistics: total users, files, and storage"),
	)
}

func handleGetSystemStats(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if _, err := RequireAdmin(ctx); err != nil {
			return toolError(err.Error()), nil
		}

		var totalUsers int64
		var totalFiles int64
		var totalBytes int64
		var deletedFiles int64

		err := deps.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&totalUsers)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to count users: %v", err)), nil
		}

		err = deps.DB.QueryRowContext(ctx, `
			SELECT
				COALESCE(COUNT(*), 0),
				COALESCE(SUM(size_bytes), 0)
			FROM files
			WHERE deleted_at IS NULL
		`).Scan(&totalFiles, &totalBytes)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to get file stats: %v", err)), nil
		}

		err = deps.DB.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM files WHERE deleted_at IS NOT NULL
		`).Scan(&deletedFiles)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to count deleted files: %v", err)), nil
		}

		resp := map[string]any{
			"users": map[string]any{
				"total": totalUsers,
			},
			"files": map[string]any{
				"active":       totalFiles,
				"soft_deleted": deletedFiles,
			},
			"storage": map[string]any{
				"used_bytes": totalBytes,
				"used_mb":    float64(totalBytes) / 1024 / 1024,
				"used_gb":    float64(totalBytes) / 1024 / 1024 / 1024,
			},
		}

		return jsonResult(resp)
	}
}

// --- run_cleanup ---

func runCleanupTool() mcpgo.Tool {
	return mcpgo.NewTool("run_cleanup",
		mcpgo.WithDescription("[Admin] Run cleanup: expire test files (7+ days) and permanently remove soft-deleted files (30+ days)"),
		mcpgo.WithBoolean("dry_run", mcpgo.Description("If true, only report what would be deleted without actually deleting (default false)")),
	)
}

func handleRunCleanup(deps *Deps) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		if _, err := RequireAdmin(ctx); err != nil {
			return toolError(err.Error()), nil
		}

		args := req.GetArguments()
		dryRun, _ := args["dry_run"].(bool)

		cleanupSvc := service.NewCleanupService(deps.DB, deps.S3Client, deps.S3Bucket, dryRun)

		testCount, testErr := cleanupSvc.CleanupExpiredTestFiles(ctx)
		softCount, softErr := cleanupSvc.CleanupExpiredSoftDeleted(ctx)

		resp := map[string]any{
			"dry_run":                    dryRun,
			"expired_test_files_deleted": testCount,
			"soft_deleted_files_removed": softCount,
		}

		if testErr != nil {
			resp["test_cleanup_error"] = testErr.Error()
		}
		if softErr != nil {
			resp["soft_delete_cleanup_error"] = softErr.Error()
		}

		return jsonResult(resp)
	}
}
