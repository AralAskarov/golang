package controller

import (
  "context"
  "encoding/json"
  "net/http"
  "time"
	
  "authservice/service"
)

// TokenController handles HTTP requests related to tokens
type TokenController struct {
	tokenService *service.TokenService
}

// NewTokenController creates a new TokenController
func NewTokenController(tokenService *service.TokenService) *TokenController {
	return &TokenController{
		tokenService: tokenService,
	}
}

// HandleTokenRequest handles token creation requests
func (c *TokenController) HandleTokenRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	
	if err := r.ParseForm(); err != nil {
		respondWithError(w, "Bad request", http.StatusBadRequest)
		return
	}
	
	email := r.FormValue("email")
	password := r.FormValue("password")
	
	if email == "" || password == "" {
		respondWithError(w, "Missing required parameters", http.StatusBadRequest)
		return
	}
	
	// Create token
	response, err := c.tokenService.CreateToken(ctx, email, password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			respondWithError(w, "Invalid credentials", http.StatusUnauthorized)
		default:
			respondWithError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	
	respondWithJSON(w, response, http.StatusOK)
}

// HandleTokenValidation handles token validation requests
// func (c *TokenController) HandleTokenValidation(w http.ResponseWriter, r *http.Request) {
// 	// Validate request method
// 	if r.Method != http.MethodGet {
// 		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}
	
// 	// Set timeout context
// 	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
// 	defer cancel()
	
// 	// Extract token from Authorization header
// 	authHeader := r.Header.Get("Authorization")
// 	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
// 		respondWithError(w, "Authorization header missing or invalid", http.StatusUnauthorized)
// 		return
// 	}
	
// 	token := strings.TrimPrefix(authHeader, "Bearer ")
	
// 	// Validate token
// 	response, err := c.tokenService.ValidateToken(ctx, token)
// 	if err != nil {
// 		switch err {
// 		case service.ErrTokenNotFound:
// 			respondWithError(w, "Token not found", http.StatusUnauthorized)
// 		case service.ErrTokenExpired:
// 			respondWithError(w, "Token expired", http.StatusUnauthorized)
// 		default:
// 			respondWithError(w, "Internal server error", http.StatusInternalServerError)
// 		}
// 		return
// 	}
	
// 	// Respond with validation result
// 	respondWithJSON(w, response, http.StatusOK)
// }

// Helper functions for HTTP responses
func respondWithError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}