package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	UUID	 uuid.UUID
	Email    string
	Password string
	Username     string
	Balance int
	CreatedAt time.Time
}

type UserService interface {
	Authenticate(ctx context.Context, email, password string) (*User, error)
	GetProfile(ctx context.Context) (*User, error)
	GetCurrentUser(ctx context.Context) (*User, error)
}