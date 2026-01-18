package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Yandex API endpoints
const (
	YandexAuthURL     = "https://oauth.yandex.ru/authorize"
	YandexTokenURL    = "https://oauth.yandex.ru/token"
	YandexUserInfoURL = "https://login.yandex.ru/info"
)

// App структура для хранения состояния приложения
type App struct {
	challengeToNonce map[string]string
	mu               sync.RWMutex
}

// TokenResponse ответ от Yandex OAuth
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// UserInfo информация о пользователе Yandex
type UserInfo struct {
	ID           string `json:"id"`
	Login        string `json:"login"`
	ClientID     string `json:"client_id"`
	DisplayName  string `json:"display_name"`
	RealName     string `json:"real_name"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Sex          string `json:"sex"`
	DefaultEmail string `json:"default_email"`
	Email        string `json:"email,omitempty"`
	Sub          string `json:"sub,omitempty"`
}

// JWTHeader заголовок JWT токена
type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// JWTPayload payload JWT токена
type JWTPayload struct {
	Iss string `json:"iss"`
	Sub string `json:"sub"`
	Aud string `json:"aud"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
	Nonce string `json:"nonce,omitempty"`
}

func main() {
	// Инициализируем приложение
	app := &App{
		challengeToNonce: make(map[string]string),
	}

	// Настраиваем Gin
	router := gin.Default()

	// Middleware для логирования
	router.Use(loggingMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Authorize endpoint
	router.GET("/authorize", app.proxyAuthorize)

	// Token endpoint
	router.POST("/token", app.proxyToken)

	// User info endpoint
	router.GET("/info", app.proxyUserInfo)

	// Запускаем сервер
	log.Println("Starting Keycloak Yandex Proxy on :8000")
	if err := router.Run(":5000"); err != nil {
		log.Fatal(err)
	}
}

// loggingMiddleware middleware для логирования запросов
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Продолжаем обработку
		c.Next()

		// Логируем после обработки
		latency := time.Since(start)
		log.Printf("[%s] %s %s - %v",
			c.Request.Method,
			c.Request.URL.Path,
			c.ClientIP(),
			latency,
		)
	}
}

// proxyAuthorize проксирует запрос к Yandex authorize endpoint
func (app *App) proxyAuthorize(c *gin.Context) {
	// Получаем query параметры
	params := c.Request.URL.Query()

	// Логируем оригинальный запрос
	log.Printf("Original authorize request: %v", params)

	// Убираем openid из scope если он есть
	if scope := params.Get("scope"); scope != "" {
		originalScope := scope
		// Убираем openid из scope
		newScope := removeOpenIDFromScope(scope)
		params.Set("scope", newScope)
		log.Printf("Modified scope: %s -> %s", originalScope, newScope)
	}

	// Сохраняем связь между code_challenge и nonce
	codeChallenge := params.Get("code_challenge")
	nonce := params.Get("nonce")
	if codeChallenge != "" && nonce != "" {
		app.mu.Lock()
		app.challengeToNonce[codeChallenge] = nonce
		app.mu.Unlock()
		log.Printf("Saved nonce for challenge: %s...", codeChallenge[:min(8, len(codeChallenge))])
	}

	// Строим URL для редиректа
	redirectURL := fmt.Sprintf("%s?%s", YandexAuthURL, params.Encode())
	log.Printf("Redirecting to: %s", redirectURL)

	// Редирект на Yandex
	c.Redirect(http.StatusFound, redirectURL)
}

