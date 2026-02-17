package main

import (
	"hubsystem/api"
	"hubsystem/data"
	"hubsystem/internal/nxd/handler"
	nxdmid "hubsystem/internal/nxd/middleware"
	nxdstore "hubsystem/internal/nxd/store"
	"hubsystem/services"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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

	// NXD v2: Postgres (opcional; rotas /nxd/* só montadas se NXD_DATABASE_URL estiver definido)
	if err := nxdstore.InitNXDDB(); err != nil {
		log.Printf("⚠️ NXD Postgres init: %v", err)
	}
	defer nxdstore.CloseNXDDB()

	// Configura rotas
	router := mux.NewRouter()

	// Rotas /nxd/* (add-only; não altera /api/*)
	if nxdstore.NXDDB() != nil {
		nxdRouter := router.PathPrefix("/nxd").Subrouter()
		nxdRouter.HandleFunc("/ready", handler.Ready).Methods("GET")
		nxdRouter.HandleFunc("/factories", nxdmid.RequireAuth(handler.ListFactories)).Methods("GET")
		nxdRouter.HandleFunc("/factories", nxdmid.RequireAuth(handler.CreateFactory)).Methods("POST")
		nxdRouter.HandleFunc("/groups", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.ListGroups))).Methods("GET")
		nxdRouter.HandleFunc("/groups", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.CreateGroup))).Methods("POST")
		nxdRouter.HandleFunc("/groups/{id}", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.UpdateGroup))).Methods("PATCH")
		nxdRouter.HandleFunc("/assets", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.ListAssets))).Methods("GET")
		nxdRouter.HandleFunc("/assets", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.CreateAsset))).Methods("POST")
		nxdRouter.HandleFunc("/assets/{id}", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.UpdateAsset))).Methods("PATCH")
		nxdRouter.HandleFunc("/assets/{id}/move", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.MoveAsset))).Methods("POST")
		nxdRouter.HandleFunc("/factories/{id}/gateway-key", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.GenerateGatewayKey))).Methods("POST")
		nxdRouter.HandleFunc("/telemetry/ingest", handler.TelemetryIngest).Methods("POST")
		nxdRouter.HandleFunc("/health", handler.Health).Methods("GET")
		nxdRouter.HandleFunc("/report-templates", handler.ListReportTemplates).Methods("GET")
		nxdRouter.HandleFunc("/reports/run", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.RunReport))).Methods("POST")
		nxdRouter.HandleFunc("/reports/{id}", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.GetReport))).Methods("GET")
		nxdRouter.HandleFunc("/reports/{id}/export-pdf", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.ExportPDF))).Methods("POST")
		nxdRouter.HandleFunc("/audit", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.AuditLog))).Methods("GET")
		nxdRouter.HandleFunc("/jobs/rollup", handler.RollupJob).Methods("POST")
		nxdRouter.HandleFunc("/jobs/alerts", handler.AlertsJob).Methods("POST")
		nxdRouter.HandleFunc("/alert-rules", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.ListAlertRules))).Methods("GET")
		nxdRouter.HandleFunc("/alert-rules", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.CreateAlertRule))).Methods("POST")
		nxdRouter.HandleFunc("/alerts", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.ListAlerts))).Methods("GET")
		nxdRouter.HandleFunc("/alerts/{id}/ack", nxdmid.RequireAuth(nxdmid.RequireFactory(handler.AckAlert))).Methods("POST")
		log.Println("✓ NXD: rotas /nxd/* montadas")
	}

	// Inicia monitor de saúde das máquinas
	services.StartHealthMonitor()

	// API Routes
	router.HandleFunc("/api/health", api.HealthHandler).Methods("GET")
	router.HandleFunc("/api/system", api.SystemParamsHandler).Methods("GET")
	router.HandleFunc("/api/ingest", api.IngestHandler).Methods("POST")
	router.HandleFunc("/api/factory/create", api.CreateFactoryHandler).Methods("POST")
	// Auth (Google OAuth2)
	router.HandleFunc("/api/auth/google/login", api.GoogleLoginHandler).Methods("GET")
	router.HandleFunc("/api/auth/google/callback", api.GoogleCallbackHandler).Methods("GET")
	router.HandleFunc("/api/auth/me", api.AuthMeHandler).Methods("GET")
	// Fábrica (autenticado)
	router.HandleFunc("/api/factory", api.RequireAuth(api.CreateFactoryAuthHandler)).Methods("POST")
	router.HandleFunc("/api/factory/regenerate-api-key", api.RequireAuth(api.RegenerateAPIKeyHandler)).Methods("POST")
	router.HandleFunc("/api/sectors", api.GetSectorsHandler).Methods("GET")
	router.HandleFunc("/api/sectors", api.CreateSectorHandler).Methods("POST")
	router.HandleFunc("/api/machine/asset", api.UpdateMachineAssetHandler).Methods("PUT")
	router.HandleFunc("/api/dashboard", api.GetDashboardHandler).Methods("GET")
	router.HandleFunc("/api/dashboard/summary", api.DashboardSummaryHandler).Methods("GET")
	router.HandleFunc("/api/report/ia", api.ReportIAHandler).Methods("POST")
	router.HandleFunc("/api/analytics", api.AnalyticsHandler).Methods("GET")
	router.HandleFunc("/api/machine/delete", api.DeleteMachineHandler).Methods("DELETE")
	router.HandleFunc("/api/connection/status", api.HealthStatusHandler).Methods("GET")
	router.HandleFunc("/api/connection/logs", api.ConnectionLogsHandler).Methods("GET")
	router.HandleFunc("/ws", api.WebSocketHandler)

	// Serve arquivos estáticos (SPA React ou Dashboard Legado)
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verifica se existe build do React em ./dist
		if _, err := os.Stat("./dist/index.html"); err == nil {
			path := filepath.Join("dist", r.URL.Path)
			// Se arquivo não existe e não é /api, serve index.html (SPA)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				http.ServeFile(w, r, "./dist/index.html")
			} else {
				http.FileServer(http.Dir("./dist")).ServeHTTP(w, r)
			}
			return
		}
		// Fallback para legado ./web
		http.FileServer(http.Dir("./web")).ServeHTTP(w, r)
	})

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
	log.Println("  - GET  /api/analytics?api_key=XXX (Analytics)")
	log.Println("  - GET  /api/connection/status?api_key=XXX (Status de Conexão)")
	log.Println("  - GET  /api/connection/logs?api_key=XXX (Logs de Conexão)")
	log.Println("  - GET  /api/health (Health check)")
	log.Println("  - GET  /api/system (Parâmetros do sistema, ex.: ia_operando_100)")
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
