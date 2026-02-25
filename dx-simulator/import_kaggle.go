//go:build ignore

// Script de importação do dataset Kaggle "Pump Sensor Data"
// Uso: go run import_kaggle.go
// Requisito: DATABASE_URL deve estar configurado, ou rodar com a variável de ambiente
// Ex: DATABASE_URL="postgres://..." go run import_kaggle.go

package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	csvFile     = "kaggle-data/sensor.csv"
	maxRows     = 5000 // importa as primeiras 5000 linhas (~83h de dados a 1min/linha)
	factoryName = "Bomba Industrial - Dataset Kaggle"
)

// 4 CLPs simulados mapeados de grupos de sensores
var clpGroups = map[string][]string{
	"CLP-BOMBA-01": {"sensor_00", "sensor_01", "sensor_02", "sensor_03", "sensor_04", "sensor_05", "sensor_06", "sensor_07", "sensor_08", "sensor_09", "sensor_10", "sensor_11", "sensor_12"},
	"CLP-BOMBA-02": {"sensor_13", "sensor_14", "sensor_15", "sensor_16", "sensor_17", "sensor_18", "sensor_19", "sensor_20", "sensor_21", "sensor_22", "sensor_23", "sensor_24", "sensor_25"},
	"CLP-BOMBA-03": {"sensor_26", "sensor_27", "sensor_28", "sensor_29", "sensor_30", "sensor_31", "sensor_32", "sensor_33", "sensor_34", "sensor_35", "sensor_36", "sensor_37", "sensor_38"},
	"CLP-BOMBA-04": {"sensor_39", "sensor_40", "sensor_41", "sensor_42", "sensor_43", "sensor_44", "sensor_45", "sensor_46", "sensor_47", "sensor_48", "sensor_49", "sensor_50", "sensor_51"},
}

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL não definido. Ex: DATABASE_URL='postgres://user:pass@host/db' go run import_kaggle.go")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Erro ao conectar: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Erro de conexão: %v", err)
	}
	log.Println("✓ Conectado ao PostgreSQL")

	// Busca ou cria factory demo
	var factoryID string
	err = db.QueryRow(`SELECT id FROM nxd.factories WHERE name = $1 LIMIT 1`, factoryName).Scan(&factoryID)
	if err == sql.ErrNoRows {
		err = db.QueryRow(`INSERT INTO nxd.factories (name, is_active) VALUES ($1, true) RETURNING id`, factoryName).Scan(&factoryID)
		if err != nil {
			log.Fatalf("Erro ao criar factory: %v", err)
		}
		log.Printf("✓ Factory criada: %s (%s)", factoryName, factoryID)
	} else if err != nil {
		log.Fatalf("Erro ao buscar factory: %v", err)
	} else {
		log.Printf("✓ Factory existente: %s (%s)", factoryName, factoryID)
	}

	// Cria/busca assets (CLPs)
	assetIDs := map[string]string{}
	for clpName := range clpGroups {
		var assetID string
		err = db.QueryRow(`SELECT id FROM nxd.assets WHERE factory_id = $1 AND source_tag_id = $2`, factoryID, clpName).Scan(&assetID)
		if err == sql.ErrNoRows {
			err = db.QueryRow(`
				INSERT INTO nxd.assets (factory_id, source_tag_id, display_name)
				VALUES ($1, $2, $3) RETURNING id
			`, factoryID, clpName, clpName).Scan(&assetID)
			if err != nil {
				log.Fatalf("Erro ao criar asset %s: %v", clpName, err)
			}
		} else if err != nil {
			log.Fatalf("Erro ao buscar asset: %v", err)
		}
		assetIDs[clpName] = assetID
		log.Printf("  ✓ Asset: %s (%s)", clpName, assetID)
	}

	// Lê o CSV
	f, err := os.Open(csvFile)
	if err != nil {
		log.Fatalf("Erro ao abrir CSV: %v", err)
	}
	defer f.Close()

	reader := csv.NewReader(bufio.NewReader(f))
	headers, err := reader.Read()
	if err != nil {
		log.Fatalf("Erro ao ler header: %v", err)
	}

	// Índices de colunas
	colIdx := map[string]int{}
	for i, h := range headers {
		colIdx[strings.TrimSpace(h)] = i
	}
	tsIdx := colIdx["timestamp"]
	statusIdx := colIdx["machine_status"]

	log.Printf("✓ CSV aberto. Importando até %d linhas...", maxRows)

	tx, _ := db.Begin()
	stmt, err := tx.Prepare(`
		INSERT INTO nxd.telemetry_log (ts, factory_id, asset_id, metric_key, metric_value, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		log.Fatalf("Erro ao preparar statement: %v", err)
	}

	rowCount := 0
	insertCount := 0

	for {
		if rowCount >= maxRows {
			break
		}
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		// Parse timestamp
		tsStr := strings.TrimSpace(record[tsIdx])
		ts, err := time.Parse("2006-01-02 15:04:05", tsStr)
		if err != nil {
			continue
		}

		machineStatus := "OK"
		if statusIdx < len(record) {
			s := strings.TrimSpace(record[statusIdx])
			if s == "BROKEN" || s == "RECOVERING" {
				machineStatus = s
			}
		}

		// Insere métricas para cada CLP
		for clpName, sensors := range clpGroups {
			assetID := assetIDs[clpName]
			for _, sensor := range sensors {
				idx, ok := colIdx[sensor]
				if !ok || idx >= len(record) {
					continue
				}
				valStr := strings.TrimSpace(record[idx])
				if valStr == "" {
					continue
				}
				val, err := strconv.ParseFloat(valStr, 64)
				if err != nil || math.IsNaN(val) || math.IsInf(val, 0) {
					continue
				}
				_, err = stmt.Exec(ts, factoryID, assetID, sensor, val, machineStatus)
				if err == nil {
					insertCount++
				}
			}
		}

		rowCount++
		if rowCount%500 == 0 {
			log.Printf("  ... %d linhas processadas, %d registros inseridos", rowCount, insertCount)
			tx.Commit()
			tx, _ = db.Begin()
			stmt, _ = tx.Prepare(`
				INSERT INTO nxd.telemetry_log (ts, factory_id, asset_id, metric_key, metric_value, status)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT DO NOTHING
			`)
		}
	}

	tx.Commit()
	stmt.Close()

	// Atualiza asset_metric_catalog
	log.Println("✓ Atualizando catálogo de métricas...")
	db.Exec(`
		INSERT INTO nxd.asset_metric_catalog (factory_id, asset_id, metric_key, first_seen, last_seen)
		SELECT factory_id, asset_id, metric_key, MIN(ts), MAX(ts)
		FROM nxd.telemetry_log
		WHERE factory_id = $1
		GROUP BY factory_id, asset_id, metric_key
		ON CONFLICT (factory_id, asset_id, metric_key) DO UPDATE
		SET last_seen = EXCLUDED.last_seen
	`, factoryID)

	log.Printf("\n✅ Importação concluída!")
	log.Printf("   Factory: %s", factoryName)
	log.Printf("   Linhas CSV: %d", rowCount)
	log.Printf("   Registros inseridos: %d", insertCount)
	log.Printf("   CLPs: %d assets criados", len(assetIDs))
}
