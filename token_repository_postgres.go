package repository

import {
	"context"
	"database/sql"
	"github.com/lib/pq"
	"authservice/model"
}

type PostgresTokenRepository struct {
	db *sql.DB
}

func NewPostgresTokenRepository(db *sql.DB) TokenRepository {
	return &PostgresTokenRepository{db: db}
}

func (r *PostgresTokenRepository) CreateToken(ctx context.Context, clientID string, scope string) (*model.Token, error) {
	var token model.Token

	query := `
		INSERT INTO public.token (client_id, access_scope, access_token, expiration_time)
		VALUES ($1, $2, SUBSTR(UPPER(md5(random()::text)), 2, 22), current_timestamp + interval '2 hours')
		RETURNING client_id, access_token, access_scope, expiration_time
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		clientID,
		pq.Array([]string{scope}),
	).Scan(
		&token.ClientID,
		&token.AccessToken,
		pq.Array(&token.AccessScope),
		&token.ExpirationTime,
	)

	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *PostgresTokenRepository) GetToken(ctx context.Context, tokenValue string) (*model.Token, error) {
	var token model.Token

	query := `
		SELECT client_id, access_token, access_scope, expiration_time
		FROM public.token
		WHERE access_token = $1
	`

	err := r.db.QueryRowContext(ctx, query, tokenValue).Scan(
		&token.ClientID,
		&token.AccessToken,
		pq.Array(&token.AccessScope),
		&token.ExpirationTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // если токен не найден, возвращаем nil, это не ошибка
		}
		return nil, err // если другая ошибка — возвращаем ошибку
	}

	return &token, nil
}