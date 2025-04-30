package repository

import (
	"context"
	"database/sql"
)

type PostgresBalanceRepository struct {
	db *sql.DB
}

func NewPostgresBalanceRepository(db *sql.DB) BalanceRepository {
	return &PostgresBalanceRepository{db: db}
}

func (r *PostgresBalanceRepository) UpdateBalanceByUUID(ctx context.Context, uuid string, amount int) error {
	query := `
		UPDATE users
		SET balance = balance + ?
		WHERE uuid = UUID_TO_BIN(REPLACE(?, '-', ''), 1)
	`
	_, err := r.db.ExecContext(ctx, query, amount, uuid)
	return err
}

func (r *PostgresBalanceRepository) UpdateBalanceByUUIDWithDrawal(ctx context.Context, uuid string, amount int) error {
	query := `
		UPDATE users
		SET balance = balance - ?
		WHERE uuid = UUID_TO_BIN(REPLACE(?, '-', ''), 1);
	`
	_, err := r.db.ExecContext(ctx, query, amount, uuid)
	return err
}

func (r *PostgresBalanceRepository) IsThereEnoughMoneyByUUID(ctx context.Context, uuid string, amount int) (bool, error) {
	query := `
		SELECT balance
		FROM users
		WHERE uuid = UUID_TO_BIN(REPLACE(?, '-', ''), 1)
	`
	var balance int
	err := r.db.QueryRowContext(ctx, query, uuid).Scan(&balance)
	if err != nil {
		return false, err
	}

	return balance >= amount, nil
}

func (r *PostgresBalanceRepository) TransactionCreate(ctx context.Context, uuid string, amount int, transactionType string) error {
    query := `
        INSERT INTO transactions (uuid, amount, type) 
        VALUES (UUID_TO_BIN(?, 1), ?, ?)
    `
    
    if transactionType != "deposit" && transactionType != "withdrawal" {
        return sql.ErrNoRows 
    }
    
    _, err := r.db.ExecContext(ctx, query, uuid, amount, transactionType)
    if err != nil {
        return err
    }
    
    return nil
}