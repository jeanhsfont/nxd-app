package data

import (
	"database/sql"
	"fmt"
	"hubsystem/core"
	"time"
)

// CreateFactory cria uma nova fábrica
func CreateFactory(name, apiKey string) (int, error) {
	result, err := DB.Exec(
		"INSERT INTO factories (name, api_key) VALUES (?, ?)",
		name, apiKey,
	)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return int(id), nil
}

// GetFactoryByAPIKey busca fábrica pela API key
func GetFactoryByAPIKey(apiKey string) (*core.Factory, error) {
	var f core.Factory
	err := DB.QueryRow(
		"SELECT id, name, api_key, created_at, is_active FROM factories WHERE api_key = ?",
		apiKey,
	).Scan(&f.ID, &f.Name, &f.APIKey, &f.CreatedAt, &f.IsActive)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

// GetOrCreateMachine busca ou cria uma máquina (auto-discovery)
func GetOrCreateMachine(factoryID int, deviceID, brand, protocol string) (*core.Machine, error) {
	// Tenta buscar primeiro
	var m core.Machine
	err := DB.QueryRow(
		`SELECT id, factory_id, device_id, name, brand, protocol, last_seen, status, created_at 
		FROM machines WHERE factory_id = ? AND device_id = ?`,
		factoryID, deviceID,
	).Scan(&m.ID, &m.FactoryID, &m.DeviceID, &m.Name, &m.Brand, &m.Protocol, &m.LastSeen, &m.Status, &m.CreatedAt)

	if err == nil {
		// Atualiza last_seen e status
		DB.Exec("UPDATE machines SET last_seen = ?, status = 'online' WHERE id = ?", time.Now(), m.ID)
		m.LastSeen = time.Now()
		m.Status = "online"
		return &m, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Cria nova máquina
	name := fmt.Sprintf("%s_%s", brand, deviceID)
	result, err := DB.Exec(
		`INSERT INTO machines (factory_id, device_id, name, brand, protocol, last_seen, status) 
		VALUES (?, ?, ?, ?, ?, ?, 'online')`,
		factoryID, deviceID, name, brand, protocol, time.Now(),
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	m.ID = int(id)
	m.FactoryID = factoryID
	m.DeviceID = deviceID
	m.Name = name
	m.Brand = brand
	m.Protocol = protocol
	m.LastSeen = time.Now()
	m.Status = "online"
	m.CreatedAt = time.Now()

	return &m, nil
}

// GetOrCreateTag busca ou cria uma tag (auto-discovery)
func GetOrCreateTag(machineID int, tagName, tagType string) (*core.Tag, error) {
	// Tenta buscar
	var t core.Tag
	var unit sql.NullString
	var minValue, maxValue sql.NullFloat64
	err := DB.QueryRow(
		`SELECT id, machine_id, tag_name, tag_type, unit, min_value, max_value, created_at, last_updated 
		FROM tags WHERE machine_id = ? AND tag_name = ?`,
		machineID, tagName,
	).Scan(&t.ID, &t.MachineID, &t.TagName, &t.TagType, &unit, &minValue, &maxValue, &t.CreatedAt, &t.LastUpdated)

	if err == nil {
		// Converte NullString/NullFloat64 para os tipos corretos
		if unit.Valid {
			t.Unit = unit.String
		}
		if minValue.Valid {
			val := minValue.Float64
			t.MinValue = &val
		}
		if maxValue.Valid {
			val := maxValue.Float64
			t.MaxValue = &val
		}
		// Atualiza last_updated
		DB.Exec("UPDATE tags SET last_updated = ? WHERE id = ?", time.Now(), t.ID)
		t.LastUpdated = time.Now()
		return &t, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Cria nova tag
	result, err := DB.Exec(
		`INSERT INTO tags (machine_id, tag_name, tag_type, last_updated) VALUES (?, ?, ?, ?)`,
		machineID, tagName, tagType, time.Now(),
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	t.ID = int(id)
	t.MachineID = machineID
	t.TagName = tagName
	t.TagType = tagType
	t.CreatedAt = time.Now()
	t.LastUpdated = time.Now()

	return &t, nil
}

// InsertDataPoint insere um ponto de dados
func InsertDataPoint(machineID, tagID int, value string, timestamp time.Time) error {
	_, err := DB.Exec(
		"INSERT INTO data_points (machine_id, tag_id, value, timestamp) VALUES (?, ?, ?, ?)",
		machineID, tagID, value, timestamp,
	)
	return err
}

// LogAudit registra log de auditoria
func LogAudit(action, apiKey, deviceID, status, message, ipAddress string) error {
	_, err := DB.Exec(
		`INSERT INTO audit_logs (action, api_key, device_id, status, message, ip_address) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		action, apiKey, deviceID, status, message, ipAddress,
	)
	return err
}

// GetMachinesByFactory retorna todas as máquinas de uma fábrica
func GetMachinesByFactory(factoryID int) ([]core.Machine, error) {
	rows, err := DB.Query(
		`SELECT id, factory_id, device_id, name, brand, protocol, last_seen, status, created_at 
		FROM machines WHERE factory_id = ? ORDER BY name`,
		factoryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var machines []core.Machine
	for rows.Next() {
		var m core.Machine
		if err := rows.Scan(&m.ID, &m.FactoryID, &m.DeviceID, &m.Name, &m.Brand, &m.Protocol, &m.LastSeen, &m.Status, &m.CreatedAt); err != nil {
			return nil, err
		}
		machines = append(machines, m)
	}
	return machines, nil
}

// GetTagsByMachine retorna todas as tags de uma máquina
func GetTagsByMachine(machineID int) ([]core.Tag, error) {
	rows, err := DB.Query(
		`SELECT id, machine_id, tag_name, tag_type, unit, min_value, max_value, created_at, last_updated 
		FROM tags WHERE machine_id = ? ORDER BY tag_name`,
		machineID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []core.Tag
	for rows.Next() {
		var t core.Tag
		var unit sql.NullString
		var minValue, maxValue sql.NullFloat64
		if err := rows.Scan(&t.ID, &t.MachineID, &t.TagName, &t.TagType, &unit, &minValue, &maxValue, &t.CreatedAt, &t.LastUpdated); err != nil {
			return nil, err
		}
		// Converte NullString/NullFloat64 para os tipos corretos
		if unit.Valid {
			t.Unit = unit.String
		}
		if minValue.Valid {
			val := minValue.Float64
			t.MinValue = &val
		}
		if maxValue.Valid {
			val := maxValue.Float64
			t.MaxValue = &val
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// DeleteMachine remove uma máquina e todos os dados relacionados
func DeleteMachine(machineID int) error {
	// Remove data_points
	_, err := DB.Exec("DELETE FROM data_points WHERE machine_id = ?", machineID)
	if err != nil {
		return err
	}
	// Remove tags
	_, err = DB.Exec("DELETE FROM tags WHERE machine_id = ?", machineID)
	if err != nil {
		return err
	}
	// Remove machine
	_, err = DB.Exec("DELETE FROM machines WHERE id = ?", machineID)
	return err
}

// GetLatestDataPoints retorna os últimos valores de todas as tags de uma máquina
func GetLatestDataPoints(machineID int) (map[string]string, error) {
	rows, err := DB.Query(
		`SELECT t.tag_name, dp.value 
		FROM data_points dp
		INNER JOIN tags t ON dp.tag_id = t.id
		WHERE dp.machine_id = ?
		AND dp.id IN (
			SELECT MAX(id) FROM data_points WHERE machine_id = ? GROUP BY tag_id
		)`,
		machineID, machineID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var tagName, value string
		if err := rows.Scan(&tagName, &value); err != nil {
			return nil, err
		}
		result[tagName] = value
	}
	return result, nil
}
