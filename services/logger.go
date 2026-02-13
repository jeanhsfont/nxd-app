package services

import (
	"fmt"
	"hubsystem/data"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var logFile *os.File

// InitLogger inicializa o sistema de logs
func InitLogger() error {
	// Cria diretório de logs
	if err := os.MkdirAll("./logs", 0755); err != nil {
		return fmt.Errorf("erro ao criar diretório de logs: %w", err)
	}

	// Cria arquivo de log com timestamp
	filename := fmt.Sprintf("./logs/hub_%s.log", time.Now().Format("2006-01-02"))
	var err error
	logFile, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de log: %w", err)
	}

	// Configura log para escrever no arquivo E no console
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Println("✓ Sistema de logs inicializado")
	return nil
}

// CloseLogger fecha o arquivo de log
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// LogSuccess registra sucesso no banco de auditoria
func LogSuccess(action, apiKey, deviceID, message, ipAddress string) {
	data.LogAudit(action, apiKey, deviceID, "success", message, ipAddress)
}

// LogError registra erro no banco de auditoria
func LogError(action, apiKey, deviceID, message, ipAddress string) {
	data.LogAudit(action, apiKey, deviceID, "fail", message, ipAddress)
	log.Printf("❌ [%s] %s", action, message)
}

// GetLogFiles retorna lista de arquivos de log
func GetLogFiles() ([]string, error) {
	files, err := filepath.Glob("./logs/*.log")
	if err != nil {
		return nil, err
	}
	return files, nil
}
