package store

// importer.go — "Download Longo" background worker for nxd.import_jobs
//
// Architecture:
//   - RunImportWorker() starts a goroutine that polls import_jobs every
//     pollInterval seconds for jobs with status='pending'.
//   - Each job is claimed atomically (UPDATE … WHERE status='pending' RETURNING id)
//     before processing to avoid duplicate runs in multi-instance deployments.
//   - Jobs with source_type='file_csv' process an in-memory []BulkTelemetryRow
//     slice supplied via SubmitCSVJob (for future: streaming CSV reader).
//   - Jobs with source_type='dx_http' (future) will connect to the DX endpoint
//     stored in source_config JSONB and page through historical data.
//   - Cancellation is checked each batch via isCancelled(db, jobID).
//   - Progress (rows_done, rows_total) is written every batch so the UI can poll.
//   - On success: status='done', finished_at=NOW().
//   - On error: status='failed', error_message=err.Error(), finished_at=NOW().
//   - On cancellation: status='cancelled', finished_at=NOW().
//
// Cloud Run note:
//   This goroutine runs in the same process as the HTTP server.  Cloud Run may
//   spin up multiple instances; each instance independently polls the job table
//   and uses the atomic claim pattern to avoid double-processing.  If the
//   instance is killed mid-job, the job remains 'running' — use the admin
//   endpoint POST /api/admin/import-jobs/{id}/retry to reset it to 'pending'.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	workerPollInterval  = 5 * time.Second  // how often to look for pending jobs
	workerBatchSize     = 1000             // rows per DB transaction (overridden by job.batch_size)
	workerMaxRuntime    = 6 * time.Hour    // hard timeout per job
	workerBatchThrottle = 50 * time.Millisecond // P5: sleep between batches to not starve real-time ingest
)

// ImportJob mirrors the nxd.import_jobs row we need for processing.
type ImportJob struct {
	ID           uuid.UUID
	FactoryID    uuid.UUID
	AssetID      *uuid.UUID
	Status       string
	SourceType   string
	SourceConfig json.RawMessage
	BatchSize    int
	RowsTotal    int64
	RowsDone     int64
}

// ─── In-process job queue ────────────────────────────────────────────────────
// For source_type='memory' jobs submitted directly (e.g. from the import-jobs
// API handler with pre-loaded data), the caller sends rows via this channel.
// The worker drains one job at a time from this channel during processing.

type memoryJobPayload struct {
	JobID   uuid.UUID
	Rows    []BulkTelemetryRow
	ErrChan chan error
}

var memoryJobChan = make(chan memoryJobPayload, 8)

// SubmitMemoryJob enqueues a batch of pre-loaded rows for a job that already
// exists in nxd.import_jobs with status='pending' and source_type='memory'.
// Returns the job error (nil = success). Blocks until the job is processed.
func SubmitMemoryJob(jobID uuid.UUID, rows []BulkTelemetryRow) error {
	errChan := make(chan error, 1)
	memoryJobChan <- memoryJobPayload{JobID: jobID, Rows: rows, ErrChan: errChan}
	return <-errChan
}

// ─── Worker lifecycle ─────────────────────────────────────────────────────────

// RecoverStaleJobs resets jobs stuck in 'running' state after a restart.
// Cloud Run instances are stateless — if the instance is killed mid-job,
// the job remains 'running' forever. We detect these on startup and reset
// them to 'pending' so they are retried automatically.
// Only jobs that have been 'running' for more than staleCutoff are reset.
func RecoverStaleJobs(db *sql.DB) {
	const staleCutoff = 10 * time.Minute
	res, err := db.Exec(`
		UPDATE nxd.import_jobs
		SET status = 'pending', error_message = 'Worker restarted (auto-recovery)',
		    rows_done = 0, started_at = NULL, updated_at = NOW()
		WHERE status = 'running'
		  AND updated_at < NOW() - $1::interval
	`, staleCutoff.String())
	if err != nil {
		log.Printf("⚠️  [ImportWorker] RecoverStaleJobs error: %v", err)
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		log.Printf("🔁 [ImportWorker] Recovered %d stale running job(s) → pending", n)
	}
}

