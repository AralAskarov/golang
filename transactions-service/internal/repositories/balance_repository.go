package repositories

import (
	"context"
	"database/sql"
	"strings"
	"fmt"
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
	result, err := r.db.ExecContext(ctx, query, amount, cleaned)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		insertQuery := `
			INSERT INTO users (uuid, balance)
			VALUES ($1, $2)
		`
		_, err = r.db.ExecContext(ctx, insertQuery, cleaned, amount)
		if err != nil {
			return err
		}
	}
	
	return nil
}

func (r *PostgresBalanceRepository) UpdateBalanceByUUIDPAY(ctx context.Context, uuid string, amount int) error {
	query := `
		UPDATE users
		SET balance = balance + $1
		WHERE uuid = $2
	`
	result, err := r.db.ExecContext(ctx, query, amount, uuid)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		insertQuery := `
			INSERT INTO users (uuid, balance)
			VALUES ($1, $2)
		`
		_, err = r.db.ExecContext(ctx, insertQuery, uuid, amount)
		if err != nil {
			return err
		}
	}
	
	return nil
}


func (r *PostgresBalanceRepository) UpdateBalanceByUUIDWithDrawal(ctx context.Context, uuid string, amount int) error {
	upperUUID := strings.ToUpper(uuid)
	noHyphens := strings.ReplaceAll(upperUUID, "-", "")
	cleaned := "0x" + noHyphens
	query := `
		UPDATE users
		SET balance = balance - $1
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
	if err == sql.ErrNoRows {
		return false, fmt.Errorf("user with uuid %s not found", uuid)
	}
	if err != nil {
		return false, err
	}

	return balance >= amount, nil
}

func (r *PostgresBalanceRepository) TransactionCreate(ctx context.Context, uuid string, amount int, transactionType string) error {
    query := `
        INSERT INTO transactions (uuid, amount, type) 
        VALUES ($1, $2, $3)
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