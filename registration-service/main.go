// main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"authservice/config"
	"authservice/handlers"
	"authservice/repository"
	"authservice/services"
)

func main() {
	// Загрузка конфигурации из переменных окружения
	cfg := config.LoadConfig()
	
	// Инициализация логгера
	logger := initLogger(cfg)
	
	// Подключение к базе данных
	dbConn, err := repository.NewPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()
	
	// Настройка пула соединений
	dbConn.SetMaxOpenConns(cfg.DBMaxOpenConns)
	dbConn.SetMaxIdleConns(cfg.DBMaxIdleConns)
	dbConn.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	
	// Проверка соединения с базой данных
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := dbConn.PingContext(ctx); err != nil {
		logger.Fatalf("Database connection test failed: %v", err)
	}
	
	// Инициализация репозиториев
	tokenRepo := repository.NewTokenRepository(dbConn)
	userRepo := repository.NewUserRepository(dbConn)
	
	// Инициализация кэша с TTL для токенов
	tokenCache := services.NewTokenCache(cfg.TokenCacheTTL)
	
	// Инициализация сервисов
	tokenService := services.NewTokenService(tokenRepo, tokenCache, cfg.TokenTTL)
	authService := services.NewAuthService(userRepo, tokenService)
	
	// Инициализация обработчиков
	authHandler := handlers.NewAuthHandler(authService, logger)
	
	// Создание роутера с middleware
	router := http.NewServeMux()
	
	// Регистрация маршрутов с рейт-лимитером и middleware
	router.Handle("/token", handlers.WithMiddlewares(
		authHandler.TokenHandler,
		handlers.RateLimiter(cfg.RateLimit),
		handlers.RequestLogger(logger),
		handlers.RecoverPanic(logger),
	))
	router.Handle("/check", handlers.WithMiddlewares(
		authHandler.CheckTokenHandler,
		handlers.RateLimiter(cfg.RateLimit),
		handlers.RequestLogger(logger),
		handlers.RecoverPanic(logger),
	))
	
	// Создание сервера с таймаутами
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  cfg.ServerReadTimeout,
		WriteTimeout: cfg.ServerWriteTimeout,
		IdleTimeout:  cfg.ServerIdleTimeout,
	}
	
	// Запуск сервера в горутине
	go func() {
		logger.Printf("Starting server on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server error: %v", err)
		}
	}()
	
	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Println("Shutting down server...")
	
	// Graceful shutdown с таймаутом
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}
	
	logger.Println("Server exited properly")
}

func initLogger(cfg *config.Config) *log.Logger {
	// В реальном проекте здесь можно настроить логирование с ротацией файлов,
	// отправкой в Sentry, Loggly или другую систему сбора логов
	return log.New(os.Stdout, "AUTH-API: ", log.LstdFlags|log.Lshortfile)
}

// config/config.go
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port               string
	DatabaseURL        string
	DBMaxOpenConns     int
	DBMaxIdleConns     int
	DBConnMaxLifetime  time.Duration
	TokenTTL           time.Duration
	TokenCacheTTL      time.Duration
	RateLimit          int
	ServerReadTimeout  time.Duration
	ServerWriteTimeout time.Duration
	ServerIdleTimeout  time.Duration
}

