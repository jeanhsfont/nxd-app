
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// Config define a estrutura do arquivo de configuração config.json
type Config struct {
	APIKey      string   `json:"api_key"`
	Endpoint    string   `json:"endpoint"`
	CLPs        []CLP    `json:"clps"`
	UpdateRateS int      `json:"update_rate_s"`
}

// CLP define a configuração de um CLP simulado
type CLP struct {
	DeviceID string `json:"device_id"`
	Brand    string `json:"brand"`
	Protocol string `json:"protocol"`
}

// IngestPayload é a estrutura de dados enviada para o NXD
type IngestPayload struct {
	APIKey    string                 `json:"api_key"`
	DeviceID  string                 `json:"device_id"`
	Brand     string                 `json:"brand"`
	Protocol  string                 `json:"protocol"`
	Timestamp time.Time              `json:"timestamp"`
	Tags      map[string]interface{} `json:"tags"`
}

func main() {
	fmt.Println("DX Simulator - Iniciando...")

	// Carrega configuração
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Printf("Erro ao carregar config.json: %v\n", err)
		return
	}

	fmt.Printf("Configuração carregada: %d CLPs, enviando para %s a cada %d segundos.\n", len(config.CLPs), config.Endpoint, config.UpdateRateS)

	// Ticker para controlar a frequência de envio
	ticker := time.NewTicker(time.Duration(config.UpdateRateS) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, clp := range config.CLPs {
			go sendData(clp, config)
		}
	}
}

// loadConfig carrega a configuração do simulador
func loadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(file, &config)
	return &config, err
}

// sendData gera dados simulados e os envia para o endpoint do NXD
func sendData(clp CLP, config *Config) {
	payload := IngestPayload{
		APIKey:    config.APIKey,
		DeviceID:  clp.DeviceID,
		Brand:     clp.Brand,
		Protocol:  clp.Protocol,
		Timestamp: time.Now(),
		Tags:      generateTags(clp.Brand),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("[%s] Erro ao serializar payload: %v\n", clp.DeviceID, err)
		return
	}

	resp, err := http.Post(config.Endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("[%s] Erro ao enviar dados: %v\n", clp.DeviceID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("✅ [%s] Dados enviados com sucesso!\n", clp.DeviceID)
	} else {
		fmt.Printf("❌ [%s] Falha ao enviar dados. Status: %s\n", clp.DeviceID, resp.Status)
	}
}

// generateTags cria um conjunto de tags com valores aleatórios
func generateTags(brand string) map[string]interface{} {
	rand.Seed(time.Now().UnixNano())
	status := rand.Intn(2) == 1
	pecas := rand.Intn(1000) + 5000
	tempMolde := 60 + rand.Float64()*20

	tags := map[string]interface{}{
		"Status_Producao":     status,
		"Total_Pecas":         pecas,
		"Temperatura_Molde":   tempMolde,
		"Consumo_Energia_kWh": 1500 + rand.Float64()*500,
		"Health_Score":        0.85 + rand.Float64()*0.14,
		"Custo_Hora_Parada":   200 + rand.Float64()*50,
	}

	// Adiciona tags específicas da marca
	switch brand {
	case "Siemens":
		tags["Cycle_Time_ms"] = 850 + rand.Intn(150)
	case "Rockwell":
		tags["Fault_Code"] = 0
		if !status {
			tags["Fault_Code"] = 100 + rand.Intn(50)
		}
	}

	return tags
}
