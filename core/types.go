package core

import "time"

// Factory representa uma unidade/fábrica cadastrada
type Factory struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

// Machine representa uma máquina conectada
type Machine struct {
	ID         int       `json:"id"`
	FactoryID  int       `json:"factory_id"`
	DeviceID   string    `json:"device_id"`
	Name       string    `json:"name"`
	Brand      string    `json:"brand"`      // Siemens, Delta, Mitsubishi
	Protocol   string    `json:"protocol"`   // S7, Modbus, etc
	LastSeen   time.Time `json:"last_seen"`
	Status     string    `json:"status"`     // online, offline, error
	CreatedAt  time.Time `json:"created_at"`
}

// Tag representa um ponto de dados (auto-discovery)
type Tag struct {
	ID          int       `json:"id"`
	MachineID   int       `json:"machine_id"`
	TagName     string    `json:"tag_name"`
	TagType     string    `json:"tag_type"`     // float, int, bool, string
	Unit        string    `json:"unit"`         // °C, bar, rpm, etc
	MinValue    *float64  `json:"min_value"`
	MaxValue    *float64  `json:"max_value"`
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_updated"`
}

// DataPoint representa um valor recebido
type DataPoint struct {
	ID        int64     `json:"id"`
	MachineID int       `json:"machine_id"`
	TagID     int       `json:"tag_id"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Quality   string    `json:"quality"` // good, bad, uncertain
}

// IngestPayload é o formato que o DX envia
type IngestPayload struct {
	APIKey    string                 `json:"api_key"`
	DeviceID  string                 `json:"device_id"`
	Brand     string                 `json:"brand"`
	Protocol  string                 `json:"protocol"`
	Timestamp time.Time              `json:"timestamp"`
	Tags      map[string]interface{} `json:"tags"`
}

// Alert representa um alerta configurado
type Alert struct {
	ID        int       `json:"id"`
	TagID     int       `json:"tag_id"`
	Condition string    `json:"condition"` // >, <, ==, !=
	Threshold float64   `json:"threshold"`
	Message   string    `json:"message"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// AuditLog representa log de auditoria
type AuditLog struct {
	ID        int64     `json:"id"`
	Action    string    `json:"action"`
	APIKey    string    `json:"api_key"`
	DeviceID  string    `json:"device_id"`
	Status    string    `json:"status"` // success, fail
	Message   string    `json:"message"`
	IPAddress string    `json:"ip_address"`
	Timestamp time.Time `json:"timestamp"`
}
