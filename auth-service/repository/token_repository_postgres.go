package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PostgresTokenRepository struct {
	db *sql.DB
}

func NewPostgresTokenRepository(db *sql.DB) TokenRepository {
	return &PostgresTokenRepository{db: db}
}

func (r *PostgresTokenRepository) CreateToken(ctx context.Context, email, clientSecret, refreshToken string, expiresAt time.Time) error {
	// header := map[string]string{
	// 	"alg": "HS256",
	// 	"typ": "JWT",
	// }
	// headerJson, _ := json.Marshal(header)
	// headerEncoded := base64.RawURLEncoding.EncodeToString(headerJson)

	// expirationTime := time.Now().Add(15 * time.Minute).Unix()
	// payload := map[string]interface{}{
	// 	"username": username,
	// 	"email":    email,
	// 	"exp":      expirationTime,
	// }
	// payloadJson, _ := json.Marshal(payload)
	// payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJson)

	// dataToSign := headerEncoded + "." + payloadEncoded

	// mac := hmac.New(sha256.New, secretKey)
	// mac.Write([]byte(dataToSign))
	// signature := mac.Sum(nil)
	// signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	// accessToken := dataToSign + "." + signatureEncoded

	// refreshToken, err := generateRandomToken(32)
	// if err != nil {
	// 	return nil, err
	// }
	// refreshTokenExpiresAt := time.Now().Add(30 * 24 * time.Hour)
	query := `
		WITH existing_user AS (
			SELECT id
			FROM users
			WHERE email = $3 AND password = $4
		)
		INSERT INTO refresh_tokens (user_id, refresh_token, expires_at)
		SELECT id, $1, $2
		FROM existing_user
	`
	res, err := r.db.ExecContext(ctx, query, refreshToken, expiresAt, email, clientSecret)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user not found or credentials invalid")
	}
	return nil
}

// func generateRandomToken(length int) (string, error) {
// 	bytes := make([]byte, length)
// 	_, err := rand.Read(bytes)
// 	if err != nil {
// 		return "", err
// 	}
// 	return base64.RawURLEncoding.EncodeToString(bytes), nil
// }

// func (r *PostgresUserRepository) ValidateCredentials(ctx context.Context, clientID, clientSecret, requestedScope string) (bool, error) {
// 	var secret string
// 	var scopes []string

// 	query := `
// 		SELECT client_secret, scope 
// 		FROM public.user 
// 		WHERE client_id = $1
// 	`

// 	err := r.db.QueryRowContext(ctx, query, clientID).Scan(&secret, pq.Array(&scopes))
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return false, nil
// 		}
// 		return false, err
// 	}

// 	// Check client secret
// 	if secret != clientSecret {
// 		return false, nil
// 	}

// 	// Check if the requested scope is allowed
// 	validScope := false
// 	for _, scope := range scopes {
// 		if scope == requestedScope {
// 			validScope = true
// 			break
// 		}
// 	}

// 	return validScope, nil
// }

// func (r *PostgresUserRepository) GetUserScopes(ctx context.Context, clientID string) ([]string, error) {
// 	var scopes []string

// 	query := `
// 		SELECT scope 
// 		FROM public.user 
// 		WHERE client_id = $1
// 	`

// 	err := r.db.QueryRowContext(ctx, query, clientID).Scan(pq.Array(&scopes))
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return []string{}, nil
// 		}
// 		return nil, err
// 	}

// 	return scopes, nil
// }