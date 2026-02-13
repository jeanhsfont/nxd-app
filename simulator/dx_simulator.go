package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// Configuração do simulador
const (
	SEND_INTERVAL = 2 * time.Second
)

var (
	NXD_ENDPOINT = "http://localhost:8080/api/ingest"
	API_KEY      = ""
)

// CLPSimulator simula um CLP específico
type CLPSimulator struct {
	Brand    string
	Protocol string
	DeviceID string
}

// NetworkSimulator simula condições de rede 4G/LTE
type NetworkSimulator struct {
	MinLatency    time.Duration
	MaxLatency    time.Duration
	DropRate      float64 // 0.0 a 1.0 (0% a 100%)
	RetryAttempts int
}

// Payload que será enviado ao NXD
type Payload struct {
	APIKey    string                 `json:"api_key"`
	DeviceID  string                 `json:"device_id"`
	Brand     string                 `json:"brand"`
	Protocol  string                 `json:"protocol"`
	Timestamp time.Time              `json:"timestamp"`
	Tags      map[string]interface{} `json:"tags"`
}

var (
	apiKey  string
	network = NetworkSimulator{
		MinLatency:    50 * time.Millisecond,
		MaxLatency:    500 * time.Millisecond,
		DropRate:      0.05, // 5% de perda
		RetryAttempts: 3,
	}
)

func main() {
	log.Println("🔧 Iniciando Simulador DX (Delta Gateway)...")

	// Lê endpoint do ambiente (se definido)
	if endpoint := os.Getenv("NXD_ENDPOINT"); endpoint != "" {
		NXD_ENDPOINT = endpoint
	}
	// Mantém compatibilidade com variável antiga
	if endpoint := os.Getenv("HUB_ENDPOINT"); endpoint != "" {
		NXD_ENDPOINT = endpoint
	}

	// Lê API Key dos argumentos ou ambiente
	if len(os.Args) >= 2 {
		API_KEY = os.Args[1]
	} else if key := os.Getenv("API_KEY"); key != "" {
		API_KEY = key
	} else {
		log.Fatal("❌ Uso: dx_simulator.exe <API_KEY> ou defina API_KEY no ambiente")
	}
	apiKey = API_KEY

	// Cria simuladores de CLPs
	clps := []CLPSimulator{
		{
			Brand:    "Siemens",
			Protocol: "S7",
			DeviceID: "CLP_SIEMENS_01",
		},
		{
			Brand:    "Delta",
			Protocol: "Modbus",
			DeviceID: "CLP_DELTA_01",
		},
	}

	log.Printf("✓ Conectado ao NXD: %s", NXD_ENDPOINT)
	log.Printf("✓ API Key: %s...", apiKey[:20])
	log.Printf("✓ Simulando %d CLPs", len(clps))
	log.Printf("✓ Condições de rede: Latência %v-%v, Perda %v%%", 
		network.MinLatency, network.MaxLatency, network.DropRate*100)

	// Loop infinito enviando dados
	ticker := time.NewTicker(SEND_INTERVAL)
	defer ticker.Stop()

	for range ticker.C {
		for _, clp := range clps {
			go sendData(clp)
		}
	}
}

// sendData envia dados de um CLP para o NXD
func sendData(clp CLPSimulator) {
	// Gera dados simulados
	tags := generateTags(clp.Brand)

	payload := Payload{
		APIKey:    apiKey,
		DeviceID:  clp.DeviceID,
		Brand:     clp.Brand,
		Protocol:  clp.Protocol,
		Timestamp: time.Now(),
		Tags:      tags,
	}

	// Serializa payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("❌ Erro ao serializar payload: %v", err)
		return
	}

	// Simula latência de rede
	latency := network.MinLatency + time.Duration(rand.Int63n(int64(network.MaxLatency-network.MinLatency)))
	time.Sleep(latency)

	// Simula perda de pacote
	if rand.Float64() < network.DropRate {
		log.Printf("📡 [%s] Pacote perdido (simulação de rede instável)", clp.DeviceID)
		return
	}

	// Tenta enviar com retry
	var lastErr error
	for attempt := 1; attempt <= network.RetryAttempts; attempt++ {
		resp, err := http.Post(NXD_ENDPOINT, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = err
			if attempt < network.RetryAttempts {
				log.Printf("⚠️  [%s] Tentativa %d falhou, retentando...", clp.DeviceID, attempt)
				time.Sleep(time.Second * time.Duration(attempt))
				continue
			}
			break
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Printf("✓ [%s] Dados enviados (latência: %v, tags: %d)", 
				clp.DeviceID, latency, len(tags))
			return
		}

		lastErr = fmt.Errorf("status code: %d", resp.StatusCode)
		if attempt < network.RetryAttempts {
			log.Printf("⚠️  [%s] Status %d, retentando...", clp.DeviceID, resp.StatusCode)
			time.Sleep(time.Second * time.Duration(attempt))
		}
	}

	log.Printf("❌ [%s] Falha após %d tentativas: %v", clp.DeviceID, network.RetryAttempts, lastErr)
}

// generateTags gera tags simuladas baseadas na marca do CLP
func generateTags(brand string) map[string]interface{} {
	tags := make(map[string]interface{})

	if brand == "Siemens" {
		tags["Temperatura_Motor"] = 45.0 + rand.Float64()*20.0
		tags["Pressao_Hidraulica"] = 120.0 + rand.Float64()*30.0
		tags["Velocidade_RPM"] = 1500.0 + rand.Float64()*500.0
		tags["Status_Producao"] = rand.Intn(2) == 1
		tags["Contador_Pecas"] = rand.Intn(10000)
		tags["Alarme_Temperatura"] = rand.Float64() > 0.9
	} else if brand == "Delta" {
		tags["Temp_Ambiente"] = 20.0 + rand.Float64()*15.0
		tags["Corrente_Motor_A"] = 5.0 + rand.Float64()*10.0
		tags["Tensao_Rede_V"] = 220.0 + rand.Float64()*10.0
		tags["Ciclos_Completos"] = rand.Intn(5000)
		tags["Modo_Operacao"] = []string{"AUTO", "MANUAL", "SETUP"}[rand.Intn(3)]
		tags["Falha_Comunicacao"] = rand.Float64() > 0.95
	}

	return tags
}
