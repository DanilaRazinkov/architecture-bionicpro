package services

import (
	"log"
	"auth/internal/config"
	"auth/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type KeycloakService struct {
	config *config.KeycloakConfig
	client *http.Client
}

func NewKeycloakService(cfg *config.KeycloakConfig) *KeycloakService {
	return &KeycloakService{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (k *KeycloakService) ExchangeCodeForToken(ctx context.Context, code, redirectURI string) (*models.TokenResponse, error) {
	return k.requestToken(ctx, url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {k.config.ClientID},
		"client_secret": {k.config.Secret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	})
}

func (k *KeycloakService) RefreshToken(ctx context.Context, refreshToken string) (*models.TokenResponse, error) {
	return k.requestToken(ctx, url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {k.config.ClientID},
		"client_secret": {k.config.Secret},
		"refresh_token": {refreshToken},
	})
}


func (k *KeycloakService) GetUserInfo(ctx context.Context, accessToken string) (*models.UserInfo, error) {
	userInfoURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", k.config.URL, k.config.Realm)

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user info request failed: %s, %d", string(body), resp.StatusCode)
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	log.Println("userInfo", userInfo)

	info := &models.UserInfo{
		ID:       getString(userInfo, "sub"),
		Username: getString(userInfo, "preferred_username"),
		Email:    getString(userInfo, "email"),
	}

	if userIDStr := getString(userInfo, "user_id"); userIDStr != "" {
		if userID, err := parseInt(userIDStr); err == nil {
			info.UserID = userID
		}
	}

	if crmUserIDStr := getString(userInfo, "crm_user_id"); crmUserIDStr != "" {
		if crmUserID, err := parseInt(crmUserIDStr); err == nil {
			info.CRMUserID = &crmUserID
		}
	}

	if roles, ok := userInfo["roles"].([]interface{}); ok {
		for _, role := range roles {
			if roleStr, ok := role.(string); ok {
				info.Roles = append(info.Roles, roleStr)
			}
		}
	}

	return info, nil
}

func (k *KeycloakService) requestToken(ctx context.Context, data url.Values) (*models.TokenResponse, error) {
	resp, err := k.postForm(ctx, k.buildURL("token"), data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token request failed: %s", string(body))
	}

	var tokenResp models.TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (k *KeycloakService) postForm(ctx context.Context, url string, data url.Values) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return k.client.Do(req)
}

func (k *KeycloakService) buildURL(endpoint string) string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/%s", k.config.URL, k.config.Realm, endpoint)
}

func (k *KeycloakService) mapUserInfo(data map[string]interface{}) *models.UserInfo {
	info := &models.UserInfo{
		ID:       getString(data, "sub"),
		Username: getString(data, "preferred_username"),
		Email:    getString(data, "email"),
		Roles:    extractRoles(data["roles"]),
	}

	if userID := parseToInt(getString(data, "user_id")); userID != 0 {
		info.UserID = userID
	}
	if crmUserID := parseToInt(getString(data, "crm_user_id")); crmUserID != 0 {
		info.CRMUserID = &crmUserID
	}

	return info
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func parseToInt(s string) int {
	if s == "" {
		return 0
	}
	val, _ := strconv.Atoi(s)
	return val
}

func extractRoles(roles interface{}) []string {
	var result []string
	if list, ok := roles.([]interface{}); ok {
		for _, role := range list {
			if str, ok := role.(string); ok {
				result = append(result, str)
			}
		}
	}
	return result
}


func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}