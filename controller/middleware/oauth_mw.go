package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gonzabosio/res-manager/controller/handlers"
)

type TokenInfo struct {
	Audience  string `json:"audience"`
	UserID    string `json:"user_id"`
	ExpiresIn int    `json:"expires_in"`
	Scope     string `json:"scope"`
}

type ctxKey string

const accToken = ctxKey("access_token")

func OAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			handlers.WriteJSON(w, map[string]string{
				"message": "Authorization header is missing",
			}, http.StatusUnauthorized)
			return
		}
		var accessToken string
		_, err := fmt.Sscanf(authorizationHeader, "Bearer %s", &accessToken)
		if err != nil || accessToken == "" {
			handlers.WriteJSON(w, map[string]string{
				"message": "Invalid Authorization header format",
			}, http.StatusUnauthorized)
			return
		}

		client := http.Client{
			Timeout: 5 * time.Second,
		}
		resp, err := client.Get("https://www.googleapis.com/oauth2/v1/tokeninfo?access_token=" + accessToken)
		if err != nil {
			handlers.WriteJSON(w, map[string]string{
				"message": "Failed to verify token",
				"error":   err.Error(),
			}, http.StatusUnauthorized)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			handlers.WriteJSON(w, map[string]string{
				"message": "Token validation failed",
			}, http.StatusUnauthorized)
			return
		}

		var tokenInfo TokenInfo
		if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
			handlers.WriteJSON(w, map[string]string{
				"message": "Failed to decode token info",
				"error":   err.Error(),
			}, http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(r.Context(), accToken, tokenInfo)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
