package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// FinancialAggregateResult — resultado por setor/linha para um período.
type FinancialAggregateResult struct {
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	SectorID        *uuid.UUID `json:"sector_id,omitempty"`
	SectorName      string    `json:"sector_name,omitempty"`
	OKCount         float64   `json:"ok_count"`
	NOKCount        float64   `json:"nok_count"`
	HoursParada     float64   `json:"hours_parada"`
	FaturamentoBruto float64  `json:"faturamento_bruto"`
	PerdaRefugo     float64   `json:"perda_refugo"`
	CustoParada     float64   `json:"custo_parada"`
	ValorVendaOk    float64   `json:"valor_venda_ok"`
	CustoRefugoUn   float64   `json:"custo_refugo_un"`
	CustoParadaH    float64   `json:"custo_parada_h"`
}

// AssetFinancialRow — breakdown por ativo.
type AssetFinancialRow struct {
	AssetID          uuid.UUID `json:"asset_id"`
	AssetName        string    `json:"asset_name"`
	OKCount          float64   `json:"ok_count"`
	NOKCount         float64   `json:"nok_count"`
	HoursParada      float64   `json:"hours_parada"`
	FaturamentoBruto float64   `json:"faturamento_bruto"`
	PerdaRefugo      float64   `json:"perda_refugo"`
	CustoParada      float64   `json:"custo_parada"`
}

// ComputeFinancialAggregate calcula OK/NOK/horas parada a partir da telemetria e aplica business_config.
// Se sectorID for nil, usa todos os ativos da fábrica e config padrão (sector_id null).
func ComputeFinancialAggregate(db *sql.DB, factoryID uuid.UUID, sectorID *uuid.UUID, periodStart, periodEnd time.Time) (*FinancialAggregateResult, []AssetFinancialRow, error) {
	config, err := GetBusinessConfigBySector(db, factoryID, sectorID)
	if err != nil || config == nil {
		return nil, nil, err
	}

	var assetIDs []uuid.UUID
	if sectorID != nil {
		assets, err := ListAssetsBySector(db, *sectorID)
		if err != nil {
			return nil, nil, err
		}
		for _, a := range assets {
			assetIDs = append(assetIDs, a.ID)
		}
	} else {
		rows, err := db.Query(`SELECT id FROM nxd.assets WHERE factory_id = $1`, factoryID)
		if err != nil {
			return nil, nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err != nil {
				return nil, nil, err
			}
			assetIDs = append(assetIDs, id)
		}
		if err := rows.Err(); err != nil {
			return nil, nil, err
		}
	}

	var sectorName string
	if sectorID != nil {
		db.QueryRow(`SELECT name FROM nxd.sectors WHERE id = $1`, sectorID).Scan(&sectorName)
	}

	var totalOK, totalNOK, totalHoursParada float64
	var breakdown []AssetFinancialRow
	for _, assetID := range assetIDs {
		mapping, err := GetTagMappingByAsset(db, assetID)
		if err != nil || mapping == nil {
			continue
		}
		okDelta, nokDelta, hoursParada := computeAssetDeltas(db, assetID, mapping, periodStart, periodEnd)
		totalOK += okDelta
		totalNOK += nokDelta
		totalHoursParada += hoursParada
		var assetName string
		db.QueryRow(`SELECT COALESCE(display_name, source_tag_id) FROM nxd.assets WHERE id = $1`, assetID).Scan(&assetName)
		breakdown = append(breakdown, AssetFinancialRow{
			AssetID:          assetID,
			AssetName:        assetName,
			OKCount:          okDelta,
			NOKCount:         nokDelta,
			HoursParada:      hoursParada,
			FaturamentoBruto: okDelta * config.ValorVendaOk,
			PerdaRefugo:      nokDelta * config.CustoRefugoUn,
			CustoParada:      hoursParada * config.CustoParadaH,
		})
	}

	res := &FinancialAggregateResult{
		PeriodStart:      periodStart,
		PeriodEnd:         periodEnd,
		SectorID:          sectorID,
		SectorName:        sectorName,
		OKCount:           totalOK,
		NOKCount:          totalNOK,
		HoursParada:       totalHoursParada,
		ValorVendaOk:      config.ValorVendaOk,
		CustoRefugoUn:     config.CustoRefugoUn,
		CustoParadaH:      config.CustoParadaH,
		FaturamentoBruto:  totalOK * config.ValorVendaOk,
		PerdaRefugo:       totalNOK * config.CustoRefugoUn,
		CustoParada:       totalHoursParada * config.CustoParadaH,
	}
	return res, breakdown, nil
}

// computeAssetDeltas retorna (delta OK, delta NOK, horas parada) para o ativo no período.
func computeAssetDeltas(db *sql.DB, assetID uuid.UUID, m *TagMappingRow, start, end time.Time) (okDelta, nokDelta, hoursParada float64) {
	if m.TagOK != "" {
		okDelta = metricDelta(db, assetID, m.TagOK, m.ReadingRule, start, end)
	}
	if m.TagNOK != "" {
		nokDelta = metricDelta(db, assetID, m.TagNOK, m.ReadingRule, start, end)
	}
	if m.TagStatus != "" {
		hoursParada = metricHoursParada(db, assetID, m.TagStatus, start, end)
	}
	return okDelta, nokDelta, hoursParada
}

func metricDelta(db *sql.DB, assetID uuid.UUID, metricKey, rule string, start, end time.Time) float64 {
	if rule == "absolute" {
		var last float64
		err := db.QueryRow(`
			SELECT metric_value FROM nxd.telemetry_log
			WHERE asset_id = $1 AND metric_key = $2 AND ts >= $3 AND ts <= $4
			ORDER BY ts DESC LIMIT 1
		`, assetID, metricKey, start, end).Scan(&last)
		if err != nil {
			return 0
		}
		return last
	}
	var first, last float64
	if err := db.QueryRow(`
		SELECT metric_value FROM nxd.telemetry_log
		WHERE asset_id = $1 AND metric_key = $2 AND ts >= $3 AND ts <= $4 ORDER BY ts ASC LIMIT 1
	`, assetID, metricKey, start, end).Scan(&first); err != nil {
		return 0
	}
	if err := db.QueryRow(`
		SELECT metric_value FROM nxd.telemetry_log
		WHERE asset_id = $1 AND metric_key = $2 AND ts >= $3 AND ts <= $4 ORDER BY ts DESC LIMIT 1
	`, assetID, metricKey, start, end).Scan(&last); err != nil {
		return 0
	}
	delta := last - first
	if delta < 0 {
		return 0
	}
	return delta
}

// metricHoursParada estima horas parada: soma intervalos onde tag_status = 0 (ou valor considerado "parado").
func metricHoursParada(db *sql.DB, assetID uuid.UUID, metricKey string, start, end time.Time) float64 {
	rows, err := db.Query(`
		SELECT ts, metric_value FROM nxd.telemetry_log
		WHERE asset_id = $1 AND metric_key = $2 AND ts >= $3 AND ts <= $4
		ORDER BY ts ASC
	`, assetID, metricKey, start, end)
	if err != nil {
		return 0
	}
	defer rows.Close()
	var prevTs time.Time
	var prevParada bool
	var sumSecs float64
	for rows.Next() {
		var ts time.Time
		var val float64
		if err := rows.Scan(&ts, &val); err != nil {
			continue
		}
		parada := val < 0.5 // 0 = parado, 1 = rodando
		if prevParada && parada {
			sumSecs += ts.Sub(prevTs).Seconds()
		}
		prevTs = ts
		prevParada = parada
	}
	return sumSecs / 3600
}
