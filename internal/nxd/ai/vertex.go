package ai

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"google.golang.org/genai"
)

const defaultVertexLocation = "us-central1"
const reportModel = "gemini-1.5-flash-002"

// VertexConfig holds Vertex AI client configuration from env.
type VertexConfig struct {
	Project  string
	Location string
}

// IsConfigured returns true when Vertex is enabled via NXD_IA_PROVIDER=vertex and VERTEX_AI_PROJECT.
func IsConfigured() bool {
	if os.Getenv("NXD_IA_PROVIDER") != "vertex" {
		return false
	}
	return os.Getenv("VERTEX_AI_PROJECT") != ""
}

// GetConfig returns project and location from env (empty if not configured).
func GetConfig() (project, location string) {
	if !IsConfigured() {
		return "", ""
	}
	project = os.Getenv("VERTEX_AI_PROJECT")
	location = os.Getenv("VERTEX_AI_LOCATION")
	if location == "" {
		location = defaultVertexLocation
	}
	return project, location
}

// NewClient creates a Vertex AI genai client. Returns nil if not configured or on error.
func NewClient(ctx context.Context) (*genai.Client, error) {
	project, location := GetConfig()
	if project == "" {
		return nil, nil
	}
	return genai.NewClient(ctx, &genai.ClientConfig{
		Project:  project,
		Location: location,
		Backend:  genai.BackendVertexAI,
	})
}

// ReportOutputSchema is the expected JSON shape for report result (output_schema_version 1).
type ReportOutputSchema struct {
	Title                 string                   `json:"title"`
	SummaryBullets        []string                 `json:"summary_bullets"`
	KPIs                  []map[string]interface{} `json:"kpis"`
	Findings              []map[string]interface{} `json:"findings"`
	Charts                []map[string]interface{} `json:"charts"`
	RisksAndAssumptions   string                   `json:"risks_and_assumptions"`
	MissingData           []string                 `json:"missing_data"`
	Auditability          map[string]interface{}   `json:"auditability"`
}

// GenerateReport calls Vertex Gemini with the given prompt and returns a result suitable for result_json.
// The model is instructed to respond with JSON matching ReportOutputSchema. If the response is not
// valid JSON or the call fails, returns nil and the error.
func GenerateReport(ctx context.Context, client *genai.Client, prompt string) (resultJSON []byte, err error) {
	if client == nil {
		return nil, nil
	}
	systemInstruction := `Você é um analista de dados industriais. Responda APENAS com um único objeto JSON válido, sem markdown e sem texto antes ou depois.
O objeto deve ter exatamente: "title" (string), "summary_bullets" (array de strings), "kpis" (array de objetos), "findings" (array de objetos), "charts" (array de objetos), "risks_and_assumptions" (string), "missing_data" (array de strings), "auditability" (objeto com data_window, sources, rollup_used).
Não invente dados; use apenas o contexto fornecido. Se faltar informação, preencha missing_data e deixe campos vazios onde apropriado.`

	contents := []*genai.Content{
		{Role: genai.RoleUser, Parts: []*genai.Part{genai.NewPartFromText(prompt)}},
	}
	resp, err := client.Models.GenerateContent(ctx, reportModel, contents, &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{genai.NewPartFromText(systemInstruction)}},
		Temperature:       genai.Ptr(float32(0.2)),
		MaxOutputTokens:   4096,
	})
	if err != nil {
		return nil, err
	}
	text := resp.Text()
	text = strings.TrimSpace(text)
	// Remove optional markdown code block
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
	}
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
	}
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return nil, err
	}
	return json.Marshal(parsed)
}
