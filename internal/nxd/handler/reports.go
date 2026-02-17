package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"hubsystem/internal/nxd/ai"
	"hubsystem/internal/nxd/middleware"
	"hubsystem/internal/nxd/store"

	"github.com/google/uuid"
)

// ListReportTemplates returns all report templates (Top 50 from seed).
func ListReportTemplates(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
		return
	}
	list, err := store.ListReportTemplates(db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"templates": list})
}

// RunReportRequest is the body for POST /nxd/reports/run.
type RunReportRequest struct {
	TemplateID string   `json:"template_id"`
	GroupID    string   `json:"group_id"`
	AssetIDs   []string `json:"asset_ids"`
	Period     string   `json:"period"`
	Detail     string   `json:"detail"`
	Nicho      string   `json:"nicho"`
}

// RunReport creates a report run, builds prompt contract, and optionally calls AI; returns run id and status.
func RunReport(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
		return
	}
	factoryID, ok := middleware.FactoryIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "factory_id obrigatório"})
		return
	}
	userID := middleware.UserIDFromRequest(r)
	var req RunReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "payload inválido"})
		return
	}
	contract := map[string]interface{}{
		"context":   map[string]string{"factory_id": factoryID.String(), "period": req.Period},
		"objective": map[string]string{"nicho": req.Nicho, "detail": req.Detail},
		"constraints": []string{"não inventar", "use apenas dados fornecidos", "se faltar dado, marque INSUFICIENTE"},
	}
	contractJSON, _ := json.Marshal(contract)
	filtersJSON, _ := json.Marshal(map[string]interface{}{
		"template_id": req.TemplateID, "group_id": req.GroupID, "asset_ids": req.AssetIDs,
		"period": req.Period, "detail": req.Detail, "nicho": req.Nicho,
	})
	id, err := store.CreateReportRun(db, factoryID, userID, filtersJSON, contractJSON)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "erro ao criar relatório"})
		return
	}
	var result map[string]interface{}
	var resultJSON []byte
	if ai.IsConfigured() {
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
		defer cancel()
		client, errClient := ai.NewClient(ctx)
		if errClient != nil {
			log.Printf("[nxd] Vertex client: %v", errClient)
		} else if client != nil {
			prompt := "Contexto do relatório (JSON): " + string(contractJSON) + "\n\nGere o relatório em JSON conforme o schema (title, summary_bullets, kpis, findings, charts, risks_and_assumptions, missing_data, auditability). Use apenas os dados acima; não invente."
			generated, errGen := ai.GenerateReport(ctx, client, prompt)
			if errGen != nil {
				log.Printf("[nxd] Vertex GenerateReport: %v", errGen)
			} else if len(generated) > 0 {
				resultJSON = generated
				_ = json.Unmarshal(generated, &result)
			}
		}
	}
	if result == nil {
		result = map[string]interface{}{
			"title":                 "Relatório NXD",
			"summary_bullets":       []string{"Relatório gerado.", "Vertex AI indisponível ou não configurado; use stub."},
			"kpis":                  []map[string]interface{}{},
			"findings":              []map[string]interface{}{},
			"charts":                 []map[string]interface{}{},
			"risks_and_assumptions": "Dados de exemplo.",
			"missing_data":          []string{},
			"auditability":          map[string]interface{}{"data_window": req.Period, "sources": "nxd", "rollup_used": false},
		}
		resultJSON, _ = json.Marshal(result)
	}
	_ = store.UpdateReportRunStatus(db, id, factoryID, "DONE", resultJSON, "")
	_ = store.AuditWrite(db, userID, &factoryID, "REPORT_GENERATE", "REPORT", id.String(), "", nil)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     id.String(),
		"status": "DONE",
		"result": result,
	})
}

// GetReport returns a report run by id.
func GetReport(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
		return
	}
	factoryID, ok := middleware.FactoryIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "factory_id obrigatório"})
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id inválido"})
		return
	}
	run, err := store.GetReportRun(db, id, factoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if run == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "relatório não encontrado"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(run)
}

// ExportPDF stub: returns a placeholder. Real implementation would generate PDF, upload to GCS, return signed URL.
func ExportPDF(w http.ResponseWriter, r *http.Request) {
	db := store.NXDDB()
	if db == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"error": "NXD não configurado"})
		return
	}
	factoryID, ok := middleware.FactoryIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "factory_id obrigatório"})
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	run, _ := store.GetReportRun(db, id, factoryID)
	if run == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Stub: no bucket configured; return placeholder URL
	placeholderURL := ""
	if run.ExportURL != "" {
		placeholderURL = run.ExportURL
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"export_url": placeholderURL,
		"message":    "Exportação PDF: configure GCS bucket (nxd-reports-exports) para URL assinada.",
	})
}
