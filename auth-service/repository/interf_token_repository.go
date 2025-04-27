package repository

import (
	"context"
	"time"
)

type TokenRepository interface {
// 	ValidateCredentials(ctx context.Context, clientID, clientSecret, requestedScope string) (bool, error)
// 	GetUserScopes(ctx context.Context, clientID string) ([]string, error)
	CreateToken(ctx context.Context, email, clientSecret, refreshToken string, expiresAt time.Time) error
}