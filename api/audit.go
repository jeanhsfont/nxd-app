package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// LogAudit registra uma ação na trilha de auditoria.
func LogAudit(userID int64, action, entityType, entityID, oldValue, newValue, ip string) {
	d := GetDB()
	if d == nil {
		return
	}
	if ip == "" {
		ip = "-"
	}
	if entityID == "" {
		entityID = "-"
	}
	_, _ = d.Exec(
		`INSERT INTO audit_log (user_id, action, entity_type, entity_id, old_value, new_value, ip) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, action, entityType, entityID, truncateForAudit(oldValue, 2000), truncateForAudit(newValue, 2000), truncateForAudit(ip, 64),
	)
}

func truncateForAudit(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// ClientIP retorna o IP do cliente a partir do request.
func ClientIP(r *http.Request) string {
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		if i := strings.Index(x, ","); i > 0 {
			return strings.TrimSpace(x[:i])
		}
		return strings.TrimSpace(x)
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

// AuditEntry representa uma linha do audit_log.
type AuditEntry struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Action     string    `json:"action"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	OldValue   string    `json:"old_value,omitempty"`
	NewValue   string    `json:"new_value,omitempty"`
	IP         string    `json:"ip,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// ListAuditLogHandler — GET /api/admin/audit-log — lista audit log (admin). Query: entity_type, action, limit.
func ListAuditLogHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	if !userHasRole(userID, "admin") {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}
	d := GetDB()
	if d == nil {
		http.Error(w, "Banco indisponível", http.StatusServiceUnavailable)
		return
	}
	entityType := r.URL.Query().Get("entity_type")
	action := r.URL.Query().Get("action")
	limit := 100
	query := `SELECT id, user_id, action, entity_type, entity_id, old_value, new_value, ip, created_at FROM audit_log WHERE 1=1`
	args := []interface{}{}
	argNum := 1
	if entityType != "" {
		query += fmt.Sprintf(" AND entity_type = $%d", argNum)
		args = append(args, entityType)
		argNum++
	}
	if action != "" {
		query += fmt.Sprintf(" AND action = $%d", argNum)
		args = append(args, action)
		argNum++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d", limit)
	rows, err := d.Query(query, args...)
	if err != nil {
		http.Error(w, "Erro ao listar audit log", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		err := rows.Scan(&e.ID, &e.UserID, &e.Action, &e.EntityType, &e.EntityID, &e.OldValue, &e.NewValue, &e.IP, &e.CreatedAt)
		if err != nil {
			continue
		}
		entries = append(entries, e)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"entries": entries})
}
