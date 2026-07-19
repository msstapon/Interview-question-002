package domain

import (
	"context"
	"time"
)

// User is the account entity. Password is stored only as an Argon2id hash.
type User struct {
	ID           string `gorm:"type:uuid;primaryKey"`
	Username     string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserRepository is the persistence port for users (implemented under repository/).
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
}
