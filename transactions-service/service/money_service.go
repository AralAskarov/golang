package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"transervice/model"
	"transervice/repository"
)

const paymentURL = "https://arlan-api.azurewebsites.net/api/payment/pay"
const paymentURL2 = "https://arlan-api.azurewebsites.net/api/payment/addMoney"

var (
	ErrNotEnoughMoney        = errors.New("Not enough money")
	ErrInvalidCredentialsCard = errors.New("Invalid card credentials")
	ErrInvalidCredentials     = errors.New("Invalid user credentials")
	ErrPaymentFailed         = errors.New("Error")
)

var secretKey []byte

func InitSecret(secret string) {
	secretKey = []byte(secret)
	log.Println("Secret key initialized with length:", len(secret))
}

type BalanceService struct {
	balanceRepo repository.BalanceRepository
}

func NewBalanceService(balanceRepo repository.BalanceRepository) *BalanceService {
	log.Println("Creating new BalanceService")
	return &BalanceService{
		balanceRepo: balanceRepo,
	}
}

func (s *BalanceService) Replenishment(ctx context.Context, accessToken, cardNumber, cardOwner, cvv string, amount int) (*model.Response, error) {
	log.Println("Starting replenishment process")
	log.Printf("Access token: %s (length: %d)", maskToken(accessToken), len(accessToken))
	log.Printf("Card details: Number: %s, Owner: %s, CVV: %s", maskCardNumber(cardNumber), cardOwner, "***")
	log.Printf("Amount to replenish: %d", amount)
	
	// Validate token format
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		log.Println("ERROR: Invalid token format - expected 3 parts, got:", len(parts))
		return nil, ErrInvalidCredentials
	}
	
	headerEncoded := parts[0]
	payloadEncoded := parts[1]
	signatureEncoded := parts[2]
	
	log.Printf("Token parts lengths - Header: %d, Payload: %d, Signature: %d", 
		len(headerEncoded), len(payloadEncoded), len(signatureEncoded))
	
	// Verify token signature
	dataToSign := headerEncoded + "." + payloadEncoded
	
	if len(secretKey) == 0 {
		log.Println("ERROR: Secret key not initialized")
		return nil, errors.New("secret key not initialized")
	}
	
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataToSign))
	expectedSignature := mac.Sum(nil)
	expectedSignatureEncoded := base64.RawURLEncoding.EncodeToString(expectedSignature)
	
	log.Printf("Token signature validation - Expected: %s, Received: %s", 
		expectedSignatureEncoded[:5]+"...", signatureEncoded[:5]+"...")
	
	if signatureEncoded != expectedSignatureEncoded {
		log.Println("ERROR: Invalid token signature")
		return nil, ErrInvalidCredentials
	}
	
	log.Println("Token signature validated successfully")
	
	// Decode token payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		log.Printf("ERROR: Failed to decode token payload: %v", err)
		return nil, err
	}
	
	// Parse token payload
	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Printf("ERROR: Failed to unmarshal token payload: %v", err)
		return nil, err
	}
	
	// Extract email from token
	email, ok := payload["email"].(string)
	if !ok {
		log.Println("ERROR: Invalid token payload - no email found")
		return nil, errors.New("invalid token payload: no email")
	}
	
	log.Printf("User email extracted from token: %s", email)
	
	// Prepare payment request
	reqBody, err := json.Marshal(map[string]interface{}{
		"cardNumber": cardNumber,
		"cardOwnerName": cardOwner,
		"cvv": cvv,
		"paymentAmount": amount,
	})
	if err != nil {
		log.Printf("ERROR: Failed to marshal payment request: %v", err)
		return nil, err
	}
	
	log.Printf("Payment request prepared: %s", maskJSON(string(reqBody)))
	
	// Send payment request
	log.Printf("Sending payment request to URL: %s", paymentURL)
	resp, err := http.Post(paymentURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("ERROR: Failed to send payment request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	
	// Read response body
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read response body: %v", err)
		return nil, err
	}
	
	bodyStr := strings.TrimSpace(string(bodyBytes))
	log.Printf("Payment API response: Status: %d, Body: %s", resp.StatusCode, bodyStr)
	
	// Handle response
	switch {
	case resp.StatusCode == http.StatusOK:
		log.Println("Payment successful, updating balance in database")
		err := s.balanceRepo.UpdateBalanceByEmail(ctx, email, amount)
		if err != nil {
			log.Printf("ERROR: Failed to update balance in database: %v", err)
			return nil, err
		}
		log.Println("Balance successfully updated")
		return &model.Response{Message: "Balance successfully replenished"}, nil
		
	case resp.StatusCode == http.StatusBadRequest && strings.Contains(bodyStr, "Invalid Credentials"):
		log.Println("ERROR: Invalid card credentials")
		return nil, ErrInvalidCredentialsCard
		
	case resp.StatusCode == http.StatusBadRequest && strings.Contains(bodyStr, "Not enough money"):
		log.Println("ERROR: Not enough money on card")
		return nil, ErrNotEnoughMoney
		
	default:
		log.Printf("ERROR: Unexpected response - Status: %d, Body: %s", resp.StatusCode, bodyStr)
		return nil, fmt.Errorf("%w: %s", ErrPaymentFailed, bodyStr)
	}
}

