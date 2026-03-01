package model

import (
	"time"

	"github.com/google/uuid"
)

// Key type constants
const (
	KeyTypeTest  = "test"
	KeyTypeLive  = "live"
	KeyTypeAdmin = "admin"
)

// User represents a user in the system with an API key
type User struct {
	ID               uuid.UUID
	APIKey           string
	Email            *string // Optional for test keys, UNIQUE for live keys
	KeyType          string  // "test" or "live" (determines API key prefix)
	Plan             string  // "test", "starter", or "pro" (determines quotas)
	StripeCustomerID *string // Optional, only for live keys
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
