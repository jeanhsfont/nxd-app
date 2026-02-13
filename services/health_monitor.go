package services

import (
	"fmt"
	"hubsystem/data"
	"log"
	"time"
)

// Configurações do monitor de saúde
const (
	// Tempo sem dados para considerar máquina offline
	OFFLINE_THRESHOLD = 2 * time.Minute
	
	// Tempo sem dados para considerar máquina crítica (alerta)
	CRITICAL_THRESHOLD = 5 * time.Minute
	
	// Intervalo de verificação
	CHECK_INTERVAL = 30 * time.Second
)

// MachineHealthStatus representa o status de saúde de uma máquina
type MachineHealthStatus struct {
	MachineID     int       `json:"machine_id"`
	MachineName   string    `json:"machine_name"`
	FactoryID     int       `json:"factory_id"`
	Status        string    `json:"status"` // "online", "offline", "critical"
	LastSeen      time.Time `json:"last_seen"`
	SilentMinutes float64   `json:"silent_minutes"`
	Message       string    `json:"message"`
}

// ConnectionLog representa um log de conexão/desconexão
type ConnectionLog struct {
	Timestamp   time.Time `json:"timestamp"`
	MachineID   int       `json:"machine_id"`
	MachineName string    `json:"machine_name"`
	Event       string    `json:"event"` // "connected", "disconnected", "data_received"
	Details     string    `json:"details"`
}

var connectionLogs []ConnectionLog

// StartHealthMonitor inicia o monitoramento de saúde das máquinas
func StartHealthMonitor() {
	log.Println("🏥 Monitor de Saúde iniciado")
	log.Printf("   • Threshold Offline: %v", OFFLINE_THRESHOLD)
	log.Printf("   • Threshold Crítico: %v", CRITICAL_THRESHOLD)
	log.Printf("   • Intervalo de Check: %v", CHECK_INTERVAL)
	
	go healthCheckLoop()
}

func healthCheckLoop() {
	ticker := time.NewTicker(CHECK_INTERVAL)
	defer ticker.Stop()
	
	for range ticker.C {
		checkAllMachines()
	}
}

func checkAllMachines() {
	// Busca todas as máquinas
	rows, err := data.DB.Query(`
		SELECT m.id, m.name, m.factory_id, m.last_seen, m.status, f.name as factory_name
		FROM machines m
		JOIN factories f ON m.factory_id = f.id
	`)
	if err != nil {
		log.Printf("❌ [HEALTH] Erro ao buscar máquinas: %v", err)
		return
	}
	defer rows.Close()
	
	now := time.Now()
	
	for rows.Next() {
		var machineID, factoryID int
		var machineName, currentStatus, factoryName string
		var lastSeen time.Time
		
		if err := rows.Scan(&machineID, &machineName, &factoryID, &lastSeen, &currentStatus, &factoryName); err != nil {
			continue
		}
		
		silentDuration := now.Sub(lastSeen)
		var newStatus string
		
		if silentDuration > CRITICAL_THRESHOLD {
			newStatus = "critical"
		} else if silentDuration > OFFLINE_THRESHOLD {
			newStatus = "offline"
		} else {
			newStatus = "online"
		}
		
		// Se o status mudou, loga a mudança
		if currentStatus != newStatus {
			logConnectionEvent(machineID, machineName, newStatus, silentDuration)
			
			// Atualiza status no banco
			data.DB.Exec("UPDATE machines SET status = ? WHERE id = ?", newStatus, machineID)
			
			if newStatus == "critical" {
				log.Printf("🚨 [HEALTH] CRÍTICO: %s (%s) - Sem dados há %.1f minutos",
					machineName, factoryName, silentDuration.Minutes())
			} else if newStatus == "offline" {
				log.Printf("⚠️  [HEALTH] OFFLINE: %s (%s) - Sem dados há %.1f minutos",
					machineName, factoryName, silentDuration.Minutes())
			} else if newStatus == "online" && currentStatus != "online" {
				log.Printf("✅ [HEALTH] RECONECTADO: %s (%s)",
					machineName, factoryName)
			}
		}
	}
}

func logConnectionEvent(machineID int, machineName, status string, silentDuration time.Duration) {
	event := "status_change"
	details := ""
	
	switch status {
	case "online":
		event = "reconnected"
		details = "Máquina voltou a enviar dados"
	case "offline":
		event = "disconnected"
		details = "Máquina parou de enviar dados há " + silentDuration.String()
	case "critical":
		event = "critical_alert"
		details = "ATENÇÃO: Máquina sem comunicação há " + silentDuration.String()
	}
	
	logEntry := ConnectionLog{
		Timestamp:   time.Now(),
		MachineID:   machineID,
		MachineName: machineName,
		Event:       event,
		Details:     details,
	}
	
	// Mantém últimos 1000 logs em memória
	connectionLogs = append(connectionLogs, logEntry)
	if len(connectionLogs) > 1000 {
		connectionLogs = connectionLogs[1:]
	}
}

// LogDataReceived registra quando dados são recebidos (chamado pelo IngestHandler)
func LogDataReceived(machineID int, machineName string, tagsCount int) {
	logEntry := ConnectionLog{
		Timestamp:   time.Now(),
		MachineID:   machineID,
		MachineName: machineName,
		Event:       "data_received",
		Details:     fmt.Sprintf("Recebido pacote com %d tags", tagsCount),
	}
	
	connectionLogs = append(connectionLogs, logEntry)
	if len(connectionLogs) > 1000 {
		connectionLogs = connectionLogs[1:]
	}
}

// GetConnectionLogs retorna os logs de conexão
func GetConnectionLogs(limit int) []ConnectionLog {
	if limit <= 0 || limit > len(connectionLogs) {
		limit = len(connectionLogs)
	}
	
	// Retorna os mais recentes
	start := len(connectionLogs) - limit
	if start < 0 {
		start = 0
	}
	
	return connectionLogs[start:]
}

// GetMachineHealthStatus retorna o status de saúde de todas as máquinas de uma fábrica
func GetMachineHealthStatus(factoryID int) ([]MachineHealthStatus, error) {
	rows, err := data.DB.Query(`
		SELECT id, name, last_seen, status
		FROM machines
		WHERE factory_id = ?
	`, factoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var statuses []MachineHealthStatus
	now := time.Now()
	
	for rows.Next() {
		var mh MachineHealthStatus
		var lastSeen time.Time
		
		if err := rows.Scan(&mh.MachineID, &mh.MachineName, &lastSeen, &mh.Status); err != nil {
			continue
		}
		
		mh.FactoryID = factoryID
		mh.LastSeen = lastSeen
		mh.SilentMinutes = now.Sub(lastSeen).Minutes()
		
		switch mh.Status {
		case "online":
			mh.Message = "✅ Comunicação normal"
		case "offline":
			mh.Message = "⚠️ Sem dados há " + formatDuration(now.Sub(lastSeen))
		case "critical":
			mh.Message = "🚨 CRÍTICO: Sem comunicação há " + formatDuration(now.Sub(lastSeen))
		}
		
		statuses = append(statuses, mh)
	}
	
	return statuses, nil
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "menos de 1 minuto"
	} else if d < time.Hour {
		return fmt.Sprintf("%.0f minutos", d.Minutes())
	} else {
		return fmt.Sprintf("%.1f horas", d.Hours())
	}
}
