package api

// import_jobs_handler.go — Admin endpoints for "Download Longo" import jobs
//
// All endpoints require JWT authentication (registered under authRouter in main.go).
// The factory_id is always inferred from the authenticated user's ID — users
// can only manage their own factory's jobs.
//
// Routes (all under /api, JWT-protected):
//   GET    /api/admin/import-jobs              — list recent jobs (default: last 20)
//   POST   /api/admin/import-jobs              — create a new job
//   GET    /api/admin/import-jobs/{id}         — get job status + progress
//   POST   /api/admin/import-jobs/{id}/cancel  — cancel a pending/running job
//   POST   /api/admin/import-jobs/{id}/retry   — reset failed/cancelled to pending

import (
	"encoding/json"
	"fmt"
	"hubsystem/internal/nxd/store"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// ListImportJobsHandler — GET /api/admin/import-jobs
func ListImportJobsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "factory not found", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "NXD store not available", http.StatusServiceUnavailable)
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	jobs, err := store.ListImportJobs(nxdDB, factoryID, limit)
	if err != nil {
		log.Printf("❌ [ImportJobs] ListImportJobs error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if jobs == nil {
		jobs = []store.ImportJobStatus{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs":  jobs,
		"count": len(jobs),
	})
}

// CreateImportJobHandler — POST /api/admin/import-jobs
//
// Body (JSON):
//
//	{
//	  "source_type": "memory",                    // required: "memory" | "dx_http"
//	  "asset_id":    "<uuid>",                    // optional
//	  "batch_size":  1000,                        // optional, default 1000
//	  "period_start": "2024-01-01T00:00:00Z",     // optional, for record-keeping
//	  "period_end":   "2024-12-31T23:59:59Z",     // optional
//	  "source_config": { ... }                    // optional, passed to the worker
//	}
func CreateImportJobHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "factory not found", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "NXD store not available", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		SourceType   string          `json:"source_type"`
		AssetID      string          `json:"asset_id"`
		BatchSize    int             `json:"batch_size"`
		PeriodStart  string          `json:"period_start"`
		PeriodEnd    string          `json:"period_end"`
		SourceConfig json.RawMessage `json:"source_config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.SourceType == "" {
		body.SourceType = "memory"
	}

	params := store.CreateImportJobParams{
		FactoryID:    factoryID,
		SourceType:   body.SourceType,
		SourceConfig: body.SourceConfig,
		BatchSize:    body.BatchSize,
	}

	// Parse optional asset_id
	if body.AssetID != "" {
		aid, err := uuid.Parse(body.AssetID)
		if err != nil {
			http.Error(w, "invalid asset_id", http.StatusBadRequest)
			return
		}
		params.AssetID = &aid
	}

	// Parse optional period fields
	if body.PeriodStart != "" {
		t, err := time.Parse(time.RFC3339, body.PeriodStart)
		if err != nil {
			http.Error(w, "invalid period_start (use RFC3339)", http.StatusBadRequest)
			return
		}
		params.PeriodStart = &t
	}
	if body.PeriodEnd != "" {
		t, err := time.Parse(time.RFC3339, body.PeriodEnd)
		if err != nil {
			http.Error(w, "invalid period_end (use RFC3339)", http.StatusBadRequest)
			return
		}
		params.PeriodEnd = &t
	}

	jobID, err := store.CreateImportJob(nxdDB, params)
	if err != nil {
		log.Printf("❌ [ImportJobs] CreateImportJob error: %v", err)
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":  jobID.String(),
		"status":  "pending",
		"message": "Job created. Worker will pick it up within 5 seconds.",
	})
}

// GetImportJobHandler — GET /api/admin/import-jobs/{id}
func GetImportJobHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "factory not found", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "NXD store not available", http.StatusServiceUnavailable)
		return
	}

	jobIDStr := mux.Vars(r)["id"]
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}

	job, err := store.GetImportJob(nxdDB, jobID, factoryID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if job == nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// CancelImportJobHandler — POST /api/admin/import-jobs/{id}/cancel
func CancelImportJobHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "factory not found", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "NXD store not available", http.StatusServiceUnavailable)
		return
	}

	jobIDStr := mux.Vars(r)["id"]
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}

	if err := store.CancelImportJob(nxdDB, jobID, factoryID); err != nil {
		log.Printf("⚠️  [ImportJobs] CancelImportJob %s: %v", jobID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"job_id":  jobID.String(),
		"status":  "cancelled",
		"message": "Cancellation requested. Worker will stop within one batch cycle.",
	})
}

// RetryImportJobHandler — POST /api/admin/import-jobs/{id}/retry
func RetryImportJobHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "factory not found", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "NXD store not available", http.StatusServiceUnavailable)
		return
	}

	jobIDStr := mux.Vars(r)["id"]
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}

	if err := store.RetryImportJob(nxdDB, jobID, factoryID); err != nil {
		log.Printf("⚠️  [ImportJobs] RetryImportJob %s: %v", jobID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"job_id":  jobID.String(),
		"status":  "pending",
		"message": "Job reset to pending. Worker will retry within 5 seconds.",
	})
}

// SubmitImportJobDataHandler — POST /api/admin/import-jobs/{id}/data
//
// Submete os dados JSON diretamente para um job do tipo 'memory' que já existe
// com status='pending'. O handler converte o payload JSON em []BulkTelemetryRow
// e chama store.SubmitMemoryJob(), que envia os dados via canal interno para o
// worker processar imediatamente.
//
// Body (JSON):
//
//	{
//	  "asset_id":   "<uuid>",                 // obrigatório
//	  "rows": [
//	    {
//	      "ts":           "2024-01-15T10:30:00Z",  // RFC3339
//	      "metric_key":   "Temperatura_Motor",
//	      "metric_value": 87.5,
//	      "status":       "OK"                     // opcional
//	    },
//	    ...
//	  ]
//	}
//
// A resposta inclui o número de linhas processadas e o status final do job.
// NOTA: este endpoint bloqueia até o worker terminar de processar o job.
// Para jobs grandes (>100k linhas), use um client com timeout adequado.
func SubmitImportJobDataHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil || factoryID == uuid.Nil {
		http.Error(w, "factory not found", http.StatusNotFound)
		return
	}
	nxdDB := store.NXDDB()
	if nxdDB == nil {
		http.Error(w, "NXD store not available", http.StatusServiceUnavailable)
		return
	}

	jobIDStr := mux.Vars(r)["id"]
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}

	// Verifica que o job existe e pertence à factory do usuário
	job, err := store.GetImportJob(nxdDB, jobID, factoryID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if job == nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	// Accept pending jobs; if it was marked failed by poll worker (race), auto-reset to pending first
	if job.Status == "failed" && job.SourceType == "memory" {
		// Reset to pending so SubmitMemoryJob can claim it
		nxdDB.Exec(`UPDATE nxd.import_jobs SET status='pending', error_message=NULL, rows_done=0, updated_at=NOW() WHERE id=$1`, jobID)
		job.Status = "pending"
		log.Printf("🔁 [ImportJobs] Auto-reset failed memory job %s to pending for data submission", jobID)
	}
	if job.Status != "pending" {
		http.Error(w, fmt.Sprintf("job is in status=%s; only pending jobs accept data (cancel and create a new job)", job.Status), http.StatusConflict)
		return
	}
	if job.SourceType != "memory" {
		http.Error(w, fmt.Sprintf("job source_type=%s; this endpoint only accepts 'memory' jobs", job.SourceType), http.StatusBadRequest)
		return
	}

	// Parse do payload
	var body struct {
		AssetID string `json:"asset_id"`
		Rows    []struct {
			Ts          string  `json:"ts"`
			MetricKey   string  `json:"metric_key"`
			MetricValue float64 `json:"metric_value"`
			Status      string  `json:"status"`
		} `json:"rows"`
	}
	// Limite de 50 MB para o corpo — permite até ~500k linhas de telemetria
	r.Body = http.MaxBytesReader(w, r.Body, 50*1024*1024)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		if err.Error() == "http: request body too large" {
			http.Error(w, "payload exceeds 50MB limit", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	if body.AssetID == "" {
		http.Error(w, "asset_id is required", http.StatusBadRequest)
		return
	}
	assetID, err := uuid.Parse(body.AssetID)
	if err != nil {
		http.Error(w, "invalid asset_id", http.StatusBadRequest)
		return
	}
	if len(body.Rows) == 0 {
		http.Error(w, "rows array is empty", http.StatusBadRequest)
		return
	}

	// Converte para []store.BulkTelemetryRow
	bulkRows := make([]store.BulkTelemetryRow, 0, len(body.Rows))
	skipped := 0
	for i, row := range body.Rows {
		if row.MetricKey == "" {
			skipped++
			continue
		}
		ts := time.Now()
		if row.Ts != "" {
			parsed, parseErr := time.Parse(time.RFC3339, row.Ts)
			if parseErr != nil {
				http.Error(w, fmt.Sprintf("row %d: invalid ts=%q (use RFC3339 e.g. 2024-01-15T10:30:00Z)", i, row.Ts), http.StatusBadRequest)
				return
			}
			ts = parsed
		}
		status := row.Status
		if status == "" {
			status = "OK"
		}
		bulkRows = append(bulkRows, store.BulkTelemetryRow{
			Ts:            ts,
			FactoryID:     factoryID,
			AssetID:       assetID,
			MetricKey:     row.MetricKey,
			MetricValue:   row.MetricValue,
			Status:        status,
			CorrelationID: jobID.String(),
		})
	}
	if len(bulkRows) == 0 {
		http.Error(w, "no valid rows after parsing (all skipped due to missing metric_key)", http.StatusBadRequest)
		return
	}

	// Atualiza rows_total no job para que o frontend possa mostrar progresso real
	nxdDB.Exec(`UPDATE nxd.import_jobs SET rows_total = $1, updated_at = NOW() WHERE id = $2`, len(bulkRows), jobID)

	log.Printf("📤 [ImportJobs] SubmitData: job=%s asset=%s rows=%d skipped=%d factory=%s",
		jobID, assetID, len(bulkRows), skipped, factoryID)

	// Envia para o worker via canal. Bloqueia até o worker processar.
	// O worker validará idempotência (range check) antes de inserir.
	if err := store.SubmitMemoryJob(jobID, bulkRows); err != nil {
		log.Printf("❌ [ImportJobs] SubmitMemoryJob error: %v", err)
		http.Error(w, "worker processing failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id":        jobID.String(),
		"status":        "done",
		"rows_submitted": len(bulkRows),
		"rows_skipped":  skipped,
		"message":       "Data processed successfully by the import worker.",
	})
}
