package services

import (
	"log"
)

// BroadcastUpdate envia notificação de atualização via WebSocket
func BroadcastUpdate(machineID int) {
	// TODO: Implementar broadcast real quando tivermos pool de conexões
	log.Printf("📡 Broadcast: Máquina %d atualizada", machineID)
}
