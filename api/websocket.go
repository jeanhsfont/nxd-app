package api

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Permite todas as origens (ajustar em produção)
	},
}

// WebSocketHandler gerencia conexões WebSocket para updates em tempo real
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Erro ao fazer upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Println("🔌 Nova conexão WebSocket estabelecida")

	// Mantém conexão aberta
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("🔌 Conexão WebSocket fechada")
			break
		}
	}
}
