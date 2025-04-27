package repository

import {
	"context"
	"authservice/model"
}

type TokenRepository interface {
	CreateToken(ctx context.Context, clientID string, scope string) (*model.Token, error)
	GetToken(ctx context.Context, tokenValue string) (*model.Token, error)
}