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

func (r *PostgresBalanceRepository) UpdateBalanceByEmail(ctx context.Context, email string, amount int) error {
	query := `
		UPDATE users
		SET balance = balance + $1
		WHERE email = $2
	`
	_, err := r.db.ExecContext(ctx, query, amount, email)
	return err
}

func (r *PostgresBalanceRepository) UpdateBalanceByEmailWithDrawal(ctx context.Context, email string, amount int) error {
	query := `
		UPDATE users
		SET balance = balance - $1
		WHERE email = $2
	`
	_, err := r.db.ExecContext(ctx, query, amount, email)
	return err
}

func (r *PostgresBalanceRepository) IsThereEnoughMoneyByEmail(ctx context.Context, email string, amount int) (bool, error) {
	query := `
		SELECT balance
		FROM users
		WHERE email = $1
	`
	var balance int
	err := r.db.QueryRowContext(ctx, query, email).Scan(&balance)
	if err != nil {
		return false, err
	}

	return balance >= amount, nil
}