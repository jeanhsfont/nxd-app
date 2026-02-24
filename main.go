package main

import (
	"context"
	"hubsystem/api"
	"hubsystem/api/middleware"
	"hubsystem/internal/nxd/config"
	"hubsystem/internal/nxd/store"

	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// BuildVersion is set at compile time via -ldflags="-X main.BuildVersion=..."
var BuildVersion = "dev"

// startupWrapper abre a porta na hora; responde 200 até o handler real estar pronto (Cloud Run).
type startupWrapper struct {
	mu   sync.RWMutex
	real http.Handler
}

func (s *startupWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	h := s.real
	s.mu.RUnlock()
	if h != nil {
		h.ServeHTTP(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *startupWrapper) setHandler(h http.Handler) {
	s.mu.Lock()
	s.real = h
	s.mu.Unlock()
}

func main() {
	log.SetOutput(os.Stdout)
	log.Printf("Application entrypoint reached. Build: %s", BuildVersion)
	if BuildVersion != "dev" {
		os.Setenv("BUILD_VERSION", BuildVersion)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	wrapper := &startupWrapper{}

	// Bind síncrono na porta para Cloud Run detectar o listen dentro do timeout de startup.
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("❌ Erro ao abrir porta %s: %v", addr, err)
	}
	log.Printf("✓ Porta %s aberta (Cloud Run startup).", addr)
	go func() {
		if err := http.Serve(ln, wrapper); err != nil {
			log.Fatalf("❌ Erro ao servir: %v", err)
		}
	}()

	log.Println("🚀 Iniciando NXD (Nexus Data Exchange)...")

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
	// Consulta CNPJ (dados públicos) — pré-preenchimento no cadastro da empresa
	authRouter.HandleFunc("/cnpj", api.CNPJLookupHandler).Methods("GET")
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
	// Config negócio + indicadores financeiros (MVP)
	authRouter.HandleFunc("/business-config", api.ListBusinessConfigHandler).Methods("GET")
	authRouter.HandleFunc("/business-config", api.UpsertBusinessConfigHandler).Methods("POST")
	authRouter.HandleFunc("/tag-mappings", api.ListTagMappingsHandler).Methods("GET")
	authRouter.HandleFunc("/tag-mappings", api.UpsertTagMappingHandler).Methods("POST")
	authRouter.HandleFunc("/financial-summary", api.GetFinancialSummaryHandler).Methods("GET")
	authRouter.HandleFunc("/financial-summary/ranges", api.GetFinancialSummaryRangesHandler).Methods("GET")
	// 2FA TOTP
	authRouter.HandleFunc("/auth/2fa/setup", api.SetupTOTPHandler).Methods("GET")
	authRouter.HandleFunc("/auth/2fa/confirm", api.ConfirmTOTPHandler).Methods("POST")
	authRouter.HandleFunc("/auth/2fa/disable", api.DisableTOTPHandler).Methods("POST")
	authRouter.HandleFunc("/auth/2fa/status", api.TOTPStatusHandler).Methods("GET")
	// Cobrança e suporte (reais: persistem no banco)
	authRouter.HandleFunc("/billing/plan", api.GetBillingPlanHandler).Methods("GET")
	authRouter.HandleFunc("/billing/plan", api.UpdateBillingPlanHandler).Methods("POST")
	authRouter.HandleFunc("/support", api.CreateSupportTicketHandler).Methods("POST")

	// ─── Admin: Import Jobs ("Download Longo") ────────────────────────────────
	// All routes require JWT. Factory is inferred from the authenticated user.
	authRouter.HandleFunc("/admin/import-jobs", api.ListImportJobsHandler).Methods("GET")
	authRouter.HandleFunc("/admin/import-jobs", api.CreateImportJobHandler).Methods("POST")
	authRouter.HandleFunc("/admin/import-jobs/{id}", api.GetImportJobHandler).Methods("GET")
	authRouter.HandleFunc("/admin/import-jobs/{id}/cancel", api.CancelImportJobHandler).Methods("POST")
	authRouter.HandleFunc("/admin/import-jobs/{id}/retry", api.RetryImportJobHandler).Methods("POST")
	authRouter.HandleFunc("/admin/import-jobs/{id}/data", api.SubmitImportJobDataHandler).Methods("POST")

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
		AllowedOrigins: []string{
			"http://localhost:5173", // Vite Dev
			"http://localhost:8080", // Local Docker
			"https://hubsystem-frontend-925156909645.us-central1.run.app", // Cloud Run Frontend
			"https://hubsystem-nxd-925156909645.us-central1.run.app",     // NXD unificado (SPA servida pelo mesmo serviço)
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := middleware.RecoverMiddleware(c.Handler(router))
	wrapper.setHandler(handler)
	log.Printf("✓ Handler principal ativo em http://0.0.0.0:%s", port)

	// Inicializa auth, config e bancos (porta já aberta)
	api.InitAuth()
	config.Load()

	// Inicializa o banco da API legada (SQLite ou Postgres)
	if os.Getenv("DATABASE_URL") == "" {
		log.Println("ℹ️  DATABASE_URL not set. Initializing legacy API database with local SQLite.")
	} else {
		log.Println("ℹ️  DATABASE_URL is set. Initializing legacy API database with PostgreSQL.")
	}
	var apiDBOk bool
	for i := 0; i < 5; i++ {
		if err := api.InitDB(); err != nil {
			log.Printf("❌ Erro ao inicializar banco da API (tentativa %d/5): %v", i+1, err)
			if i < 4 {
				time.Sleep(3 * time.Second)
			}
			continue
		}
		apiDBOk = true
		break
	}
	if !apiDBOk {
		log.Printf("❌ Banco da API não inicializado após 5 tentativas — registro/login vão retornar 503")
	}
	if err := store.InitNXDDB(); err != nil {
		log.Printf("⚠️  Erro ao inicializar banco NXD (store): %v — sistema continua sem NXD store", err)
	} else {
		log.Println("✓ Banco de dados NXD (store) inicializado.")
		workerCtx, workerCancel := context.WithCancel(context.Background())
		go store.RunImportWorker(workerCtx, store.NXDDB())
		log.Println("✓ Worker de importação histórica iniciado.")
		_ = workerCancel
	}

	// Bloqueia até sinal de encerramento
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Println("\n🛑 Encerrando servidor...")
}
