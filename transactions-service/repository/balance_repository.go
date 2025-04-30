package repository

import (
	"context"
	"database/sql"
	"encoding/hex"
	"strings"
	"fmt"
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

func UUIDToBinReordered(uuid string) ([]byte, error) {
	uuid = strings.ReplaceAll(uuid, "-", "")
	if len(uuid) != 32 {
		return nil, fmt.Errorf("invalid UUID length")
	}
	raw, err := hex.DecodeString(uuid)
	if err != nil {
		return nil, err
	}
	if len(raw) != 16 {
		return nil, fmt.Errorf("decoded UUID must be 16 bytes")
	}

	reordered := make([]byte, 16)
	copy(reordered[0:4], raw[6:8]) 
	copy(reordered[2:4], raw[4:6])  
	copy(reordered[4:8], raw[0:4]) 
	copy(reordered[8:16], raw[8:16]) 

	return reordered, nil
}

func (r *PostgresBalanceRepository) UpdateBalanceByUUID(ctx context.Context, uuid string, amount int) error {
	hexUUID, err := UUIDToBinReordered(uuid)
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
	hexUUID, err := UUIDToBinReordered(uuid)
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
	hexUUID, err := UUIDToBinReordered(uuid)
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