// Helper functions for logging
func maskToken(token string) string {
	if len(token) <= 10 {
		return "***"
	}
	return token[:5] + "..." + token[len(token)-5:]
}

func maskCardNumber(number string) string {
	if len(number) <= 8 {
		return "****"
	}
	return number[:4] + "..." + number[len(number)-4:]
}

func maskJSON(jsonStr string) string {
	// Simple masking for logging sensitive JSON data
	jsonStr = strings.Replace(jsonStr, "\"cardNumber\":\""+`[^"]+`+"\"", "\"cardNumber\":\"****\"", -1)
	jsonStr = strings.Replace(jsonStr, "\"cvv\":\""+`[^"]+`+"\"", "\"cvv\":\"***\"", -1)
	return jsonStr
}

func (s *BalanceService) Withdrawal(ctx context.Context, accessToken, cardNumber string, amount int) (*model.Response, error) {
	log.Printf("Starting withdrawal process. Card: %s, Amount: %d", cardNumber, amount)
	
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		log.Printf("Error: Invalid token format. Expected 3 parts, got %d", len(parts))
		return nil, ErrInvalidCredentials
	}
	log.Printf("Token format validated successfully")

	headerEncoded := parts[0]
	payloadEncoded := parts[1]
	signatureEncoded := parts[2]
	log.Printf("Token parts extracted: header, payload, and signature")

	dataToSign := headerEncoded + "." + payloadEncoded
	mac := hmac.New(sha256.New, secretKey)
	mac.Write([]byte(dataToSign))
	expectedSignature := mac.Sum(nil)
	expectedSignatureEncoded := base64.RawURLEncoding.EncodeToString(expectedSignature)
	log.Printf("Signature verification: calculating expected signature")

	if signatureEncoded != expectedSignatureEncoded {
		log.Printf("Error: Signature verification failed. Token is invalid")
		return nil, ErrInvalidCredentials
	}
	log.Printf("Signature verified successfully")

	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		log.Printf("Error decoding payload: %v", err)
		return nil, err
	}
	log.Printf("Payload decoded successfully")

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		log.Printf("Error unmarshaling payload: %v", err)
		return nil, err
	}
	log.Printf("Payload unmarshaled successfully: %+v", payload)

	email, ok := payload["email"].(string)
	if !ok {
		log.Printf("Error: email field not found in payload or not a string")
		return nil, errors.New("invalid token payload: no email")
	}
	log.Printf("Email extracted from payload: %s", email)

	ok, err = s.balanceRepo.IsThereEnoughMoneyByEmail(ctx, email, amount)
	if err != nil {
		log.Printf("Error checking balance: %v", err)
		return nil, err
	}
	if !ok {
		log.Printf("Error: Not enough money in balance for withdrawal")
		return nil, ErrNotEnoughMoney
	}
	log.Printf("Balance check passed: sufficient funds available")

	reqBody, _ := json.Marshal(map[string]interface{}{
		"cardNumber":    cardNumber,
		"paymentAmount": amount,
	})
	log.Printf("Payment request prepared: %s", string(reqBody))
	
	log.Printf("Sending payment request to %s", paymentURL2)
	resp, err := http.Post(paymentURL2, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("Error sending payment request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("Payment request sent. Status code: %d", resp.StatusCode)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}
	bodyStr := strings.TrimSpace(string(bodyBytes))
	log.Printf("Payment response received: [%d] %s", resp.StatusCode, bodyStr)

	switch {
	case resp.StatusCode == http.StatusOK:
		log.Printf("Payment successful, updating balance")
		err := s.balanceRepo.UpdateBalanceByEmailWithDrawal(ctx, email, amount)
		if err != nil {
			log.Printf("Error updating balance: %v", err)
			return nil, err
		}
		log.Printf("Balance successfully updated")
		return &model.Response{Message: "Balance successfully replenished"}, nil

	case resp.StatusCode == http.StatusBadRequest && strings.Contains(bodyStr, "Invalid Credentials"):
		log.Printf("Error: Invalid card credentials")
		return nil, ErrInvalidCredentialsCard

	case resp.StatusCode == http.StatusBadRequest && strings.Contains(bodyStr, "Not enough money"):
		log.Printf("Error: Not enough money on card")
		return nil, ErrNotEnoughMoney

	default:
		log.Printf("Error: Payment failed with unexpected response: [%d] %s", resp.StatusCode, bodyStr)
		return nil, ErrPaymentFailed
	}
}