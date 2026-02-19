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
