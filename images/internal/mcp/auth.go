package mcp

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/jjenkins/labnocturne/images/internal/model"
	"github.com/jjenkins/labnocturne/images/internal/store"
)

type contextKey string

const userContextKey contextKey = "mcp_user"

// UserFromContext extracts the authenticated user from the context.
func UserFromContext(ctx context.Context) *model.User {
	user, _ := ctx.Value(userContextKey).(*model.User)
	return user
}

// RequireAuth returns the user from context or an error if not authenticated.
func RequireAuth(ctx context.Context) (*model.User, error) {
	user := UserFromContext(ctx)
	if user == nil {
		return nil, fmt.Errorf("authentication required: provide a valid API key")
	}
	return user, nil
}

// RequireAdmin returns the user from context if they have an admin key, or an error.
func RequireAdmin(ctx context.Context) (*model.User, error) {
	user, err := RequireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if user.KeyType != model.KeyTypeAdmin {
		return nil, fmt.Errorf("admin access required: this tool requires an ln_admin_* API key")
	}
	return user, nil
}

// SSEContextFunc returns a function that extracts the Bearer token from the
// HTTP request, validates it against the database, and injects the user into context.
func SSEContextFunc(db *sql.DB) func(ctx context.Context, r *http.Request) context.Context {
	return func(ctx context.Context, r *http.Request) context.Context {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return ctx
		}

		apiKey := strings.TrimPrefix(authHeader, "Bearer ")

		userStore := store.NewUserStore(db)
		user, err := userStore.FindByAPIKey(ctx, apiKey)
		if err != nil {
			return ctx
		}

		return context.WithValue(ctx, userContextKey, user)
	}
}

// StdioContextFunc returns a function that injects a pre-authenticated user
// into every request context. Used for stdio transport where the API key is
// validated once at startup.
func StdioContextFunc(user *model.User) func(ctx context.Context) context.Context {
	return func(ctx context.Context) context.Context {
		return context.WithValue(ctx, userContextKey, user)
	}
}
