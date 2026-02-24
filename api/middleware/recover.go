package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// RecoverMiddleware captura panic em handlers, loga stack em stdout e retorna 500.
// Permite ver em Cloud Run logs a causa exata do crash.
func RecoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()
				log.Printf("[PANIC] %s %s: %v\n%s", r.Method, r.URL.Path, rec, stack)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"Erro interno. Tente novamente."}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
