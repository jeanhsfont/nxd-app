package main

import (
	"hubsystem/api"
	"hubsystem/data"
	"hubsystem/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	log.Println("🚀 Iniciando NXD (Nexus Data Exchange)...")

	// Inicializa logger
	if err := services.InitLogger(); err != nil {
		log.Fatalf("❌ Erro ao inicializar logger: %v", err)
	}
	defer services.CloseLogger()

	// Inicializa banco de dados
	if err := data.InitDatabase(); err != nil {
		log.Fatalf("❌ Erro ao inicializar banco: %v", err)
	}
	defer data.Close()

	// Executa migrações
	if err := data.RunMigrations(); err != nil {
		log.Fatalf("❌ Erro ao executar migrações: %v", err)
	}

	// Configura rotas
	router := mux.NewRouter()

	// API Routes
	router.HandleFunc("/api/health", api.HealthHandler).Methods("GET")
	router.HandleFunc("/api/ingest", api.IngestHandler).Methods("POST")
	router.HandleFunc("/api/factory/create", api.CreateFactoryHandler).Methods("POST")
	router.HandleFunc("/api/dashboard", api.GetDashboardHandler).Methods("GET")
	router.HandleFunc("/api/analytics", api.AnalyticsHandler).Methods("GET")
	router.HandleFunc("/api/machine/delete", api.DeleteMachineHandler).Methods("DELETE")
	router.HandleFunc("/ws", api.WebSocketHandler)

	// Serve arquivos estáticos (Dashboard)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))

	// Configura CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	// Inicia servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("✓ Servidor rodando em http://localhost:%s", port)
	log.Println("✓ Endpoints disponíveis:")
	log.Println("  - POST /api/ingest (Recebe dados do DX)")
	log.Println("  - POST /api/factory/create (Cria nova fábrica)")
	log.Println("  - GET  /api/dashboard?api_key=XXX (Dashboard)")
	log.Println("  - GET  /api/health (Health check)")
	log.Println("  - WS   /ws (WebSocket)")

	// Captura sinais de interrupção
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("\n🛑 Encerrando servidor...")
		os.Exit(0)
	}()

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("❌ Erro ao iniciar servidor: %v", err)
	}
}
