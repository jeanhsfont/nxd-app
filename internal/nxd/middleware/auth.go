package middleware

import (
	"context"
	"net/http"
	"strings"

	"hubsystem/internal/nxd/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type contextKey string

const (
	ctxKeyUserID    contextKey = "nxd_user_id"
	ctxKeyFactoryID contextKey = "nxd_factory_id"
	ctxKeyRole      contextKey = "nxd_role"
)

// UserIDFromRequest returns the NXD user_id (email) from context.
func UserIDFromRequest(r *http.Request) string {
	if v := r.Context().Value(ctxKeyUserID); v != nil {
		return v.(string)
	}
	return ""
}

// FactoryIDFromContext returns the factory UUID from context (set by RequireFactory).
func FactoryIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	v := ctx.Value(ctxKeyFactoryID)
	if v == nil {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// RoleFromContext returns the role from context.
func RoleFromContext(ctx context.Context) string {
	v := ctx.Value(ctxKeyRole)
	if v == nil {
		return ""
	}
	return v.(string)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return config.JWTKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyUserID, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
