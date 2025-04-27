package repository

import (
	"context"
)

type BalanceRepository interface {
	IsThereEnoughMoneyByEmail(ctx context.Context, email string, amount int) (bool, error)
	UpdateBalanceByEmail(ctx context.Context, email string, amount int) error
	UpdateBalanceByEmailWithDrawal(ctx context.Context, email string, amount int) error
}