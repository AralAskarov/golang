package model

type User struct {
	Username 	 string   `json:"username"`
	Email        string   `json:"email"`  
	ClientSecret string   `json:"client_secret"`
}

// type Token struct {
// 	AccessToken    string    `json:"access_token"`
// 	RefreshToken       string    `json:"refresh_token"`
// }

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType   string `json:"token_type"`
}

// type TokenValidationResponse struct {
// 	ClientID string   `json:"client_id"`
// 	Scope    []string `json:"scope"`
// 	Valid    bool     `json:"valid"`
// }