func LoadConfig() *Config {
	return &Config{
		Port:               getEnv("PORT", "8083"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:pass123@localhost:5432/authf?sslmode=disable"),
		DBMaxOpenConns:     getEnvAsInt("DB_MAX_OPEN_CONNS", 500),
		DBMaxIdleConns:     getEnvAsInt("DB_MAX_IDLE_CONNS", 500),
		DBConnMaxLifetime:  getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		TokenTTL:           getEnvAsDuration("TOKEN_TTL", 2*time.Hour),
		TokenCacheTTL:      getEnvAsDuration("TOKEN_CACHE_TTL", 5*time.Minute),
		RateLimit:          getEnvAsInt("RATE_LIMIT", 100), // запросов в секунду
		ServerReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 5*time.Second),
		ServerWriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
		ServerIdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// repository/db.go
package repository

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgresConnection(connString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// repository/token_repository.go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type TokenRepository struct {
	db *sql.DB
}

type Token struct {
	ID            int64
	ClientID      string
	AccessScope   []string
	AccessToken   string
	ExpirationTime time.Time
	CreatedAt     time.Time
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (r *TokenRepository) CreateToken(ctx context.Context, clientID string, scope string, tokenTTL time.Duration) (*Token, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var token Token
	query := `
		INSERT INTO public.token (client_id, access_scope, access_token, expiration_time, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, client_id, access_scope, access_token, expiration_time, created_at;
	`

	// В реальном приложении использовать криптостойкий генератор
	accessToken, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}

	expirationTime := time.Now().Add(tokenTTL)
	scopes := []string{scope}
	now := time.Now()

	err = tx.QueryRowContext(
		ctx,
		query,
		clientID,
		pq.Array(scopes),
		accessToken,
		expirationTime,
		now,
	).Scan(
		&token.ID,
		&token.ClientID,
		pq.Array(&token.AccessScope),
		&token.AccessToken,
		&token.ExpirationTime,
		&token.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *TokenRepository) GetByAccessToken(ctx context.Context, accessToken string) (*Token, error) {
	query := `
		SELECT id, client_id, access_scope, access_token, expiration_time, created_at 
		FROM public.token 
		WHERE access_token = $1
	`

	var token Token
	err := r.db.QueryRowContext(ctx, query, accessToken).Scan(
		&token.ID,
		&token.ClientID,
		pq.Array(&token.AccessScope),
		&token.AccessToken,
		&token.ExpirationTime,
		&token.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

// Здесь будет функция для генерации безопасного токена через crypto/rand

// repository/user_repository.go
package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

type User struct {
	ID           int64
	ClientID     string
	ClientSecret string
	Scope        []string
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) ValidateCredentials(ctx context.Context, clientID, clientSecret, requestedScope string) (bool, error) {
	query := `
		SELECT u.id, u.client_id, u.client_secret, u.scope
		FROM public.user u
		WHERE u.client_id = $1
		LIMIT 1
	`

	var user User
	err := r.db.QueryRowContext(ctx, query, clientID).Scan(
		&user.ID,
		&user.ClientID,
		&user.ClientSecret,
		pq.Array(&user.Scope),
	)
	if err != nil {
		return false, err
	}

	// Проверка секрета
	if user.ClientSecret != clientSecret {
		return false, nil
	}

	// Проверка запрашиваемого scope
	scopeValid := false
	for _, s := range user.Scope {
		if s == requestedScope {
			scopeValid = true
			break
		}
	}

	return scopeValid, nil
}

// services/token_service.go
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"authservice/repository"
)

type TokenCache struct {
	cache      map[string]*repository.Token
	mu         sync.RWMutex
	defaultTTL time.Duration
}

func NewTokenCache(defaultTTL time.Duration) *TokenCache {
	cache := &TokenCache{
		cache:      make(map[string]*repository.Token),
		defaultTTL: defaultTTL,
	}
	
	// Запускаем горутину для очистки просроченных токенов
	go cache.periodicCleanup()
	
	return cache
}

func (c *TokenCache) Get(token string) (*repository.Token, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	cachedToken, exists := c.cache[token]
	if !exists {
		return nil, false
	}
	
	// Проверяем срок действия токена
	if time.Now().After(cachedToken.ExpirationTime) {
		return nil, false
	}
	
	return cachedToken, true
}

func (c *TokenCache) Set(token *repository.Token) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.cache[token.AccessToken] = token
}

func (c *TokenCache) periodicCleanup() {
	ticker := time.NewTicker(c.defaultTTL / 2)
	defer ticker.Stop()
	
	for range ticker.C {
		c.cleanup()
	}
}

func (c *TokenCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Now()
	for key, token := range c.cache {
		if now.After(token.ExpirationTime) {
			delete(c.cache, key)
		}
	}
}

type TokenService struct {
	repository repository.TokenRepository
	cache      *TokenCache
	tokenTTL   time.Duration
}

func NewTokenService(repo *repository.TokenRepository, cache *TokenCache, tokenTTL time.Duration) *TokenService {
	return &TokenService{
		repository: *repo,
		cache:      cache,
		tokenTTL:   tokenTTL,
	}
}

func (s *TokenService) CreateToken(ctx context.Context, clientID, scope string) (*repository.Token, error) {
	token, err := s.repository.CreateToken(ctx, clientID, scope, s.tokenTTL)
	if err != nil {
		return nil, err
	}
	
	// Сохраняем в кэш
	s.cache.Set(token)
	
	return token, nil
}

func (s *TokenService) ValidateToken(ctx context.Context, accessToken string) (*repository.Token, error) {
	// Сначала проверяем в кэше
	if cachedToken, found := s.cache.Get(accessToken); found {
		return cachedToken, nil
	}
	
	// Если нет в кэше, ищем в БД
	token, err := s.repository.GetByAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	
	// Проверяем срок действия
	if time.Now().After(token.ExpirationTime) {
		return nil, errors.New("token expired")
	}
	
	// Сохраняем в кэш
	s.cache.Set(token)
	
	return token, nil
}

// Функция для генерации криптостойкого токена
func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// services/auth_service.go
package services

import (
	"context"
	"errors"
	"time"

	"authservice/repository"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *TokenService
}

func NewAuthService(userRepo *repository.UserRepository, tokenRepo *TokenService) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type TokenValidationResponse struct {
	ClientID string   `json:"client_id"`
	Scope    []string `json:"scope"`
	Valid    bool     `json:"valid"`
}

func (s *AuthService) CreateAccessToken(ctx context.Context, clientID, clientSecret, scope string) (*TokenResponse, error) {
	// Проверяем учетные данные
	valid, err := s.userRepo.ValidateCredentials(ctx, clientID, clientSecret, scope)
	if err != nil {
		return nil, err
	}
	
	if !valid {
		return nil, errors.New("invalid credentials or scope")
	}
	
	// Создаем токен
	token, err := s.tokenRepo.CreateToken(ctx, clientID, scope)
	if err != nil {
		return nil, err
	}
	
	expiresIn := int(time.Until(token.ExpirationTime).Seconds())
	
	return &TokenResponse{
		AccessToken: token.AccessToken,
		ExpiresIn:   expiresIn,
		TokenType:   "Bearer",
	}, nil
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, accessToken string) (*TokenValidationResponse, error) {
	token, err := s.tokenRepo.ValidateToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	
	return &TokenValidationResponse{
		ClientID: token.ClientID,
		Scope:    token.AccessScope,
		Valid:    true,
	}, nil
}

// handlers/middleware.go
package handlers

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

// Middleware для объединения нескольких middleware
func WithMiddlewares(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}

// Middleware для логирования запросов
func RequestLogger(logger *log.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Обертка для записи статуса ответа
			wrappedWriter := NewResponseWriter(w)
			
			// Вызов следующего обработчика
			next(wrappedWriter, r)
			
			// Логирование после обработки
			duration := time.Since(start)
			logger.Printf(
				"[%s] %s %s %d %s",
				r.Method,
				r.RequestURI,
				r.RemoteAddr,
				wrappedWriter.Status(),
				duration,
			)
		}
	}
}

// ResponseWriter с возможностью отслеживания HTTP-статуса
type ResponseWriter struct {
	http.ResponseWriter
	status int
	written bool
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.written = true
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

func (rw *ResponseWriter) Status() int {
	return rw.status
}

// Middleware for rate limiting
func RateLimiter(requestsPerSecond int) func(http.HandlerFunc) http.HandlerFunc {
	type client struct {
		lastSeen time.Time
		count    int
	}
	
	var (
		clients = make(map[string]*client)
		mu      sync.Mutex
	)
	
	// Запуск горутины для очистки старых записей
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			
			mu.Lock()
			if _, exists := clients[ip]; !exists {
				clients[ip] = &client{
					lastSeen: time.Now(),
					count:    0,
				}
			}
			
			c := clients[ip]
			
			// Сброс счетчика через 1 секунду
			if time.Since(c.lastSeen) > time.Second {
				c.count = 0
				c.lastSeen = time.Now()
			}
			
			c.count++
			
			if c.count > requestsPerSecond {
				mu.Unlock()
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			
			mu.Unlock()
			
			next(w, r)
		}
	}
}

// Middleware для восстановления после паники
func RecoverPanic(logger *log.Logger) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Printf("Recovered from panic: %v", err)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			
			next(w, r)
		}
	}
}

// handlers/auth_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"authservice/services"
)

type AuthHandler struct {
	service *services.AuthService
	logger  *log.Logger
}

func NewAuthHandler(service *services.AuthService, logger *log.Logger) *AuthHandler {
	return &AuthHandler{
		service: service,
		logger:  logger,
	}
}

func (h *AuthHandler) TokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Установка таймаута для контекста
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	
	// Разбор формы
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	
	// Получение параметров
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	scope := r.FormValue("scope")
	
	// Валидация входных данных
	if clientID == "" || clientSecret == "" || scope == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}
	
	// Создание токена
	response, err := h.service.CreateAccessToken(ctx, clientID, clientSecret, scope)
	if err != nil {
		h.logger.Printf("Error creating token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Отправка ответа
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *AuthHandler) CheckTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Установка таймаута для контекста
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	
	// Извлечение токена из заголовка
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Authorization header is missing or format is incorrect", http.StatusUnauthorized)
		return
	}
	
	token := strings.TrimPrefix(authHeader, "Bearer ")
	
	// Проверка токена
	response, err := h.service.ValidateAccessToken(ctx, token)
	if err != nil {
		h.logger.Printf("Error validating token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	// Отправка ответа
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}