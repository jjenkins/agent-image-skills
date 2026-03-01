package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jjenkins/labnocturne/images/internal/model"
)

// UserStore handles database operations for users
type UserStore struct {
	db *sql.DB
}

// NewUserStore creates a new UserStore
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// Create inserts a new user into the database
func (s *UserStore) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (api_key, key_type, plan, email, stripe_customer_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRowContext(ctx, query,
		user.APIKey,
		user.KeyType,
		user.Plan,
		user.Email,
		user.StripeCustomerID,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

// FindByAPIKey retrieves a user by their API key
func (s *UserStore) FindByAPIKey(ctx context.Context, apiKey string) (*model.User, error) {
	query := `
		SELECT id, api_key, email, key_type, plan, stripe_customer_id, created_at, updated_at
		FROM users
		WHERE api_key = $1
	`

	user := &model.User{}
	err := s.db.QueryRowContext(ctx, query, apiKey).Scan(
		&user.ID,
		&user.APIKey,
		&user.Email,
		&user.KeyType,
		&user.Plan,
		&user.StripeCustomerID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid API key")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

// GetStorageUsage returns total storage used and file count for a user's non-deleted files
func (s *UserStore) GetStorageUsage(ctx context.Context, userID string) (int64, int64, error) {
	query := `
		SELECT
			COALESCE(SUM(size_bytes), 0) as total_bytes,
			COUNT(*) as file_count
		FROM files
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	var totalBytes int64
	var fileCount int64
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&totalBytes, &fileCount)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to calculate storage: %w", err)
	}

	return totalBytes, fileCount, nil
}

// FindByEmail retrieves a user by their email address
func (s *UserStore) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, api_key, email, key_type, plan, stripe_customer_id, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	user := &model.User{}
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.APIKey,
		&user.Email,
		&user.KeyType,
		&user.Plan,
		&user.StripeCustomerID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}

	return user, nil
}

// FindByID retrieves a user by their UUID
func (s *UserStore) FindByID(ctx context.Context, userID string) (*model.User, error) {
	query := `
		SELECT id, api_key, email, key_type, plan, stripe_customer_id, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &model.User{}
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.APIKey,
		&user.Email,
		&user.KeyType,
		&user.Plan,
		&user.StripeCustomerID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

// ListAll retrieves all users with pagination
func (s *UserStore) ListAll(ctx context.Context, limit int, offset int) ([]*model.User, int64, error) {
	// Get total count
	var total int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	query := `
		SELECT id, api_key, email, key_type, plan, stripe_customer_id, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
			&user.ID,
			&user.APIKey,
			&user.Email,
			&user.KeyType,
			&user.Plan,
			&user.StripeCustomerID,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}

	return users, total, nil
}

// UpdatePlan updates a user's plan (for upgrades/downgrades)
func (s *UserStore) UpdatePlan(ctx context.Context, userID string, plan string) error {
	query := `
		UPDATE users
		SET plan = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := s.db.ExecContext(ctx, query, plan, userID)
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
