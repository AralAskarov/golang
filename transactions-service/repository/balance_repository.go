package repository

import (
	"context"
	"database/sql"
	"strings"
)

type PostgresBalanceRepository struct {
	db *sql.DB
}

func NewPostgresBalanceRepository(db *sql.DB) BalanceRepository {
	return &PostgresBalanceRepository{db: db}
}

func (r *PostgresBalanceRepository) UpdateBalanceByUUID(ctx context.Context, uuid string, amount int) error {
	upperUUID := strings.ToUpper(uuid)
	noHyphens := strings.ReplaceAll(upperUUID, "-", "")
	cleaned := "0x" + noHyphens
	query := `
		UPDATE users
		SET balance = balance + $1
		WHERE uuid = $2
	`
	_, err := r.db.ExecContext(ctx, query, amount, cleaned)
	return err
}

func (r *PostgresBalanceRepository) UpdateBalanceByUUIDWithDrawal(ctx context.Context, uuid string, amount int) error {
	upperUUID := strings.ToUpper(uuid)
	noHyphens := strings.ReplaceAll(upperUUID, "-", "")
	cleaned := "0x" + noHyphens
	query := `
		UPDATE users
		SET balance = balance + $1
		WHERE uuid = $2
	`
	_, err := r.db.ExecContext(ctx, query, amount, cleaned)
	return err
}

func (r *PostgresBalanceRepository) IsThereEnoughMoneyByUUID(ctx context.Context, uuid string, amount int) (bool, error) {
	upperUUID := strings.ToUpper(uuid)
	noHyphens := strings.ReplaceAll(upperUUID, "-", "")
	cleaned := "0x" + noHyphens
	query := `
		SELECT balance
		FROM users
		WHERE uuid = $1
	`
	var balance int
	err := r.db.QueryRowContext(ctx, query, cleaned).Scan(&balance)
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