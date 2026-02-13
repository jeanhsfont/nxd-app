package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateAPIKey gera uma chave API única e segura
func GenerateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("erro ao gerar API key: %w", err)
	}
	return "NXD_" + hex.EncodeToString(bytes), nil
}

// ValidateAPIKey verifica se a chave tem formato válido
func ValidateAPIKey(key string) bool {
	if len(key) < 10 {
		return false
	}
	if len(key) != 68 { // NXD_ + 64 chars hex
		return false
	}
	// Aceita tanto NXD_ quanto HUB_ (compatibilidade)
	return key[:4] == "NXD_" || key[:4] == "HUB_"
}

// SanitizeDeviceID limpa o device ID para evitar SQL injection
func SanitizeDeviceID(deviceID string) string {
	// Remove caracteres perigosos
	safe := ""
	for _, char := range deviceID {
		if (char >= 'a' && char <= 'z') || 
		   (char >= 'A' && char <= 'Z') || 
		   (char >= '0' && char <= '9') || 
		   char == '-' || char == '_' {
			safe += string(char)
		}
	}
	return safe
}
