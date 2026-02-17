package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"hubsystem/api"
	"hubsystem/internal/nxd/store"

	"github.com/google/uuid"
)

type contextKey string

const (
	ctxKeyUserID    contextKey = "nxd_user_id"
	ctxKeyFactoryID contextKey = "nxd_factory_id"
	ctxKeyRole      contextKey = "nxd_role"
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

// RequireFactory reads factory_id from query "factory_id" or header "X-NXD-Factory-ID",
// verifies the user (from context or auth) is a member of that factory, and sets factory_id and role in context.
// Returns 403 and writes to audit_log on cross-factory access. Must be used after RequireAuth (or with user in context).
func RequireFactory(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := store.NXDDB()
		if db == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
			return
		}
		user := api.UserFromRequest(r)
		userID := ""
		if user != nil {
			userID = user.Email
			if userID == "" {
				userID = fmt.Sprintf("id:%d", user.ID)
			}
		}
		if userID == "" {
			if v := r.Context().Value(ctxKeyUserID); v != nil {
				userID = v.(string)
			}
		}
		if userID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "faça login"})
			return
		}
		factoryIDStr := r.URL.Query().Get("factory_id")
		if factoryIDStr == "" {
			factoryIDStr = r.Header.Get("X-NXD-Factory-ID")
		}
		if factoryIDStr == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "factory_id obrigatório (query ou header X-NXD-Factory-ID)"})
			return
		}
		factoryID, err := uuid.Parse(factoryIDStr)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "factory_id inválido"})
			return
		}
		role, err := store.GetMemberRole(db, factoryID, userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "erro ao verificar permissão"})
			return
		}
		if role == "" {
			_ = store.AuditWrite(db, userID, &factoryID, "ACCESS_DENIED", "FACTORY", factoryIDStr, "", map[string]string{"reason": "not_member"})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "acesso negado a esta fábrica"})
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxKeyUserID, userID)
		ctx = context.WithValue(ctx, ctxKeyFactoryID, factoryID)
		ctx = context.WithValue(ctx, ctxKeyRole, role)
		next(w, r.WithContext(ctx))
	}
}

// UserIDFromRequest returns the NXD user_id (email) from context, or from auth as fallback.
func UserIDFromRequest(r *http.Request) string {
	if v := r.Context().Value(ctxKeyUserID); v != nil {
		return v.(string)
	}
	user := api.UserFromRequest(r)
	if user != nil && user.Email != "" {
		return user.Email
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

// RequireRole wraps a handler and returns 403 if the context role is not in the allowed list.
func RequireRole(allowedRoles []string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		role := RoleFromContext(r.Context())
		if !store.RoleAllowed(role, allowedRoles) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "permissão insuficiente"})
			return
		}
		next(w, r)
	}
}