// RunImportWorker starts the background job processor.  Call once from main()
// after the DB is initialized.  ctx should be cancelled on server shutdown.
func RunImportWorker(ctx context.Context, db *sql.DB) {
	// Recover stale jobs from previous instance before polling.
	RecoverStaleJobs(db)
	log.Println("✓ [ImportWorker] Background import worker started (poll interval: 5s)")
	ticker := time.NewTicker(workerPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("⏹  [ImportWorker] Shutdown signal received, worker stopping.")
			return

		case payload := <-memoryJobChan:
			// A memory job was submitted directly — process it immediately.
			err := processMemoryJob(ctx, db, payload.JobID, payload.Rows)
			payload.ErrChan <- err

		case <-ticker.C:
			// Poll for pending jobs in the DB.
			if err := claimAndProcessPending(ctx, db); err != nil {
				log.Printf("⚠️  [ImportWorker] Poll error: %v", err)
			}
		}
	}
}

// claimAndProcessPending finds one pending job, claims it, and runs it.
func claimAndProcessPending(ctx context.Context, db *sql.DB) error {
	// Atomic claim: transition pending → running in one UPDATE+RETURNING.
	// If two instances race, only one gets the row (PostgreSQL row-level lock).
	var job ImportJob
	var factoryIDStr, jobIDStr string
	var assetIDStr sql.NullString

	err := db.QueryRowContext(ctx, `
		UPDATE nxd.import_jobs
		SET status = 'running', started_at = NOW(), updated_at = NOW()
		WHERE id = (
			SELECT id FROM nxd.import_jobs
			WHERE status = 'pending'
			ORDER BY created_at
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, factory_id, asset_id, source_type, source_config, batch_size, rows_total, rows_done
	`).Scan(&jobIDStr, &factoryIDStr, &assetIDStr,
		&job.SourceType, &job.SourceConfig,
		&job.BatchSize, &job.RowsTotal, &job.RowsDone)

	if err == sql.ErrNoRows {
		return nil // nothing pending — normal
	}
	if err != nil {
		return fmt.Errorf("claim job: %w", err)
	}

	job.ID, _ = uuid.Parse(jobIDStr)
	job.FactoryID, _ = uuid.Parse(factoryIDStr)
	if assetIDStr.Valid {
		aid, _ := uuid.Parse(assetIDStr.String)
		job.AssetID = &aid
	}
	if job.BatchSize <= 0 {
		job.BatchSize = workerBatchSize
	}

	log.Printf("🔄 [ImportWorker] Claimed job %s (factory=%s, type=%s)", job.ID, job.FactoryID, job.SourceType)
	return dispatchJob(ctx, db, &job)
}

// dispatchJob routes the job to the appropriate processor by source_type.
func dispatchJob(ctx context.Context, db *sql.DB, job *ImportJob) error {
	jobCtx, cancel := context.WithTimeout(ctx, workerMaxRuntime)
	defer cancel()

	var err error
	switch {
	case job.SourceType == "memory":
		// Memory jobs are processed via the memoryJobChan path; if we get here
		// via polling it means the channel wasn't used — mark as failed.
		err = fmt.Errorf("source_type=memory job found via poll (no data channel); use SubmitMemoryJob")

	case strings.HasPrefix(job.SourceType, "dx_http"):
		err = processDXHTTPJob(jobCtx, db, job)

	default:
		err = fmt.Errorf("unknown source_type: %q", job.SourceType)
	}

	if err != nil {
		markJobFailed(db, job.ID, err)
	}
	return err
}

// ─── Memory job processor ────────────────────────────────────────────────────

