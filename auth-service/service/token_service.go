package service

import (
	"context"
	"errors"
	"time"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"authservice/model"
	"authservice/repository"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid client credentials")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenNotFound      = errors.New("token not found")
)

var secretKey []byte

func InitSecret(secret string) {
	secretKey = []byte(secret)
}

type TokenService struct {
	tokenRepo repository.TokenRepository
	// userRepo  *repository.UserRepository
}

// NewTokenService creates a new TokenService
func NewTokenService(tokenRepo repository.TokenRepository) *TokenService {
	return &TokenService{
		tokenRepo: tokenRepo,
		// userRepo:  userRepo,
	}
}

// CreateToken creates a new access token
func (s *TokenService) CreateToken(ctx context.Context, email, clientSecret string) (*model.TokenResponse, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJson, _ := json.Marshal(header)
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJson)

	expirationTime := time.Now().Add(75 * time.Minute).Unix()
	payload := map[string]interface{}{
		"email":    email,
		"exp":      expirationTime,
	}
	payloadJson, _ := json.Marshal(payload)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJson)

	dataToSign := headerEncoded + "." + payloadEncoded

	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataToSign))
	signature := mac.Sum(nil)
	signatureEncoded := base64.RawURLEncoding.EncodeToString(signature)

	accessToken := dataToSign + "." + signatureEncoded

	refreshToken, err := generateRandomToken(32)
	if err != nil {
		return nil, err
	}
	refreshTokenExpiresAt := time.Now().Add(30 * 24 * time.Hour)
	// Create token
	err = s.tokenRepo.CreateToken(ctx, email, clientSecret, refreshToken, refreshTokenExpiresAt)
	if err != nil {
		return nil, err
	}
		
	// Prepare response
	response := &model.TokenResponse{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
		TokenType:   "Bearer",
	}
	
	return response, nil
}

func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// ValidateToken validates an access token
// func (s *TokenService) ValidateToken(ctx context.Context, tokenStr string) (*model.TokenValidationResponse, error) {
// 	// Get token from repository
// 	token, err := s.tokenRepo.GetToken(ctx, tokenStr)
// 	if err != nil {
// 		return nil, ErrTokenNotFound
// 	}
	
// 	// Check token expiration
// 	if time.Now().After(token.ExpirationTime) {
// 		return nil, ErrTokenExpired
// 	}
	
// 	// Prepare response
// 	response := &model.TokenValidationResponse{
// 		ClientID: token.ClientID,
// 		Scope:    token.AccessScope,
// 		Valid:    true,
// 	}
	
// 	return response, nil
// }