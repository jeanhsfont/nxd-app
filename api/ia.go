package api

import (
	"context"
	"encoding/json"
	"fmt"
	"hubsystem/internal/nxd/store"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/genai"
)

// ChatRequest é o payload do POST /api/ia/chat
type ChatRequest struct {
	Message  string `json:"message"`
	SectorID string `json:"sector_id,omitempty"` // UUID do setor (filtro opcional)
}

// ChatResponse é o retorno do chat
type ChatResponse struct {
	Reply   string `json:"reply"`
	Sources string `json:"sources,omitempty"`
}

// IAChatHandler — POST /api/ia/chat
// Recebe uma pergunta, monta contexto de telemetria real do banco, chama Gemini e retorna resposta
func IAChatHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int64)
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Message) == "" {
		http.Error(w, "Mensagem obrigatória", http.StatusBadRequest)
		return
	}

	// Monta contexto de telemetria
	ctx := r.Context()
	telemetryContext, err := buildTelemetryContext(userID, req.SectorID)
	if err != nil {
		log.Printf("[IA] Erro ao montar contexto: %v", err)
		telemetryContext = "Sem dados de telemetria disponíveis no momento."
	}

	// Chama Gemini
	reply, err := callGemini(ctx, req.Message, telemetryContext)
	if err != nil {
		log.Printf("[IA] Erro Gemini: %v", err)
		http.Error(w, "Erro ao consultar IA: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ChatResponse{Reply: reply})
}

// buildTelemetryContext monta um resumo dos dados reais do banco para o prompt da IA
func buildTelemetryContext(userID int64, sectorID string) (string, error) {
	factoryID, err := getFactoryIDForUser(userID)
	if err != nil {
		return "", err
	}

	nxdDB := store.NXDDB()
	if nxdDB == nil {
		return "", fmt.Errorf("NXD DB não disponível")
	}

	var sb strings.Builder
	sb.WriteString("=== DADOS DE TELEMETRIA EM TEMPO REAL ===\n")
	sb.WriteString(fmt.Sprintf("Fábrica: %s | Timestamp: %s\n\n", factoryID, time.Now().Format("02/01/2006 15:04:05")))

	// Se sectorID fornecido, filtra assets do setor
	sectorFilter := ""
	sectorArgs := []interface{}{factoryID}
	if sectorID != "" {
		sectorFilter = "AND a.group_id = $2"
		sectorArgs = append(sectorArgs, sectorID)

		// Nome do setor
		var sectorName string
		nxdDB.QueryRow(`SELECT name FROM nxd.sectors WHERE id = $1`, sectorID).Scan(&sectorName)
		if sectorName != "" {
			sb.WriteString(fmt.Sprintf("Filtro de setor: %s\n\n", sectorName))
		}
	}

	// Busca assets com última telemetria
	query := fmt.Sprintf(`
		SELECT
			a.id, a.display_name, a.source_tag_id,
			MAX(t.ts) as last_seen,
			COUNT(t.ts) as reading_count
		FROM nxd.assets a
		LEFT JOIN nxd.telemetry_log t ON t.asset_id = a.id
		WHERE a.factory_id = $1 %s
		GROUP BY a.id, a.display_name, a.source_tag_id
		ORDER BY a.display_name
		LIMIT 20
	`, sectorFilter)

	rows, err := nxdDB.Query(query, sectorArgs...)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	type assetInfo struct {
		id          string
		displayName string
		sourceTagID string
		lastSeen    *time.Time
		readingCount int
	}
	var assets []assetInfo

	for rows.Next() {
		var a assetInfo
		var lastSeen *time.Time
		var ls *time.Time
		if err := rows.Scan(&a.id, &a.displayName, &a.sourceTagID, &ls, &a.readingCount); err == nil {
			lastSeen = ls
			a.lastSeen = lastSeen
			assets = append(assets, a)
		}
	}
	rows.Close()

	if len(assets) == 0 {
		sb.WriteString("Nenhum ativo encontrado para este filtro.\n")
		return sb.String(), nil
	}

	// Para cada asset, busca as últimas métricas
	for _, a := range assets {
		now := time.Now()
		isOnline := false
		if a.lastSeen != nil {
			isOnline = now.Sub(*a.lastSeen) < 60*time.Second
		}
		status := "OFFLINE"
		if isOnline {
			status = "ONLINE"
		}
		lastSeenStr := "nunca"
		if a.lastSeen != nil {
			lastSeenStr = a.lastSeen.Format("15:04:05")
		}

		sb.WriteString(fmt.Sprintf("--- CLP: %s (%s) | Status: %s | Último dado: %s | Leituras: %d ---\n",
			a.displayName, a.sourceTagID, status, lastSeenStr, a.readingCount))

		// Últimas métricas
		metricRows, err := nxdDB.Query(`
			SELECT DISTINCT ON (metric_key) metric_key, metric_value, ts
			FROM nxd.telemetry_log
			WHERE asset_id = $1
			ORDER BY metric_key, ts DESC
			LIMIT 15
		`, a.id)
		if err == nil {
			for metricRows.Next() {
				var key string
				var val float64
				var ts time.Time
				if err := metricRows.Scan(&key, &val, &ts); err == nil {
					sb.WriteString(fmt.Sprintf("  %s = %.4f\n", key, val))
				}
			}
			metricRows.Close()
		}

		// Estatísticas das últimas 5 minutos
		statsRows, err := nxdDB.Query(`
			SELECT metric_key,
				   AVG(metric_value) as avg_val,
				   MIN(metric_value) as min_val,
				   MAX(metric_value) as max_val,
				   COUNT(*) as samples
			FROM nxd.telemetry_log
			WHERE asset_id = $1
			  AND ts >= NOW() - INTERVAL '5 minutes'
			GROUP BY metric_key
		`, a.id)
		if err == nil {
			hasStats := false
			for statsRows.Next() {
				if !hasStats {
					sb.WriteString("  [Estatísticas últimos 5 min]\n")
					hasStats = true
				}
				var key string
				var avg, min, max float64
				var samples int
				if err := statsRows.Scan(&key, &avg, &min, &max, &samples); err == nil {
					sb.WriteString(fmt.Sprintf("    %s: avg=%.3f min=%.3f max=%.3f samples=%d\n",
						key, avg, min, max, samples))
				}
			}
			statsRows.Close()
		}
		sb.WriteString("\n")
	}

	// Indicadores financeiros (se config e tag_mapping existirem) — para a IA responder perguntas de perda/faturamento/custo parada
	var sectorUUID *uuid.UUID
	if sectorID != "" {
		if u, err := uuid.Parse(sectorID); err == nil {
			sectorUUID = &u
		}
	}
	now := time.Now()
	for _, period := range []struct{ label string; start time.Time }{
		{"24h", now.Add(-24 * time.Hour)},
		{"7d", now.Add(-7 * 24 * time.Hour)},
	} {
		res, _, err := store.ComputeFinancialAggregate(nxdDB, factoryID, sectorUUID, period.start, now)
		if err != nil || res == nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("=== INDICADORES FINANCEIROS (%s) ===\n", period.label))
		sb.WriteString(fmt.Sprintf("Peças OK: %.0f | Refugo: %.0f | Horas parada: %.2f\n",
			res.OKCount, res.NOKCount, res.HoursParada))
		sb.WriteString(fmt.Sprintf("Faturamento bruto: R$ %.2f | Perda refugo: R$ %.2f | Custo parada: R$ %.2f\n\n",
			res.FaturamentoBruto, res.PerdaRefugo, res.CustoParada))
	}

	return sb.String(), nil
}