func processMemoryJob(ctx context.Context, db *sql.DB, jobID uuid.UUID, rows []BulkTelemetryRow) error {
	// Claim the job (it must be 'pending' in the DB).
	res, err := db.ExecContext(ctx, `
		UPDATE nxd.import_jobs
		SET status = 'running', started_at = NOW(), rows_total = $2, updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, jobID, int64(len(rows)))
	if err != nil {
		return fmt.Errorf("claim memory job %s: %w", jobID, err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("job %s not found or not in pending state", jobID)
	}

	log.Printf("🔄 [ImportWorker] Memory job %s started (%d rows)", jobID, len(rows))

	// Query batch_size for this job.
	var batchSize int
	_ = db.QueryRowContext(ctx, `SELECT batch_size FROM nxd.import_jobs WHERE id = $1`, jobID).Scan(&batchSize)
	if batchSize <= 0 {
		batchSize = workerBatchSize
	}

	var rowsDone int64
	start := time.Now()

	for offset := 0; offset < len(rows); offset += batchSize {
		// Check cancellation before each batch.
		if cancelled, _ := isCancelled(db, jobID); cancelled {
			markJobCancelled(db, jobID, rowsDone)
			return nil
		}

		end := offset + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[offset:end]

		batchStart := time.Now()

		// P4 — Idempotency: range check before inserting.
		// Strategy: RANGE CHECK per batch.
		// We check whether any rows already exist for this asset over the time
		// range of the current batch. If rows exist, we skip the batch and log
		// a warning. This prevents double-import if the same job is retried.
		//
		// Trade-off: if only SOME rows in the range exist (partial previous run),
		// the entire batch is skipped. This is intentional — partial inserts
		// from a crashed job are indistinguishable from legitimate data.
		// Full re-import requires cancelling the previous job and using a fresh
		// asset or a different time range.
		if len(batch) > 0 && batch[0].AssetID != (uuid.UUID{}) {
			existing, checkErr := countRangeRows(db, batch[0].AssetID, batch[0].Ts, batch[len(batch)-1].Ts)
			if checkErr != nil {
				log.Printf("⚠️  [Job %s] range check error (skipping idempotency): %v", jobID, checkErr)
			} else if existing > 0 {
				log.Printf("⚠️  [Job %s] batch %d/%d — SKIPPED (idempotent): %d rows already exist for ts [%s — %s]",
					jobID, offset/batchSize+1, (len(rows)+batchSize-1)/batchSize,
					existing, batch[0].Ts.Format(time.RFC3339), batch[len(batch)-1].Ts.Format(time.RFC3339))
				rowsDone += int64(len(batch)) // count as "done" even if skipped
				updateJobProgress(db, jobID, rowsDone)
				time.Sleep(workerBatchThrottle) // P5: throttle even on skip
				continue
			}
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			markJobFailed(db, jobID, fmt.Errorf("begin tx at offset %d: %w", offset, err))
			return err
		}

		n, err := BulkCopyTelemetryLog(tx, batch)
		if err != nil {
			_ = tx.Rollback()
			markJobFailed(db, jobID, fmt.Errorf("COPY at offset %d: %w", offset, err))
			return err
		}
		if err := tx.Commit(); err != nil {
			markJobFailed(db, jobID, fmt.Errorf("commit at offset %d: %w", offset, err))
			return err
		}

		rowsDone += n
		elapsed := time.Since(batchStart)
		rps := float64(n) / elapsed.Seconds()
		log.Printf("   [Job %s] batch %d/%d — %d rows in %s (%.0f rows/s)",
			jobID, offset/batchSize+1, (len(rows)+batchSize-1)/batchSize,
			n, elapsed.Round(time.Millisecond), rps)

		// Update progress.
		updateJobProgress(db, jobID, rowsDone)

		// P5: throttle between batches to avoid starving real-time ingest.
		// 50ms sleep gives real-time requests priority on the DB connection pool.
		time.Sleep(workerBatchThrottle)
	}

	totalElapsed := time.Since(start)
	log.Printf("✅ [ImportWorker] Job %s done: %d rows in %s (%.0f rows/s avg)",
		jobID, rowsDone, totalElapsed.Round(time.Second), float64(rowsDone)/totalElapsed.Seconds())

	markJobDone(db, jobID, rowsDone)
	return nil
}

// ─── DX HTTP job processor (stub for future expansion) ───────────────────────

func processDXHTTPJob(ctx context.Context, db *sql.DB, job *ImportJob) error {
	log.Printf("⚠️  [ImportWorker] Job %s: source_type=dx_http not yet implemented — marking failed", job.ID)
	return fmt.Errorf("dx_http import not yet implemented in this version")
}

// ─── Job status helpers ───────────────────────────────────────────────────────

func updateJobProgress(db *sql.DB, jobID uuid.UUID, rowsDone int64) {
	_, err := db.Exec(`
		UPDATE nxd.import_jobs SET rows_done = $1, updated_at = NOW() WHERE id = $2
	`, rowsDone, jobID)
	if err != nil {
		log.Printf("⚠️  [ImportWorker] Failed to update progress for job %s: %v", jobID, err)
	}
}

func markJobDone(db *sql.DB, jobID uuid.UUID, rowsDone int64) {
	_, err := db.Exec(`
		UPDATE nxd.import_jobs
		SET status = 'done', rows_done = $1, finished_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, rowsDone, jobID)
	if err != nil {
		log.Printf("⚠️  [ImportWorker] Failed to mark job %s done: %v", jobID, err)
	}
}

