package core

import "time"

// User representa um usuário logado (Google)
type User struct {
	ID               int       `json:"id"`
	Email            string    `json:"email"`
	Name             string    `json:"name"`
	GoogleUID        string    `json:"-"` // não expor
	TermsAcceptedAt  *time.Time `json:"terms_accepted_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// Factory representa uma unidade/fábrica cadastrada
type Factory struct {
	ID        int       `json:"id"`
	UserID    *int      `json:"user_id,omitempty"` // dono da fábrica (1:1 por enquanto)
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key,omitempty"` // omitir em listagens; mostrar só na criação/regeneração
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

// Sector representa um setor/baia criado pelo usuário
type Sector struct {
	ID         int       `json:"id"`
	FactoryID  int       `json:"factory_id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
}

// Machine representa uma máquina conectada
type Machine struct {
	ID         int       `json:"id"`
	FactoryID  int       `json:"factory_id"`
	DeviceID   string    `json:"device_id"`
	Name       string    `json:"name"`
	Brand      string    `json:"brand"`      // Siemens, Delta, Mitsubishi
	Protocol   string    `json:"protocol"`   // S7, Modbus, etc
	LastSeen    time.Time `json:"last_seen"`
	Status      string    `json:"status"`       // online, offline, error
	CreatedAt   time.Time `json:"created_at"`
	DisplayName string    `json:"display_name"` // nome que o gestor dá
	Notes       string    `json:"notes"`
	SectorID    *int      `json:"sector_id,omitempty"`
	SectorName  string    `json:"sector_name,omitempty"`
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
