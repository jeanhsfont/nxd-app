package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"hubsystem/api"
)

// RequireAuth wraps a handler and ensures a user is logged in (via existing api auth).
// Sets user_id (email) in context for downstream use. Returns 401 if not authenticated.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := api.UserFromRequest(r)
		if user == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "faça login"})
			return
		}
		userID := user.Email
		if userID == "" {
			userID = fmt.Sprintf("id:%d", user.ID)
		}
		ctx := context.WithValue(r.Context(), ctxKeyUserID, userID)
		next(w, r.WithContext(ctx))
	}
}