func markJobFailed(db *sql.DB, jobID uuid.UUID, cause error) {
	msg := cause.Error()
	log.Printf("❌ [ImportWorker] Job %s FAILED: %v", jobID, cause)
	_, err := db.Exec(`
		UPDATE nxd.import_jobs
		SET status = 'failed', error_message = $1, finished_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, msg, jobID)
	if err != nil {
		log.Printf("⚠️  [ImportWorker] Failed to mark job %s failed: %v", jobID, err)
	}
}

func markJobCancelled(db *sql.DB, jobID uuid.UUID, rowsDone int64) {
	log.Printf("🚫 [ImportWorker] Job %s CANCELLED at %d rows", jobID, rowsDone)
	_, err := db.Exec(`
		UPDATE nxd.import_jobs
		SET status = 'cancelled', rows_done = $1, finished_at = NOW(), updated_at = NOW()
		WHERE id = $2
	`, rowsDone, jobID)
	if err != nil {
		log.Printf("⚠️  [ImportWorker] Failed to mark job %s cancelled: %v", jobID, err)
	}
}

// countRangeRows counts existing telemetry_log rows for an asset in [tsStart, tsEnd].
// Used for idempotency check before inserting a historical batch.
// Returns 0 if no rows exist (safe to insert), >0 if data already present (skip).
func countRangeRows(db *sql.DB, assetID uuid.UUID, tsStart, tsEnd time.Time) (int64, error) {
	var n int64
	err := db.QueryRow(`
		SELECT COUNT(*) FROM nxd.telemetry_log
		WHERE asset_id = $1 AND ts >= $2 AND ts <= $3
		LIMIT 1
	`, assetID, tsStart, tsEnd).Scan(&n)
	return n, err
}

// isCancelled checks whether a job has been set to status='cancelled' by the admin.
// Called before each batch to support graceful mid-job cancellation.
func isCancelled(db *sql.DB, jobID uuid.UUID) (bool, error) {
	var status string
	err := db.QueryRow(`SELECT status FROM nxd.import_jobs WHERE id = $1`, jobID).Scan(&status)
	if err != nil {
		return false, err
	}
	return status == "cancelled", nil
}

// ─── CreateImportJob ──────────────────────────────────────────────────────────

// CreateImportJobParams are the parameters for creating a new import job.
type CreateImportJobParams struct {
	FactoryID   uuid.UUID
	AssetID     *uuid.UUID
	RequestedBy *uuid.UUID
	SourceType  string // "memory", "dx_http"
	SourceConfig json.RawMessage
	BatchSize   int
	PeriodStart *time.Time
	PeriodEnd   *time.Time
}

// CreateImportJob inserts a new job in status='pending' and returns its ID.
func CreateImportJob(db *sql.DB, p CreateImportJobParams) (uuid.UUID, error) {
	if p.BatchSize <= 0 {
		p.BatchSize = workerBatchSize
	}
	if p.SourceType == "" {
		p.SourceType = "memory"
	}
	cfgRaw := p.SourceConfig
	if len(cfgRaw) == 0 {
		cfgRaw = json.RawMessage("{}")
	}

	id := uuid.New()
	_, err := db.Exec(`
		INSERT INTO nxd.import_jobs
		    (id, factory_id, asset_id, requested_by, source_type, source_config, batch_size, period_start, period_end)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9)
	`, id, p.FactoryID, p.AssetID, p.RequestedBy, p.SourceType, string(cfgRaw),
		p.BatchSize, p.PeriodStart, p.PeriodEnd)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create import job: %w", err)
	}
	log.Printf("➕ [ImportWorker] Created job %s (factory=%s, type=%s)", id, p.FactoryID, p.SourceType)
	return id, nil
}

// GetImportJob returns a job by ID scoped to a factory.
func GetImportJob(db *sql.DB, jobID, factoryID uuid.UUID) (*ImportJobStatus, error) {
	return scanImportJobStatus(db.QueryRow(`
		SELECT id, factory_id, asset_id, status, source_type, rows_total, rows_done,
		       batch_size, error_message, started_at, finished_at, created_at, updated_at
		FROM nxd.import_jobs
		WHERE id = $1 AND factory_id = $2
	`, jobID, factoryID))
}

// ListImportJobs returns recent jobs for a factory.
func ListImportJobs(db *sql.DB, factoryID uuid.UUID, limit int) ([]ImportJobStatus, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.Query(`
		SELECT id, factory_id, asset_id, status, source_type, rows_total, rows_done,
		       batch_size, error_message, started_at, finished_at, created_at, updated_at
		FROM nxd.import_jobs
		WHERE factory_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, factoryID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []ImportJobStatus
	for rows.Next() {
		j, err := scanImportJobStatusRow(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *j)
	}
	return list, rows.Err()
}

