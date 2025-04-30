package repository

import (
	"context"
)

type BalanceRepository interface {
	IsThereEnoughMoneyByUUID(ctx context.Context, uuid string, amount int) (bool, error)
	UpdateBalanceByUUID(ctx context.Context, uuid string, amount int) error
	UpdateBalanceByUUIDPAY(ctx context.Context, uuid string, amount int) error
	UpdateBalanceByUUIDWithDrawal(ctx context.Context, uuid string, amount int) error
	TransactionCreate(ctx context.Context, uuid string, amount int, transactionType string) error
}