// proxyToken проксирует запрос к Yandex token endpoint
func (app *App) proxyToken(c *gin.Context) {
	// Парсим form data из тела запроса
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	formData := c.Request.PostForm
	log.Printf("Token request: %v", formData)

	// Создаем HTTP клиент
	client := &http.Client{Timeout: 30 * time.Second}

	// Проксируем запрос к Yandex
	resp, err := client.PostForm(YandexTokenURL, formData)
	if err != nil {
		log.Printf("Error proxying to Yandex token endpoint: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact Yandex"})
		return
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	log.Printf("Yandex token response: %d - %s", resp.StatusCode, string(body))

	if resp.StatusCode == http.StatusOK {
		// Парсим токен
		var tokenData TokenResponse
		if err := json.Unmarshal(body, &tokenData); err != nil {
			log.Printf("Error parsing token response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse token"})
			return
		}

		log.Printf("Token data: %+v", tokenData)

		// Добавляем id_token если его нет
		if tokenData.IDToken == "" {
			// Получаем userinfo для получения sub
			userSub := "yandex_user"
			if tokenData.AccessToken != "" {
				if userInfo, err := app.getUserInfo(tokenData.AccessToken); err == nil && userInfo.ID != "" {
					userSub = userInfo.ID
				}
			}

			// Создаем fake JWT
			clientID := formData.Get("client_id")
			if clientID == "" {
				clientID = "24284f6232cc4802a10ae9e73360fdfa"
			}

			// Восстанавливаем nonce из code_verifier
			var nonce string
			codeVerifier := formData.Get("code_verifier")
			if codeVerifier != "" {
				challenge := computeCodeChallenge(codeVerifier)
				app.mu.RLock()
				storedNonce := app.challengeToNonce[challenge]
				app.mu.RUnlock()
				if storedNonce != "" {
					nonce = storedNonce
					log.Printf("Added nonce to fake id_token via PKCE mapping")
				} else {
					log.Printf("No stored nonce for computed challenge; proceeding without nonce")
				}
			}

			// Создаем JWT
			fakeJWT := createFakeJWT(userSub, clientID, nonce)
			tokenData.IDToken = fakeJWT
			log.Printf("Added fake id_token: %s", fakeJWT)
		}

		// Устанавливаем заголовки и возвращаем ответ
		c.Header("Content-Type", "application/json")
		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.JSON(resp.StatusCode, tokenData)
	} else {
		// Возвращаем ошибку как есть
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
}

// proxyUserInfo проксирует запрос к Yandex userinfo endpoint
func (app *App) proxyUserInfo(c *gin.Context) {
	// Создаем HTTP клиент
	client := &http.Client{Timeout: 30 * time.Second}

	// Создаем запрос к Yandex
	req, err := http.NewRequest("GET", YandexUserInfoURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Копируем заголовки, кроме Host
	for k, v := range c.Request.Header {
		if strings.ToLower(k) != "host" {
			req.Header[k] = v
		}
	}

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error proxying to Yandex userinfo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact Yandex"})
		return
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	log.Printf("Yandex userinfo response: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusOK {
		// Парсим userinfo
		var userData UserInfo
		if err := json.Unmarshal(body, &userData); err != nil {
			log.Printf("Error parsing userinfo: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse userinfo"})
			return
		}

		log.Printf("Original userinfo: %+v", userData)

		// Добавляем поле 'sub' для OIDC совместимости
		if userData.ID != "" && userData.Sub == "" {
			userData.Sub = userData.ID
			log.Printf("Added sub field: %s", userData.Sub)
		}

		// Убеждаемся что есть email
		if userData.DefaultEmail != "" && userData.Email == "" {
			userData.Email = userData.DefaultEmail
		}

		// Устанавливаем заголовки и возвращаем ответ
		c.Header("Content-Type", "application/json")
		c.Header("Cache-Control", "no-store")
		c.JSON(resp.StatusCode, userData)
	} else {
		// Возвращаем ошибку как есть
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
}

// getUserInfo получает информацию о пользователе
func (app *App) getUserInfo(accessToken string) (*UserInfo, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", YandexUserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", accessToken))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("yandex returned status: %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// removeOpenIDFromScope удаляет openid из scope
func removeOpenIDFromScope(scope string) string {
	// Удаляем слово openid и лишние пробелы
	re := regexp.MustCompile(`\bopenid\b\s*`)
	newScope := re.ReplaceAllString(scope, "")

	// Убираем лишние пробелы
	reSpaces := regexp.MustCompile(`\s+`)
	newScope = reSpaces.ReplaceAllString(newScope, " ")

	return strings.TrimSpace(newScope)
}

// computeCodeChallenge вычисляет code_challenge из code_verifier
func computeCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	challenge := base64.URLEncoding.EncodeToString(hash[:])
	return strings.TrimRight(challenge, "=")
}

// createFakeJWT создает fake JWT токен без подписи
func createFakeJWT(userSub, clientID, nonce string) string {
	header := JWTHeader{
		Alg: "none",
		Typ: "JWT",
	}

	payload := JWTPayload{
		Iss: "https://oauth.yandex.ru",
		Sub: userSub,
		Aud: clientID,
		Exp: 9999999999,
		Iat: 1726517000,
	}

	if nonce != "" {
		payload.Nonce = nonce
	}

	// Кодируем header и payload
	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.URLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.URLEncoding.EncodeToString(payloadJSON)

	// Удаляем padding
	headerB64 = strings.TrimRight(headerB64, "=")
	payloadB64 = strings.TrimRight(payloadB64, "=")

	return fmt.Sprintf("%s.%s.", headerB64, payloadB64)
}

// min возвращает минимальное из двух чисел
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
