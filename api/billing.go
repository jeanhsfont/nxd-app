package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// Limites por plano (ex.: máx. ativos).
var planLimits = map[string]int{"free": 5, "pro": 50, "enterprise": -1} // -1 = ilimitado

// GetBillingPlanHandler retorna o plano, status, limites e dados de cobrança do usuário (fábrica).
func GetBillingPlanHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}
	db := GetDB()
	if db == nil {
		http.Error(w, "Indisponível", http.StatusServiceUnavailable)
		return
	}
	var plan, nextBilling, status string
	var trialEndsAt *time.Time
	var gatewaySubID *string
	err := db.QueryRow(`
		SELECT COALESCE(subscription_plan, 'free'), COALESCE(CAST(next_billing_date AS TEXT), ''),
		       COALESCE(subscription_status, 'active'), trial_ends_at, gateway_subscription_id
		FROM factories WHERE user_id = $1 LIMIT 1
	`, userID).Scan(&plan, &nextBilling, &status, &trialEndsAt, &gatewaySubID)
	if err != nil {
		http.Error(w, "Fábrica não encontrada", http.StatusNotFound)
		return
	}
	limit := planLimits[plan]
	if limit < 0 {
		limit = 99999
	}
	trialEnds := ""
	if trialEndsAt != nil {
		trialEnds = trialEndsAt.Format("2006-01-02")
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"plan":                   plan,
		"status":                 status,
		"next_billing":           nextBilling,
		"trial_ends_at":          trialEnds,
		"max_assets":             limit,
		"checkout_url":           getCheckoutBaseURL(),
		"gateway_subscription_id": gatewaySubID,
	})
}

func getCheckoutBaseURL() string {
	if u := os.Getenv("STRIPE_CHECKOUT_URL"); u != "" {
		return u
	}
	return os.Getenv("BILLING_CHECKOUT_URL")
}

// CreateCheckoutSessionHandler — POST /api/billing/create-checkout-session
// Body: { "plan": "pro"|"enterprise", "success_url", "cancel_url" }. Retorna { "url": "..." } para redirecionar ao gateway.
func CreateCheckoutSessionHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}
	var req struct {
		Plan       string `json:"plan"`
		SuccessURL string `json:"success_url"`
		CancelURL  string `json:"cancel_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if req.Plan != "pro" && req.Plan != "enterprise" {
		req.Plan = "pro"
	}
	base := getCheckoutBaseURL()
	if base == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"url": "",
			"error": "Checkout não configurado. Defina STRIPE_CHECKOUT_URL ou BILLING_CHECKOUT_URL.",
		})
		return
	}
	// Monta URL de checkout (gateway pode exigir query params; Stripe usa session_id após criar sessão).
	// Aqui retornamos base + plan para redirect genérico; integração Stripe criaria sessão e retornaria session.url
	checkoutURL := base
	if req.SuccessURL != "" {
		checkoutURL += "?success_url=" + req.SuccessURL
	}
	if req.CancelURL != "" {
		checkoutURL += "&cancel_url=" + req.CancelURL
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"url":   checkoutURL,
		"plan":  req.Plan,
		"user_id": userID,
	})
}

// BillingWebhookHandler — POST /api/billing/webhook — recebe eventos do gateway (Stripe, etc.).
// Em produção: validar assinatura (Stripe-Signature). Atualiza subscription_status e plan conforme evento.
func BillingWebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Type     string `json:"type"`
		UserID   int64  `json:"user_id"`
		FactoryID int64 `json:"factory_id"`
		Plan     string `json:"plan"`
		Status   string `json:"status"`
		SubscriptionID string `json:"subscription_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	// TODO: em produção validar Stripe-Signature com STRIPE_WEBHOOK_SECRET
	db := GetDB()
	if db == nil {
		http.Error(w, "Indisponível", http.StatusServiceUnavailable)
		return
	}
	if payload.Status == "" {
		payload.Status = "active"
	}
	if payload.Plan == "" {
		payload.Plan = "free"
	}
	var targetUserID int64
	if payload.UserID != 0 {
		targetUserID = payload.UserID
	} else if payload.FactoryID != 0 {
		_ = db.QueryRow("SELECT user_id FROM factories WHERE id = $1", payload.FactoryID).Scan(&targetUserID)
	}
	if targetUserID == 0 {
		log.Printf("[BillingWebhook] ignorando evento sem user_id/factory_id: %+v", payload)
		w.WriteHeader(http.StatusOK)
		return
	}
	_, err := db.Exec(`
		UPDATE factories SET subscription_plan = $1, subscription_status = $2, gateway_subscription_id = $3
		WHERE user_id = $4
	`, payload.Plan, payload.Status, nilIfEmpty(payload.SubscriptionID), targetUserID)
	if err != nil {
		log.Printf("[BillingWebhook] update error: %v", err)
		http.Error(w, "Erro ao atualizar", http.StatusInternalServerError)
		return
	}
	log.Printf("[BillingWebhook] atualizado user_id=%d plan=%s status=%s", targetUserID, payload.Plan, payload.Status)
	w.WriteHeader(http.StatusOK)
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// UpdateBillingPlanHandler atualiza o plano (admin ou após confirmação de pagamento).
func UpdateBillingPlanHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autorizado", http.StatusUnauthorized)
		return
	}
	var req struct {
		Plan        string `json:"plan"`
		NextBilling string `json:"next_billing"` // YYYY-MM-DD
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if req.Plan == "" {
		req.Plan = "free"
	}
	db := GetDB()
	if db == nil {
		http.Error(w, "Indisponível", http.StatusServiceUnavailable)
		return
	}
	var nextBilling interface{} = req.NextBilling
	if req.NextBilling == "" {
		nextBilling = nil
	}
	if req.Status == "" {
		req.Status = "active"
	}
	_, err := db.Exec(`
		UPDATE factories SET subscription_plan = $1, next_billing_date = $2, subscription_status = $3 WHERE user_id = $4
	`, req.Plan, nextBilling, req.Status, userID)
	if err != nil {
		http.Error(w, "Erro ao atualizar plano", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
