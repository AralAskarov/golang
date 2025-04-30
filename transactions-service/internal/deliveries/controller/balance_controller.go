package controller

import (
  "context"
  "encoding/json"
  "net/http"
  "time"
  "strconv"
  "errors"
  "log"
  "transervice/service"
)

type BalanceController struct {
	balanceService *service.BalanceService
}
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrInsufficientFunds = errors.New("insufficient funds")
)
func NewBalanceController(balanceService *service.BalanceService) *BalanceController {
	return &BalanceController{
		balanceService: balanceService,
	}
}

func (c *BalanceController) ReplenishmentRequest(w http.ResponseWriter, r *http.Request) {
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
	
	accesstoken := r.FormValue("access_token")
	cardnumber := r.FormValue("card_number")
	cardowner := r.FormValue("card_owner")
	cvv := r.FormValue("cvv")
	money := r.FormValue("money")

	if accesstoken == "" || cardnumber == "" || cardowner == "" || cvv == "" || money == "" {
		respondWithError(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	moneyInt, err := strconv.Atoi(money)
	if err != nil {
		respondWithError(w, "Invalid money value", http.StatusBadRequest)
		return
	}
	
	response, err := c.balanceService.Replenishment(ctx, accesstoken, cardnumber, cardowner, cvv, moneyInt)
	if err != nil {
		log.Printf("ReplenishmentRequest error: %v", err)
		switch err {
		case service.ErrNotEnoughMoney:
			respondWithError(w, "Not enough money on the card", http.StatusUnauthorized)
		case service.ErrInvalidCredentials:
			respondWithError(w, "Invalid card credentials", http.StatusUnauthorized)
		case service.ErrInvalidCredentials:
			respondWithError(w, "Invalid user credentials", http.StatusUnauthorized)
		default:
			respondWithError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	
	respondWithJSON(w, response, http.StatusOK)
}

func (c *BalanceController) WithdrawalRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 16*time.Second)
	defer cancel()
	
	if err := r.ParseForm(); err != nil {
		respondWithError(w, "Bad request", http.StatusBadRequest)
		return
	}
	
	accesstoken := r.FormValue("access_token")
	cardnumber := r.FormValue("card_number")
	money := r.FormValue("money")

	if accesstoken == "" || cardnumber == "" || money == "" {
		respondWithError(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	moneyInt, err := strconv.Atoi(money)
	if err != nil {
		respondWithError(w, "Invalid money value", http.StatusBadRequest)
		return
	}
	
	response, err := c.balanceService.Withdrawal(ctx, accesstoken, cardnumber, moneyInt)
	if err != nil {
		log.Printf("ReplenishmentRequest error: %v", err)
		switch err {
		case service.ErrNotEnoughMoney:
			respondWithError(w, "Not enough money on the card", http.StatusUnauthorized)
		case service.ErrInvalidCredentials:
			respondWithError(w, "Invalid card credentials", http.StatusUnauthorized)
		default:
			respondWithError(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	
	respondWithJSON(w, response, http.StatusOK)
}


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

func (c *BalanceController) UpdateBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithErrorr(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	ctx, cancel := context.WithTimeout(r.Context(), 16*time.Second)
	defer cancel()
	
	var payoutRequest struct {
		UserID string  `json:"userId"`
		Amount int `json:"amount"`
	}
	
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payoutRequest); err != nil {
		respondWithErrorr(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	
	err := c.balanceService.ProcessUserPayout(ctx, payoutRequest.UserID, payoutRequest.Amount)
	if err != nil {
		log.Printf("ProcessUserPayout error: %v", err)
		switch {
		case errors.Is(err, ErrUserNotFound):
			respondWithErrorr(w, "User not found", http.StatusNotFound)
		case errors.Is(err, ErrInsufficientFunds):
			respondWithErrorr(w, "Insufficient system funds", http.StatusConflict)
		default:
			respondWithErrorr(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	
	response := struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Success: true,
		Message: "Payout processed successfully",
	}
	
	respondWithJSONN(w, response, http.StatusOK)
}

func respondWithErrorr(w http.ResponseWriter, message string, statusCode int) {
	response := struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}{
		Success: false,
		Error:   message,
	}
	respondWithJSON(w, response, statusCode)
}

func respondWithJSONN(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}