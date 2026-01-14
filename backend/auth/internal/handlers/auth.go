package handlers

import (
	"auth/internal/services"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	sessionService  *services.SessionService
	keycloakService *services.KeycloakService
}

func NewAuthHandler(sessionService *services.SessionService, keycloakService *services.KeycloakService) *AuthHandler {
	return &AuthHandler{
		sessionService:  sessionService,
		keycloakService: keycloakService,
	}
}

func (h *AuthHandler) GetAuthStatus(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"isAuthenticated": false})
		return
	}

	session, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"isAuthenticated": false})
		return
	}

	userInfo, err := h.keycloakService.GetUserInfo(c.Request.Context(), session.AccessToken)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"isAuthenticated": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"isAuthenticated": true,
		"name":            userInfo.Username,
		"username":        userInfo.Username,
		"userId":          userInfo.UserID,
		"crm_user_id":     userInfo.CRMUserID,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	authURL := fmt.Sprintf("http://localhost:8080/realms/reports-realm/protocol/openid-connect/auth?client_id=backend-auth&response_type=code&scope=openid backend-auth&redirect_uri=http://localhost:5001/api/auth/callback")
	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) HandleCallback(c *gin.Context) {
	code := c.Query("code")
	_ = c.Query("state")

	if code == "" {
		c.Redirect(http.StatusFound, "http://localhost:3000?error=no_code")
		return
	}

	redirectURI := "http://localhost:5001/api/auth/callback"

	existingSession, err := h.sessionService.GetSessionByCode(c.Request.Context(), code)
	if err == nil && existingSession != nil {
		c.SetCookie("session_id", existingSession.ID, int(24*time.Hour.Seconds()), "/", "localhost", false, true)
		c.Redirect(http.StatusFound, "http://localhost:3000")
		return
	}

	session, err := h.sessionService.CreateSession(c.Request.Context(), code, redirectURI)
	if err != nil {
		log.Println(err)
		c.Redirect(http.StatusFound, "http://localhost:3000?error=login_failed")
		return
	}

	c.SetCookie("session_id", session.ID, int(24*time.Hour.Seconds()), "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "http://localhost:3000")
}

func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No session found"})
		return
	}

	session, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          session.UserInfo.ID,
		"username":    session.UserInfo.Username,
		"email":       session.UserInfo.Email,
		"roles":       session.UserInfo.Roles,
		"userId":      session.UserInfo.UserID,
		"crm_user_id": session.UserInfo.CRMUserID,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No session found"})
		return
	}

	_, err = h.sessionService.RefreshSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token refreshed successfully"})
}

func (h *AuthHandler) GetReports(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No session found"})
		return
	}

	session, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	if session.UserInfo.CRMUserID == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Reports not available for this user (no CRM ID)"})
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", "http://reports-api:5003/api/v1/reports", nil)
	if err != nil {
		log.Println("failed to create request:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("failed to create request: %w", err)})
		return
	}

	req.Header.Set("Authorization", "Bearer "+session.AccessToken)
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", *session.UserInfo.CRMUserID))

	resp, err := client.Do(req)
	if err != nil {
		log.Println("failed to do:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get reports"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("failed to get reports:", err)
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to get reports"})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("failed to read response body:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	var report interface{}
	if err := json.Unmarshal(body, &report); err != nil {
		log.Println("failed to parse JSON:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	log.Println("report", report)

	c.JSON(http.StatusOK, report)
}

func (h *AuthHandler) GenerateReports(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No session found"})
		return
	}

	session, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequestWithContext(c.Request.Context(), "POST", "http://reports-api:5003/api/v1/reports/generate", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.Header.Set("Authorization", "Bearer "+session.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reports"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to generate reports"})
		return
	}

	var result map[string]interface{}
	if err := c.ShouldBindJSON(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	c.JSON(http.StatusOK, result)
}
