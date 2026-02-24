package api

import (
	"encoding/json"
	"net/http"
	"os"
)

// GetBillingPlanHandler retorna o plano e dados de cobrança do usuário (fábrica).
func GetBillingPlanHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}
	db := GetDB()
	var plan, nextBilling string
	err := db.QueryRow(`
		SELECT COALESCE(subscription_plan, 'free'), COALESCE(CAST(next_billing_date AS TEXT), '')
		FROM factories WHERE user_id = $1 LIMIT 1
	`, userID).Scan(&plan, &nextBilling)
	if err != nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}
	checkoutURL := os.Getenv("STRIPE_CHECKOUT_URL") // opcional: link Stripe ou página de pagamento
	if checkoutURL == "" {
		checkoutURL = os.Getenv("BILLING_CHECKOUT_URL") // fallback genérico
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plan":           plan,
		"next_billing":   nextBilling,
		"checkout_url":   checkoutURL,
	})
}

// UpdateBillingPlanHandler atualiza o plano (admin ou após confirmação de pagamento).
func UpdateBillingPlanHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}
	var req struct {
		Plan          string `json:"plan"`
		NextBilling   string `json:"next_billing"` // YYYY-MM-DD
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if req.Plan == "" {
		req.Plan = "free"
	}
	db := GetDB()
	var nextBilling interface{} = req.NextBilling
	if req.NextBilling == "" {
		nextBilling = nil
	}
	_, err := db.Exec(`
		UPDATE factories SET subscription_plan = $1, next_billing_date = $2 WHERE user_id = $3
	`, req.Plan, nextBilling, userID)
	if err != nil {
		http.Error(w, "Erro ao atualizar plano", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
