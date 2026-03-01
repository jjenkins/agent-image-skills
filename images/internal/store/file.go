package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jjenkins/labnocturne/images/internal/model"
)

// FileStore handles database operations for files
type FileStore struct {
	db *sql.DB
}

// NewFileStore creates a new FileStore
func NewFileStore(db *sql.DB) *FileStore {
	return &FileStore{db: db}
}

// Create inserts a new file record into the database
func (s *FileStore) Create(ctx context.Context, file *model.File) error {
	query := `
		INSERT INTO files (id, external_id, user_id, filename, extension, size_bytes, mime_type, s3_key, cdn_url, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING uploaded_at
	`

	err := s.db.QueryRowContext(ctx, query,
		file.ID,
		file.ExternalID,
		file.UserID,
		file.Filename,
		file.Extension,
		file.SizeBytes,
		file.MimeType,
		file.S3Key,
		file.CDNURL,
	).Scan(&file.UploadedAt)

	if err != nil {
		return fmt.Errorf("failed to insert file: %w", err)
	}

	return nil
}

// FindByID retrieves a file by its ULID
func (s *FileStore) FindByID(ctx context.Context, ulid string) (*model.File, error) {
	query := `
		SELECT id, external_id, user_id, filename, extension, size_bytes, mime_type, s3_key, cdn_url, uploaded_at, deleted_at
		FROM files
		WHERE id = $1 AND deleted_at IS NULL
	`

	file := &model.File{}
	err := s.db.QueryRowContext(ctx, query, ulid).Scan(
		&file.ID,
		&file.ExternalID,
		&file.UserID,
		&file.Filename,
		&file.Extension,
		&file.SizeBytes,
		&file.MimeType,
		&file.S3Key,
		&file.CDNURL,
		&file.UploadedAt,
		&file.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("file not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query file: %w", err)
	}

	return file, nil
}

// SoftDelete marks a file as deleted by setting deleted_at timestamp
func (s *FileStore) SoftDelete(ctx context.Context, ulid string) error {
	query := `
		UPDATE files
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, ulid)
	if err != nil {
		return fmt.Errorf("failed to soft delete file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found or already deleted")
	}

	return nil
}

// SoftDeleteAllByUserID soft-deletes all non-deleted files for a user
func (s *FileStore) SoftDeleteAllByUserID(ctx context.Context, userID string) (int64, error) {
	query := `
		UPDATE files
		SET deleted_at = NOW()
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to soft delete user files: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to check rows affected: %w", err)
	}

	return rowsAffected, nil
}

// FindByUserID retrieves files for a user with pagination and sorting
// The sortOrder parameter must be validated by the caller (service layer) to prevent SQL injection
func (s *FileStore) FindByUserID(ctx context.Context, userID string, limit int, offset int, sortOrder string) ([]*model.File, int64, error) {
	// Build ORDER BY clause (safe: sortOrder is validated in service layer)
	var orderBy string
	switch sortOrder {
	case "uploaded_at_asc":
		orderBy = "uploaded_at ASC"
	case "size_desc":
		orderBy = "size_bytes DESC, uploaded_at DESC"
	case "size_asc":
		orderBy = "size_bytes ASC, uploaded_at DESC"
	default: // "uploaded_at_desc"
		orderBy = "uploaded_at DESC"
	}

	// Get total count
	var total int64
	countQuery := `
		SELECT COUNT(*)
		FROM files
		WHERE user_id = $1 AND deleted_at IS NULL
	`
	err := s.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", err)
	}

	// Get paginated files
	query := fmt.Sprintf(`
		SELECT id, external_id, user_id, filename, extension, size_bytes, mime_type, s3_key, cdn_url, uploaded_at, deleted_at
		FROM files
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY %s
		LIMIT $2 OFFSET $3
	`, orderBy)

	rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var files []*model.File
	for rows.Next() {
		file := &model.File{}
		err := rows.Scan(
			&file.ID,
			&file.ExternalID,
			&file.UserID,
			&file.Filename,
			&file.Extension,
			&file.SizeBytes,
			&file.MimeType,
			&file.S3Key,
			&file.CDNURL,
			&file.UploadedAt,
			&file.DeletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating files: %w", err)
	}

	return files, total, nil
}

// FindExpiredTestFiles retrieves test key files older than 7 days (not yet deleted)
func (s *FileStore) FindExpiredTestFiles(ctx context.Context) ([]*model.File, error) {
	query := `
		SELECT f.id, f.external_id, f.user_id, f.filename, f.extension, f.size_bytes, f.mime_type, f.s3_key, f.cdn_url, f.uploaded_at, f.deleted_at
		FROM files f
		INNER JOIN users u ON f.user_id = u.id
		WHERE u.key_type = 'test'
		  AND f.uploaded_at < NOW() - INTERVAL '7 days'
		  AND f.deleted_at IS NULL
		ORDER BY f.uploaded_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired test files: %w", err)
	}
	defer rows.Close()

	var files []*model.File
	for rows.Next() {
		file := &model.File{}
		err := rows.Scan(
			&file.ID,
			&file.ExternalID,
			&file.UserID,
			&file.Filename,
			&file.Extension,
			&file.SizeBytes,
			&file.MimeType,
			&file.S3Key,
			&file.CDNURL,
			&file.UploadedAt,
			&file.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expired test file: %w", err)
		}
		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expired test files: %w", err)
	}

	return files, nil
}

// FindExpiredSoftDeleted retrieves soft-deleted files older than 30 days
func (s *FileStore) FindExpiredSoftDeleted(ctx context.Context) ([]*model.File, error) {
	query := `
		SELECT id, external_id, user_id, filename, extension, size_bytes, mime_type, s3_key, cdn_url, uploaded_at, deleted_at
		FROM files
		WHERE deleted_at < NOW() - INTERVAL '30 days'
		ORDER BY deleted_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query expired soft-deleted files: %w", err)
	}
	defer rows.Close()

	var files []*model.File
	for rows.Next() {
		file := &model.File{}
		err := rows.Scan(
			&file.ID,
			&file.ExternalID,
			&file.UserID,
			&file.Filename,
			&file.Extension,
			&file.SizeBytes,
			&file.MimeType,
			&file.S3Key,
			&file.CDNURL,
			&file.UploadedAt,
			&file.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expired soft-deleted file: %w", err)
		}
		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expired soft-deleted files: %w", err)
	}

	return files, nil
}

// PermanentlyDelete permanently deletes a file record from the database (hard delete)
func (s *FileStore) PermanentlyDelete(ctx context.Context, ulid string) error {
	query := `DELETE FROM files WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, ulid)
	if err != nil {
		return fmt.Errorf("failed to permanently delete file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found")
	}

	return nil
}
