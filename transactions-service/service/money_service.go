package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"transervice/model"
	"transervice/repository"
	"io"
)

const paymentURL = "https://arlan-api.azurewebsites.net/api/payment/pay"
const paymentURL2 = "https://arlan-api.azurewebsites.net/api/payment/addMoney"
const profileURL = "http://golang.medhelper.xyz/profile"

var (
	ErrNotEnoughMoney        = errors.New("Not enough money")
	ErrInvalidCredentialsCard = errors.New("Invalid card credentials")
	ErrInvalidCredentials     = errors.New("Invalid user credentials")
	ErrPaymentFailed         = errors.New("Error")
)

// var secretKey []byte

// func InitSecret(secret string) {
// 	secretKey = []byte(secret)
// 	log.Println("Secret key initialized with length:", len(secret))
// }

type BalanceService struct {
	balanceRepo repository.BalanceRepository
}

func NewBalanceService(balanceRepo repository.BalanceRepository) *BalanceService {
	log.Println("Creating new BalanceService")
	return &BalanceService{
		balanceRepo: balanceRepo,
	}
}

func GetUserUUID(token string) (string, error) {
	req, err := http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("non-200 response: %s", string(body))
	}

	var result struct {
		UUID string `json:"uuid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.UUID, nil
}


func (s *BalanceService) Replenishment(ctx context.Context, accessToken, cardNumber, cardOwner, cvv string, amount int) (*model.Response, error) {
	uuid, err := GetUserUUID(accessToken)
	if err != nil {
		log.Printf("ERROR: Failed to get user UUID, invalid token: %v", err)
		return nil, ErrInvalidCredentials
	}
	
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

	resp, err := http.Post(paymentURL, "application/json", bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("ERROR: Failed to send payment request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read response body: %v", err)
		return nil, err
	}
	
	bodyStr := strings.TrimSpace(string(bodyBytes))
	log.Printf("Payment API response: Status: %d, Body: %s", resp.StatusCode, bodyStr)
	
	switch {
	case resp.StatusCode == http.StatusOK:
		log.Println("Payment successful, updating balance in database")
		err := s.balanceRepo.UpdateBalanceByUUID(ctx, uuid, amount)
		if err != nil {
			log.Printf("ERROR: Failed to update balance in database: %v", err)
			return nil, err
		}
		log.Println("Balance successfully updated")

		err = s.balanceRepo.TransactionCreate(ctx, uuid, amount, "deposit")
        if err != nil {
            log.Printf("ERROR: Failed to create transaction record: %v", err)
            return nil, err
        }
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
	uuid, err := GetUserUUID(accessToken)
	if err != nil {
		log.Printf("ERROR: Failed to get user UUID, invalid token: %v", err)
		return nil, ErrInvalidCredentials
	}
	

	ok, err := s.balanceRepo.IsThereEnoughMoneyByUUID(ctx, uuid, amount)
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
		err := s.balanceRepo.UpdateBalanceByUUIDWithDrawal(ctx, uuid, amount)
		if err != nil {
			log.Printf("Error updating balance: %v", err)
			return nil, err
		}
		log.Printf("Balance successfully updated")
		err = s.balanceRepo.TransactionCreate(ctx, uuid, amount, "withdrawal")
        if err != nil {
            log.Printf("ERROR: Failed to create transaction record: %v", err)
            return nil, err
        }
		return &model.Response{Message: "Balance successfully replenished to card back"}, nil

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