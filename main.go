package main

import (
	"context"
	"hubsystem/api"
	"hubsystem/api/middleware"
	"hubsystem/internal/nxd/config"
	"hubsystem/internal/nxd/store"

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
	log.SetOutput(os.Stdout)
	log.Println("Application entrypoint reached. Starting server...")
	api.InitAuth() // Carrega o segredo JWT
	config.Load()
	log.Println("🚀 Iniciando NXD (Nexus Data Exchange)...")

	// Inicializa o banco de dados da API legada (SQLite ou Postgres via DATABASE_URL)
	if os.Getenv("DATABASE_URL") == "" {
		log.Println("ℹ️  DATABASE_URL not set. Initializing legacy API database with local SQLite.")
	} else {
		log.Println("ℹ️  DATABASE_URL is set. Initializing legacy API database with PostgreSQL.")
	}
	if err := api.InitDB(); err != nil {
		log.Fatalf("❌ Erro ao inicializar banco de dados da API: %v", err)
	}

	// Inicializa o banco de dados NXD (schema nxd.*)
	if err := store.InitNXDDB(); err != nil {
		log.Printf("⚠️  Erro ao inicializar banco NXD (store): %v — sistema continua sem NXD store", err)
	} else {
		log.Println("✓ Banco de dados NXD (store) inicializado.")
	}

	// Inicia o worker de importação de dados históricos ("Download Longo").
	// Roda como goroutine; será interrompido quando o contexto for cancelado.
	workerCtx, workerCancel := context.WithCancel(context.Background())
	if nxdDB := store.NXDDB(); nxdDB != nil {
		go store.RunImportWorker(workerCtx, nxdDB)
		log.Println("✓ Worker de importação histórica iniciado.")
	} else {
		log.Println("⚠️  NXD DB indisponível — worker de importação não iniciado.")
		workerCancel() // cancel imediatamente se não há DB
		_ = workerCancel
	}
	defer workerCancel()

	// Configura rotas
	router := mux.NewRouter()

	// API Routes
	router.HandleFunc("/api/health", api.HealthHandler).Methods("GET")
	router.HandleFunc("/api/system", api.SystemParamsHandler).Methods("GET")
	router.HandleFunc("/api/ingest", api.IngestHandler).Methods("POST")
	router.HandleFunc("/api/factory/create", api.CreateFactoryHandler).Methods("POST")
	// Auth - Rotas Públicas
	router.HandleFunc("/api/register", api.RegisterHandler).Methods("POST")
	router.HandleFunc("/api/login", api.LoginHandler).Methods("POST")

	// Rotas Autenticadas via JWT
	authRouter := router.PathPrefix("/api").Subrouter()
	authRouter.Use(middleware.AuthMiddleware)
	// Onboarding (autenticado — precisa do JWT para saber qual usuário está completando o cadastro)
	authRouter.HandleFunc("/onboarding", api.OnboardingHandler).Methods("POST")
	// Fábrica (autenticado)
	authRouter.HandleFunc("/factory/generate-key", api.GenerateAPIKey).Methods("POST")
	authRouter.HandleFunc("/factory", api.CreateFactoryAuthHandler).Methods("POST")
	authRouter.HandleFunc("/factory/details", api.GetFactoryDetailsHandler).Methods("GET")
	authRouter.HandleFunc("/factory/regenerate-api-key", api.RegenerateAPIKeyHandler).Methods("POST")
	authRouter.HandleFunc("/sectors", api.GetSectorsHandler).Methods("GET")
	authRouter.HandleFunc("/sectors", api.CreateSectorHandler).Methods("POST")
	authRouter.HandleFunc("/sectors/{id}", api.UpdateSectorHandler).Methods("PUT")
	authRouter.HandleFunc("/sectors/{id}", api.DeleteSectorHandler).Methods("DELETE")
	authRouter.HandleFunc("/dashboard/data", api.GetDashboardDataHandler).Methods("GET")
	authRouter.HandleFunc("/ia/chat", api.IAChatHandler).Methods("POST")
	authRouter.HandleFunc("/ia/analysis", api.ReportIAHandler).Methods("GET")
	authRouter.HandleFunc("/machine/asset", api.UpdateMachineAssetHandler).Methods("PUT")
	authRouter.HandleFunc("/report/ia", api.ReportIAHandler).Methods("POST")
	authRouter.HandleFunc("/machine/delete", api.DeleteMachineHandler).Methods("DELETE")
	// 2FA TOTP
	authRouter.HandleFunc("/auth/2fa/setup", api.SetupTOTPHandler).Methods("GET")
	authRouter.HandleFunc("/auth/2fa/confirm", api.ConfirmTOTPHandler).Methods("POST")
	authRouter.HandleFunc("/auth/2fa/disable", api.DisableTOTPHandler).Methods("POST")
	authRouter.HandleFunc("/auth/2fa/status", api.TOTPStatusHandler).Methods("GET")

	// ─── Admin: Import Jobs ("Download Longo") ────────────────────────────────
	// All routes require JWT. Factory is inferred from the authenticated user.
	authRouter.HandleFunc("/admin/import-jobs", api.ListImportJobsHandler).Methods("GET")
	authRouter.HandleFunc("/admin/import-jobs", api.CreateImportJobHandler).Methods("POST")
	authRouter.HandleFunc("/admin/import-jobs/{id}", api.GetImportJobHandler).Methods("GET")
	authRouter.HandleFunc("/admin/import-jobs/{id}/cancel", api.CancelImportJobHandler).Methods("POST")
	authRouter.HandleFunc("/admin/import-jobs/{id}/retry", api.RetryImportJobHandler).Methods("POST")

	// Rotas com autenticação via API Key (não usam JWT middleware)
	router.HandleFunc("/api/dashboard", api.GetDashboardHandler).Methods("GET")
	router.HandleFunc("/api/dashboard/summary", api.DashboardSummaryHandler).Methods("GET")
	router.HandleFunc("/api/analytics", api.AnalyticsHandler).Methods("GET")
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

	// O Cloud Run define a variável de ambiente PORT para informar em qual porta
	// o contêiner deve escutar.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
		log.Printf("⚠️ Variável de ambiente PORT não definida, usando porta padrão: %s", port)
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

	// É crucial escutar em "0.0.0.0" (representado por ":" antes da porta)
	// para aceitar conexões de fora do contêiner.
	addr := ":" + port
	log.Printf("🚀 Servidor escutando em %s", addr)

	// Inicia o servidor HTTP. O log.Fatal irá encerrar a aplicação se o servidor não conseguir iniciar.
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("❌ Erro ao iniciar servidor: %v", err)
	}
}
