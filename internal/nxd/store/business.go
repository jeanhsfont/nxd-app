package store

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// BusinessConfigRow — configuração financeira por setor (ou fábrica quando sector_id é nil).
type BusinessConfigRow struct {
	ID            uuid.UUID `json:"id"`
	FactoryID     uuid.UUID `json:"factory_id"`
	SectorID      *uuid.UUID `json:"sector_id,omitempty"`
	ValorVendaOk  float64   `json:"valor_venda_ok"`
	CustoRefugoUn float64   `json:"custo_refugo_un"`
	CustoParadaH  float64   `json:"custo_parada_h"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TagMappingRow — mapeamento de tags do CLP por ativo.
type TagMappingRow struct {
	ID          uuid.UUID `json:"id"`
	AssetID     uuid.UUID `json:"asset_id"`
	TagOK       string   `json:"tag_ok"`
	TagNOK      string   `json:"tag_nok"`
	TagStatus   string   `json:"tag_status"`
	ReadingRule string   `json:"reading_rule"` // "delta" | "absolute"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetBusinessConfigBySector retorna a config do setor ou a config padrão da fábrica (sector_id nil).
// Se sectorID for nil (zero UUID), retorna apenas a config com sector_id IS NULL.
func GetBusinessConfigBySector(db *sql.DB, factoryID uuid.UUID, sectorID *uuid.UUID) (*BusinessConfigRow, error) {
	if sectorID != nil && *sectorID != uuid.Nil {
		r, err := getBusinessConfigRow(db, factoryID, sectorID)
		if err != nil || r != nil {
			return r, err
		}
	}
	return getBusinessConfigRow(db, factoryID, nil)
}

func getBusinessConfigRow(db *sql.DB, factoryID uuid.UUID, sectorID *uuid.UUID) (*BusinessConfigRow, error) {
	var r BusinessConfigRow
	var sectorIDNull sql.NullString
	var err error
	if sectorID == nil || *sectorID == uuid.Nil {
		err = db.QueryRow(`
			SELECT id, factory_id, sector_id, valor_venda_ok, custo_refugo_un, custo_parada_h, created_at, updated_at
			FROM nxd.business_config WHERE factory_id = $1 AND sector_id IS NULL LIMIT 1
		`, factoryID).Scan(&r.ID, &r.FactoryID, &sectorIDNull, &r.ValorVendaOk, &r.CustoRefugoUn, &r.CustoParadaH, &r.CreatedAt, &r.UpdatedAt)
	} else {
		err = db.QueryRow(`
			SELECT id, factory_id, sector_id, valor_venda_ok, custo_refugo_un, custo_parada_h, created_at, updated_at
			FROM nxd.business_config WHERE factory_id = $1 AND sector_id = $2 LIMIT 1
		`, factoryID, *sectorID).Scan(&r.ID, &r.FactoryID, &sectorIDNull, &r.ValorVendaOk, &r.CustoRefugoUn, &r.CustoParadaH, &r.CreatedAt, &r.UpdatedAt)
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if sectorIDNull.Valid {
		u, _ := uuid.Parse(sectorIDNull.String)
		r.SectorID = &u
	}
	return &r, nil
}

// ListBusinessConfigs retorna todas as configs da fábrica (por setor + padrão).
func ListBusinessConfigs(db *sql.DB, factoryID uuid.UUID) ([]BusinessConfigRow, error) {
	rows, err := db.Query(`
		SELECT id, factory_id, sector_id, valor_venda_ok, custo_refugo_un, custo_parada_h, created_at, updated_at
		FROM nxd.business_config WHERE factory_id = $1 ORDER BY sector_id NULLS LAST
	`, factoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []BusinessConfigRow
	for rows.Next() {
		var r BusinessConfigRow
		var sectorIDNull sql.NullString
		if err := rows.Scan(&r.ID, &r.FactoryID, &sectorIDNull, &r.ValorVendaOk, &r.CustoRefugoUn, &r.CustoParadaH, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		if sectorIDNull.Valid {
			u, _ := uuid.Parse(sectorIDNull.String)
			r.SectorID = &u
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// UpsertBusinessConfig insere ou atualiza config (por factory_id + sector_id).
// Para sector_id NULL (config padrão fábrica), usa UPDATE então INSERT por causa de UNIQUE NULLS.
func UpsertBusinessConfig(db *sql.DB, factoryID uuid.UUID, sectorID *uuid.UUID, valorVendaOk, custoRefugoUn, custoParadaH float64) (uuid.UUID, error) {
	var id uuid.UUID
	if sectorID == nil || *sectorID == uuid.Nil {
		res, err := db.Exec(`
			UPDATE nxd.business_config SET valor_venda_ok = $1, custo_refugo_un = $2, custo_parada_h = $3, updated_at = NOW()
			WHERE factory_id = $4 AND sector_id IS NULL
		`, valorVendaOk, custoRefugoUn, custoParadaH, factoryID)
		if err != nil {
			return uuid.Nil, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			err = db.QueryRow(`SELECT id FROM nxd.business_config WHERE factory_id = $1 AND sector_id IS NULL`, factoryID).Scan(&id)
			return id, err
		}
		err = db.QueryRow(`
			INSERT INTO nxd.business_config (factory_id, sector_id, valor_venda_ok, custo_refugo_un, custo_parada_h, updated_at)
			VALUES ($1, NULL, $2, $3, $4, NOW())
			RETURNING id
		`, factoryID, valorVendaOk, custoRefugoUn, custoParadaH).Scan(&id)
		return id, err
	}
	err := db.QueryRow(`
		INSERT INTO nxd.business_config (factory_id, sector_id, valor_venda_ok, custo_refugo_un, custo_parada_h, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (factory_id, sector_id) DO UPDATE SET
			valor_venda_ok = EXCLUDED.valor_venda_ok,
			custo_refugo_un = EXCLUDED.custo_refugo_un,
			custo_parada_h = EXCLUDED.custo_parada_h,
			updated_at = NOW()
		RETURNING id
	`, factoryID, *sectorID, valorVendaOk, custoRefugoUn, custoParadaH).Scan(&id)
	return id, err
}

// GetTagMappingByAsset retorna o mapeamento de tags do ativo.
func GetTagMappingByAsset(db *sql.DB, assetID uuid.UUID) (*TagMappingRow, error) {
	var r TagMappingRow
	err := db.QueryRow(`
		SELECT id, asset_id, COALESCE(tag_ok,''), COALESCE(tag_nok,''), COALESCE(tag_status,''), COALESCE(reading_rule,'delta'), created_at, updated_at
		FROM nxd.tag_mapping WHERE asset_id = $1
	`, assetID).Scan(&r.ID, &r.AssetID, &r.TagOK, &r.TagNOK, &r.TagStatus, &r.ReadingRule, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// ListTagMappingsByFactory retorna mapeamentos de todos os ativos da fábrica.
func ListTagMappingsByFactory(db *sql.DB, factoryID uuid.UUID) ([]TagMappingRow, error) {
	rows, err := db.Query(`
		SELECT t.id, t.asset_id, COALESCE(t.tag_ok,''), COALESCE(t.tag_nok,''), COALESCE(t.tag_status,''), COALESCE(t.reading_rule,'delta'), t.created_at, t.updated_at
		FROM nxd.tag_mapping t
		JOIN nxd.assets a ON a.id = t.asset_id
		WHERE a.factory_id = $1 ORDER BY a.display_name
	`, factoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []TagMappingRow
	for rows.Next() {
		var r TagMappingRow
		if err := rows.Scan(&r.ID, &r.AssetID, &r.TagOK, &r.TagNOK, &r.TagStatus, &r.ReadingRule, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// UpsertTagMapping insere ou atualiza mapeamento por asset_id.
func UpsertTagMapping(db *sql.DB, assetID uuid.UUID, tagOK, tagNOK, tagStatus, readingRule string) (uuid.UUID, error) {
	if readingRule == "" {
		readingRule = "delta"
	}
	var id uuid.UUID
	err := db.QueryRow(`
		INSERT INTO nxd.tag_mapping (asset_id, tag_ok, tag_nok, tag_status, reading_rule, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (asset_id) DO UPDATE SET
			tag_ok = EXCLUDED.tag_ok,
			tag_nok = EXCLUDED.tag_nok,
			tag_status = EXCLUDED.tag_status,
			reading_rule = EXCLUDED.reading_rule,
			updated_at = NOW()
		RETURNING id
	`, assetID, ptrOrNull(tagOK), ptrOrNull(tagNOK), ptrOrNull(tagStatus), readingRule).Scan(&id)
	return id, err
}

func ptrOrNull(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
