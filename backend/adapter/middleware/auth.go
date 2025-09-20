package middleware

import (
	"chat-app/backend/adapter/util"
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "userID"

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

func (m *AuthMiddleware) Validate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			util.RespondWithError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			util.RespondWithError(w, http.StatusUnauthorized, "Bearer token required")
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			util.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			util.RespondWithError(w, http.StatusUnauthorized, "Invalid token claims")
			return
		}

		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			util.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID in token")
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			util.RespondWithError(w, http.StatusUnauthorized, "Invalid user ID format in token")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

