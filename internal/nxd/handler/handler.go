package handler

import (
	"encoding/json"
	"net/http"

	"hubsystem/internal/nxd/store"
)

// Ready returns 200 if NXD Postgres is configured and connected; 503 otherwise.
func Ready(w http.ResponseWriter, r *http.Request) {
	if store.NXDDB() == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "nxd_disabled",
			"reason": "NXD_DATABASE_URL not set",
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"nxd":    "enabled",
	})
}
