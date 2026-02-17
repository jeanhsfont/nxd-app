package store

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ReportRunRow represents a row from nxd.report_runs.
type ReportRunRow struct {
	ID             uuid.UUID       `json:"id"`
	FactoryID      uuid.UUID       `json:"factory_id"`
	RequestedBy    string          `json:"requested_by,omitempty"`
	Filters        json.RawMessage `json:"filters,omitempty"`
	PromptContract json.RawMessage `json:"prompt_contract,omitempty"`
	Status         string          `json:"status"`
	ResultJSON     json.RawMessage `json:"result_json,omitempty"`
	ExportURL      string          `json:"export_url,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// CreateReportRun inserts a report run (status PENDING) and returns id.
func CreateReportRun(db *sql.DB, factoryID uuid.UUID, requestedBy string, filters, promptContract json.RawMessage) (uuid.UUID, error) {
	id := uuid.New()
	_, err := db.Exec(
		`INSERT INTO nxd.report_runs (id, factory_id, requested_by, filters, prompt_contract, status) VALUES ($1, $2, $3, $4, $5, 'PENDING')`,
		id, factoryID, requestedBy, filters, promptContract,
	)
	return id, err
}

// GetReportRun returns a report run by id and factory_id.
func GetReportRun(db *sql.DB, id, factoryID uuid.UUID) (*ReportRunRow, error) {
	var r ReportRunRow
	var reqBy, exportURL sql.NullString
	var filters, contract, result []byte
	err := db.QueryRow(
		`SELECT id, factory_id, requested_by, filters, prompt_contract, status, result_json, export_url, created_at, updated_at FROM nxd.report_runs WHERE id = $1 AND factory_id = $2`,
		id, factoryID,
	).Scan(&r.ID, &r.FactoryID, &reqBy, &filters, &contract, &r.Status, &result, &exportURL, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if reqBy.Valid {
		r.RequestedBy = reqBy.String
	}
	if exportURL.Valid {
		r.ExportURL = exportURL.String
	}
	r.Filters = filters
	r.PromptContract = contract
	r.ResultJSON = result
	return &r, nil
}

// UpdateReportRunStatus updates status and optionally result_json and export_url.
func UpdateReportRunStatus(db *sql.DB, id, factoryID uuid.UUID, status string, resultJSON json.RawMessage, exportURL string) error {
	if resultJSON != nil {
		_, err := db.Exec(
			`UPDATE nxd.report_runs SET status = $1, result_json = $2, export_url = NULLIF($3,''), updated_at = NOW() WHERE id = $4 AND factory_id = $5`,
			status, resultJSON, exportURL, id, factoryID,
		)
		return err
	}
	_, err := db.Exec(
		`UPDATE nxd.report_runs SET status = $1, updated_at = NOW() WHERE id = $2 AND factory_id = $3`,
		status, id, factoryID,
	)
	return err
}