// callGemini envia o prompt para o Gemini via Vertex AI e retorna a resposta
func callGemini(ctx context.Context, userMessage, telemetryContext string) (string, error) {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "slideflow-prod"
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  projectID,
		Location: "us-central1",
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return "", fmt.Errorf("erro ao criar cliente Gemini: %w", err)
	}

	systemPrompt := `Você é o NXD Intelligence, um assistente especializado em análise de dados industriais e telemetria de CLPs (Controladores Lógicos Programáveis).

Você tem acesso aos dados em tempo real da fábrica do usuário. Use esses dados para responder com precisão e objetividade.

Diretrizes:
- Responda em português brasileiro
- Seja direto e objetivo, foque nos dados relevantes
- Quando identificar anomalias (ex: Fault_Code > 0, Health_Score < 0.7), destaque-as
- Forneça insights acionáveis, não apenas descrições
- Use os dados de telemetria fornecidos como fonte primária
- Se não houver dados suficientes para responder, diga isso claramente
- Formate números com 2 casas decimais quando relevante
- Máximo de 3 parágrafos na resposta, seja conciso`

	fullPrompt := fmt.Sprintf("%s\n\n%s\n\n=== PERGUNTA DO USUÁRIO ===\n%s",
		systemPrompt, telemetryContext, userMessage)

	result, err := client.Models.GenerateContent(ctx,
		"gemini-2.0-flash-001",
		genai.Text(fullPrompt),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("erro na geração: %w", err)
	}

	if result == nil || len(result.Candidates) == 0 {
		return "", fmt.Errorf("resposta vazia do modelo")
	}

	return strings.TrimSpace(result.Text()), nil
}
