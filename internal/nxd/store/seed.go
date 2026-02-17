package store

import (
	"database/sql"
	"encoding/json"
)

// SeedReportTemplates inserts default report templates if the table is empty. Idempotent.
func SeedReportTemplates(db *sql.DB) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM nxd.report_templates").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	defaults := []struct {
		Category             string
		Name                 string
		Description          string
		PromptInstructions   string
		OutputSchemaVersion  string
	}{
		{"Producao", "Resumo de Produção Diária", "Produção total, peças, paradas e eficiência do dia.", "Gere um resumo executivo com KPIs de produção e paradas.", "1"},
		{"Producao", "OEE por Setor", "OEE (Overall Equipment Effectiveness) por setor no período.", "Calcule e apresente OEE com disponibilidade, desempenho e qualidade.", "1"},
		{"Producao", "Paradas e Causas", "Análise de paradas com duração e causas raiz.", "Liste paradas, duração e indique causas quando houver dados.", "1"},
		{"Financeiro", "Lucro Cessante", "Estimativa de lucro cessante por paradas no período.", "Use apenas custo/hora e tempo parado configurados; marque INSUFICIENTE se faltar.", "1"},
		{"Financeiro", "Custo Energia vs Produção", "Consumo de energia e custo versus peças produzidas.", "Relação energia/produção; marque dados faltantes em missing_data.", "1"},
		{"Qualidade", "Refugo e Não Conformidades", "Volume de refugo e eventos de qualidade no período.", "Refugo e NC quando houver métricas; senão missing_data.", "1"},
		{"Manutencao", "Saúde dos Ativos", "Status e alertas de manutenção por máquina/setor.", "Health score e alertas; recomendações apenas com evidência.", "1"},
		{"Manutencao", "Tendência de Falhas", "Tendência de falhas e avisos ao longo do tempo.", "Séries de falhas/avisos; sem inventar causas.", "1"},
		{"Estrategia", "Visão Executiva 30 dias", "Resumo para diretoria: produção, paradas, principais achados.", "Máximo 7 bullets; riscos e premissas; missing_data explícito.", "1"},
		{"Estrategia", "Comparativo Setores", "Comparativo de desempenho entre setores.", "Compare apenas métricas disponíveis; evidências em evidence_refs.", "1"},
	}
	for _, t := range defaults {
		df, _ := json.Marshal(map[string]interface{}{"period": "7d", "detail": "medio"})
		_, err := db.Exec(
			`INSERT INTO nxd.report_templates (category, name, description, default_filters, prompt_instructions, output_schema_version)
			 VALUES ($1, $2, $3, $4::jsonb, $5, $6)`,
			t.Category, t.Name, t.Description, df, t.PromptInstructions, t.OutputSchemaVersion,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
