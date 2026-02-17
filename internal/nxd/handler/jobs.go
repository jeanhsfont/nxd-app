package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"hubsystem/internal/nxd/store"
)

// RollupJob runs the telemetry rollup for the cold window [now-120d, now-90d]. Call from Cloud Scheduler.
func RollupJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	secret := os.Getenv("NXD_JOB_SECRET")
	if secret != "" && r.Header.Get("X-NXD-Job-Secret") != secret {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "não autorizado"})
		return
	}
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
		return
	}
	now := time.Now()
	from := now.AddDate(0, 0, -120)
	to := now.AddDate(0, 0, -90)
	n, err := store.RunRollup(db, from, to)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"rows":   n,
		"window": "120d..90d",
	})
}

// AlertsJob evaluates alert rules and creates alerts. Call from Cloud Scheduler every 5–10 min.
func AlertsJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	secret := os.Getenv("NXD_JOB_SECRET")
	if secret != "" && r.Header.Get("X-NXD-Job-Secret") != secret {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "não autorizado"})
		return
	}
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
		return
	}
	// Minimal: no-op for now; real implementation would query telemetry and rules, then CreateAlert where condition met
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok", "evaluated": 0})
}