// CancelImportJob sets a job to status='cancelled'. If the worker is mid-run,
// it will detect this on the next batch and stop gracefully.
func CancelImportJob(db *sql.DB, jobID, factoryID uuid.UUID) error {
	res, err := db.Exec(`
		UPDATE nxd.import_jobs
		SET status = 'cancelled', finished_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND factory_id = $2 AND status IN ('pending', 'running')
	`, jobID, factoryID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("job not found or already in terminal state")
	}
	log.Printf("🚫 [ImportWorker] Job %s cancelled via API", jobID)
	return nil
}

// RetryImportJob resets a failed/cancelled job back to pending so the worker
// will pick it up again.
func RetryImportJob(db *sql.DB, jobID, factoryID uuid.UUID) error {
	res, err := db.Exec(`
		UPDATE nxd.import_jobs
		SET status = 'pending', error_message = NULL,
		    rows_done = 0, started_at = NULL, finished_at = NULL, updated_at = NOW()
		WHERE id = $1 AND factory_id = $2 AND status IN ('failed', 'cancelled')
	`, jobID, factoryID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("job not found or not in failed/cancelled state")
	}
	log.Printf("🔁 [ImportWorker] Job %s reset to pending for retry", jobID)
	return nil
}

// ─── ImportJobStatus (API response shape) ────────────────────────────────────

type ImportJobStatus struct {
	ID          string     `json:"id"`
	FactoryID   string     `json:"factory_id"`
	AssetID     *string    `json:"asset_id,omitempty"`
	Status      string     `json:"status"`
	SourceType  string     `json:"source_type"`
	RowsTotal   int64      `json:"rows_total"`
	RowsDone    int64      `json:"rows_done"`
	ProgressPct float64    `json:"progress_pct"`
	BatchSize   int        `json:"batch_size"`
	ErrorMsg    *string    `json:"error_message,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func scanImportJobStatus(row *sql.Row) (*ImportJobStatus, error) {
	var j ImportJobStatus
	var assetID sql.NullString
	var errMsg sql.NullString
	var startedAt, finishedAt sql.NullTime
	err := row.Scan(
		&j.ID, &j.FactoryID, &assetID, &j.Status, &j.SourceType,
		&j.RowsTotal, &j.RowsDone, &j.BatchSize,
		&errMsg, &startedAt, &finishedAt, &j.CreatedAt, &j.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if assetID.Valid {
		j.AssetID = &assetID.String
	}
	if errMsg.Valid {
		j.ErrorMsg = &errMsg.String
	}
	if startedAt.Valid {
		j.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		j.FinishedAt = &finishedAt.Time
	}
	if j.RowsTotal > 0 {
		j.ProgressPct = float64(j.RowsDone) / float64(j.RowsTotal) * 100
	}
	return &j, nil
}

func scanImportJobStatusRow(rows *sql.Rows) (*ImportJobStatus, error) {
	var j ImportJobStatus
	var assetID sql.NullString
	var errMsg sql.NullString
	var startedAt, finishedAt sql.NullTime
	err := rows.Scan(
		&j.ID, &j.FactoryID, &assetID, &j.Status, &j.SourceType,
		&j.RowsTotal, &j.RowsDone, &j.BatchSize,
		&errMsg, &startedAt, &finishedAt, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if assetID.Valid {
		j.AssetID = &assetID.String
	}
	if errMsg.Valid {
		j.ErrorMsg = &errMsg.String
	}
	if startedAt.Valid {
		j.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		j.FinishedAt = &finishedAt.Time
	}
	if j.RowsTotal > 0 {
		j.ProgressPct = float64(j.RowsDone) / float64(j.RowsTotal) * 100
	}
	return &j, nil
}
