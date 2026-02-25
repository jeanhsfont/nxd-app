# Conferência API NXD (backend vs frontend e simulador)

## Rotas públicas (sem JWT)

| Rota | Método | Backend (main.go) | Uso |
|------|--------|-------------------|-----|
| /api/health | GET | HealthHandler | Health check |
| /api/register | POST | RegisterHandler | Frontend: Register.jsx |
| /api/login | POST | LoginHandler | Frontend: Login.jsx |
| /api/ingest | POST | IngestHandler | **Simulador DX** + gateways (API Key no body) |

## Rotas autenticadas (JWT no header)

| Rota | Método | Backend | Uso |
|------|--------|---------|-----|
| /api/onboarding | POST | OnboardingHandler | Frontend: Onboarding.jsx |
| /api/factory/details | GET | GetFactoryDetailsHandler | AssetManagement, etc. |
| /api/factory/regenerate-api-key | POST | RegenerateAPIKeyHandler | Settings / RegenerateKeyModal |
| /api/sectors | GET, POST | GetSectorsHandler, CreateSectorHandler | Sectors, AssetManagement, Chat, FinancialIndicators |
| /api/sectors/{id} | PUT, DELETE | UpdateSectorHandler, DeleteSectorHandler | Sectors, AssetManagement |
| /api/dashboard/data | GET | GetDashboardDataHandler | Dashboard, ImportHistoric, FinancialIndicators |
| /api/ia/chat | POST | IAChatHandler | Chat.jsx |
| /api/ia/analysis | GET | ReportIAHandler | IAReports.jsx |
| /api/business-config | GET, POST | ListBusinessConfigHandler, UpsertBusinessConfigHandler | FinancialIndicators.jsx |
| /api/tag-mappings | GET, POST | ListTagMappingsHandler, UpsertTagMappingHandler | FinancialIndicators.jsx |
| /api/financial-summary | GET | GetFinancialSummaryHandler | FinancialIndicators.jsx |
| /api/financial-summary/ranges | GET | GetFinancialSummaryRangesHandler | FinancialIndicators.jsx |
| /api/machine/asset | PUT | UpdateMachineAssetHandler | AssetManagement.jsx |
| /api/admin/import-jobs | GET, POST | ListImportJobsHandler, CreateImportJobHandler | ImportHistoric.jsx |
| /api/admin/import-jobs/{id}/... | GET, POST | Get, Cancel, Retry, SubmitImportJobDataHandler | ImportHistoric.jsx |
| /api/auth/2fa/* | GET, POST | Setup, Confirm, Disable, Status | Settings.jsx |

## Banco de dados

- **API legada (auth, factories, onboarding):** `DATABASE_URL` → PostgreSQL no Cloud Run (mesma URL do NXD quando em produção).
- **NXD store (telemetria, setores, ativos, import, business config, financial agg):** `NXD_DATABASE_URL` ou `DATABASE_URL` → migrations automáticas no `store.InitNXDDB()` (main.go).
- **Cloud Run deploy:** workflow injeta `DATABASE_URL=NXD_DATABASE_URL:latest`; o store usa `DATABASE_URL` quando `NXD_DATABASE_URL` está vazio, então uma única URL atende API + NXD.

## Conclusão

- Todas as rotas usadas pelo frontend e pelo simulador existem no backend.
- Ingest (`/api/ingest`) aceita API Key no body; simulador envia `api_key`, `device_id`, `brand`, `protocol`, `timestamp`, `tags` — compatível com `IngestHandler`.
