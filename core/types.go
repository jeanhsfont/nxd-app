package core

import "time"

// User representa um usuário logado
type User struct {
	ID              string    `json:"id"`
	Email           string    `json:"email"`
	Name            string    `json:"name"`
	PasswordHash    string    `json:"-"` // não expor
	TermsAcceptedAt *time.Time `json:"terms_accepted_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Factory representa uma unidade/fábrica cadastrada
type Factory struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	APIKey    string    `json:"api_key,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}

// Sector representa um setor/baia criado pelo usuário
type Sector struct {
	ID        string    `json:"id"`
	FactoryID string    `json:"factory_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Machine representa uma máquina conectada
type Machine struct {
	ID          string    `json:"id"`
	FactoryID   string    `json:"factory_id"`
	DeviceID    string    `json:"device_id"`
	Name        string    `json:"name"`
	Brand       string    `json:"brand"`
	Protocol    string    `json:"protocol"`
	LastSeen    time.Time `json:"last_seen"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	DisplayName string    `json:"display_name"`
	Notes       string    `json:"notes"`
	SectorID    *string   `json:"sector_id,omitempty"`
	SectorName  string    `json:"sector_name,omitempty"`
}

// Tag representa um ponto de dados (auto-discovery)
type Tag struct {
	ID          string    `json:"id"`
	MachineID   string    `json:"machine_id"`
	TagName     string    `json:"tag_name"`
	TagType     string    `json:"tag_type"`
	Unit        string    `json:"unit"`
	MinValue    *float64  `json:"min_value"`
	MaxValue    *float64  `json:"max_value"`
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_updated"`
}

// DataPoint representa um valor recebido
type DataPoint struct {
	ID        string    `json:"id"`
	MachineID string    `json:"machine_id"`
	TagID     string    `json:"tag_id"`
	Value     string    `json:"value"`
	Timestamp time.Time `json:"timestamp"`
	Quality   string    `json:"quality"`
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
