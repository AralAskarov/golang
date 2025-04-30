package repository

import (
	"context"
	"database/sql"
	"encoding/hex"
	"strings"
)

type PostgresBalanceRepository struct {
	db *sql.DB
}

func NewPostgresBalanceRepository(db *sql.DB) BalanceRepository {
	return &PostgresBalanceRepository{db: db}
}

func UUIDToHexBin(uuid string) (string, error) {
	uuid = strings.ReplaceAll(uuid, "-", "")
	bytes, err := hex.DecodeString(uuid)
	if err != nil {
		return "", err
	}
	return "0x" + strings.ToUpper(hex.EncodeToString(bytes)), nil
}

func (r *PostgresBalanceRepository) UpdateBalanceByUUID(ctx context.Context, uuid string, amount int) error {
	hexUUID, err := UUIDToHexBin(uuid)
	if err != nil {
		return err
	}
	query := `
		UPDATE users
		SET balance = balance + ?
		WHERE uuid = ?
	`
	_, err = r.db.ExecContext(ctx, query, amount, hexUUID)
	return err
}

func (r *PostgresBalanceRepository) UpdateBalanceByUUIDWithDrawal(ctx context.Context, uuid string, amount int) error {
	hexUUID, err := UUIDToHexBin(uuid)
	if err != nil {
		return err
	}
	query := `
		UPDATE users
		SET balance = balance - ?
		WHERE uuid = ?
	`
	_, err = r.db.ExecContext(ctx, query, amount, hexUUID)
	return err
}

func (r *PostgresBalanceRepository) IsThereEnoughMoneyByUUID(ctx context.Context, uuid string, amount int) (bool, error) {
	hexUUID, err := UUIDToHexBin(uuid)
	if err != nil {
		return false, err
	}
	query := `
		SELECT balance
		FROM users
		WHERE uuid = ?
	`
	var balance int
	err = r.db.QueryRowContext(ctx, query, hexUUID).Scan(&balance)
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