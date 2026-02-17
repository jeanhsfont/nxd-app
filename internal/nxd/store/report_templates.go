package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// ReportTemplateRow is a row from nxd.report_templates.
type ReportTemplateRow struct {
	ID                  uuid.UUID `json:"id"`
	Category            string    `json:"category"`
	Name                string    `json:"name"`
	Description         string    `json:"description,omitempty"`
	DefaultFilters      []byte    `json:"-"` // JSON
	PromptInstructions  string    `json:"prompt_instructions,omitempty"`
	OutputSchemaVersion string    `json:"output_schema_version"`
	CreatedAt           time.Time `json:"created_at"`
}

// ListReportTemplates returns all report templates (for accordion by category).
func ListReportTemplates(db *sql.DB) ([]ReportTemplateRow, error) {
	rows, err := db.Query(
		`SELECT id, category, name, description, default_filters, prompt_instructions, output_schema_version, created_at FROM nxd.report_templates ORDER BY category, name`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ReportTemplateRow
	for rows.Next() {
		var r ReportTemplateRow
		var desc, prompt sql.NullString
		if err := rows.Scan(&r.ID, &r.Category, &r.Name, &desc, &r.DefaultFilters, &prompt, &r.OutputSchemaVersion, &r.CreatedAt); err != nil {
			return nil, err
		}
		if desc.Valid {
			r.Description = desc.String
		}
		if prompt.Valid {
			r.PromptInstructions = prompt.String
		}
		list = append(list, r)
	}
	return list, rows.Err()
}
