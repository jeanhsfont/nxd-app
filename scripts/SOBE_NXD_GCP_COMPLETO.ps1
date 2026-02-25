# =============================================================================
# NXD — Limpa antigos, reconfigura banco/secrets e sobe no GCP (site funcional)
# =============================================================================
# Faz:
#   1. Remove apenas servicos Cloud Run ANTIGOS (hubsystem-backend, hubsystem-frontend, hubsystem-nxd)
#   2. Garante Cloud SQL (nxd-sql-instance) + banco nxd + secrets (NXD_DATABASE_URL, JWT_SECRET_NXD)
#   3. Build + deploy do servico "nxd" (frontend novo + backend Go)
# Resultado: site pronto para registro, login, persistencia em Postgres.
# =============================================================================
# Uso (na raiz do repo):
#   .\scripts\SOBE_NXD_GCP_COMPLETO.ps1
#   .\scripts\SOBE_NXD_GCP_COMPLETO.ps1 -ProjectId slideflow-prod
# =============================================================================

param(
    [string]$ProjectId = "slideflow-prod",
    [string]$Region = "us-central1",
    [switch]$SkipDeleteOld,
    [switch]$SkipSqlAndSecrets,
    [switch]$WhatIf
)

$ErrorActionPreference = "Stop"
$root = Split-Path $PSScriptRoot -Parent
Set-Location $root

Write-Host ""
Write-Host "=== NXD: Limpeza + Banco + Deploy no GCP ===" -ForegroundColor Cyan
Write-Host "Projeto: $ProjectId | Regiao: $Region" -ForegroundColor Gray
Write-Host ""

# --- 1. Remover servicos antigos (manter apenas "nxd") ---
if (-not $SkipDeleteOld) {
    Write-Host "[1/3] Removendo servicos Cloud Run antigos (hubsystem-*) ..." -ForegroundColor Yellow
    $oldServices = @("hubsystem-backend", "hubsystem-frontend", "hubsystem-nxd")
    foreach ($name in $oldServices) {
        if ($WhatIf) {
            Write-Host "  [WhatIf] Deletaria: $name" -ForegroundColor DarkGray
        } else {
            gcloud run services delete $name --project=$ProjectId --region=$Region --quiet 2>&1
            if ($LASTEXITCODE -ne 0) {
                Write-Host "  $name (nao existe, ok)" -ForegroundColor DarkGray
            } else {
                Write-Host "  Removido: $name" -ForegroundColor Green
            }
        }
    }
    Write-Host "  Servico 'nxd' mantido (unico em uso)." -ForegroundColor Green
    Write-Host ""
} else {
    Write-Host "[1/3] Pulando remocao de servicos antigos (-SkipDeleteOld)." -ForegroundColor Gray
    Write-Host ""
}

# --- 2. Cloud SQL + Secrets (banco linkado ao Run) ---
if (-not $SkipSqlAndSecrets) {
    Write-Host "[2/3] Garantindo Cloud SQL (nxd-sql-instance) e secrets ..." -ForegroundColor Yellow
    $opsScript = Join-Path $root "NXD_OPS_KIT\ps\create_nxd_sql_and_secrets.ps1"
    if (Test-Path $opsScript) {
        if ($WhatIf) {
            Write-Host "  [WhatIf] Executaria: create_nxd_sql_and_secrets.ps1 -ProjectId $ProjectId" -ForegroundColor DarkGray
        } else {
            & $opsScript -ProjectId $ProjectId -Region $Region
            if ($LASTEXITCODE -ne 0) {
                Write-Host "  Aviso: script de SQL/secrets falhou. Se a instancia e os secrets ja existem, prossiga com o deploy." -ForegroundColor Yellow
            }
        }
    } else {
        Write-Host "  NXD_OPS_KIT\ps\create_nxd_sql_and_secrets.ps1 nao encontrado. Execute manualmente se precisar criar banco/secrets." -ForegroundColor Yellow
    }
    Write-Host ""
} else {
    Write-Host "[2/3] Pulando SQL/secrets (-SkipSqlAndSecrets)." -ForegroundColor Gray
    Write-Host ""
}

# --- 3. Build e deploy (Cloud Build -> Cloud Run) ---
Write-Host "[3/3] Build + Deploy (Cloud Build -> servico nxd) ..." -ForegroundColor Yellow
if ($WhatIf) {
    Write-Host "  [WhatIf] Executaria: gcloud builds submit --config=cloudbuild.yaml --project=$ProjectId ." -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "Para rodar de verdade, execute sem -WhatIf." -ForegroundColor Cyan
    exit 0
}

gcloud builds submit --config=cloudbuild.yaml --project=$ProjectId .

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "=== Pronto ===" -ForegroundColor Green
    Write-Host "  URL: https://nxd-925156909645.us-central1.run.app" -ForegroundColor White
    Write-Host "  Registrar -> Login -> Onboarding -> usar o site (persistencia em Postgres)." -ForegroundColor White
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "Deploy falhou. Verifique os logs acima." -ForegroundColor Red
    exit 1
}
