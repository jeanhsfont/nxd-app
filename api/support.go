package api

import (
	"encoding/json"
	"net/http"
)

// CreateSupportTicketHandler cria um ticket de suporte (persistido no banco).
func CreateSupportTicketHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}
	var req struct {
		Subject string `json:"subject"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if req.Subject == "" || req.Message == "" {
		http.Error(w, "Assunto e mensagem são obrigatórios", http.StatusBadRequest)
		return
	}
	db := GetDB()
	var email string
	if err := db.QueryRow("SELECT email FROM users WHERE id = $1", userID).Scan(&email); err != nil {
		http.Error(w, "Usuário não encontrado", http.StatusInternalServerError)
		return
	}
	_, err := db.Exec(`
		INSERT INTO support_tickets (user_id, email, subject, message, status) VALUES ($1, $2, $3, $4, 'open')
	`, userID, email, req.Subject, req.Message)
	if err != nil {
		http.Error(w, "Erro ao registrar ticket", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Ticket registrado. Entraremos em contato em breve."})
}
