# =============================================================================
# NXD - Criar recursos no Google Cloud (SEPARADO do Slideflow)
# =============================================================================
# REGRAS:
# - So cria recursos com nome nxd-* (nxd-db, nxd-api)
# - NAO altera NENHUM recurso do Slideflow
# - NAO usa portas/rotas iguais ao Slideflow (NXD usa 8080 no container)
# =============================================================================
# USO: Defina NXD_GCP_PROJECT e execute no PowerShell
#      $env:NXD_GCP_PROJECT = "seu-project-id"
#      .\scripts\GCP_CREATE_NXD_RESOURCES.ps1
# =============================================================================

$ErrorActionPreference = "Stop"
$project = $env:NXD_GCP_PROJECT
if (-not $project) {
    Write-Host "ERRO: Defina o projeto primeiro:" -ForegroundColor Red
    Write-Host '  $env:NXD_GCP_PROJECT = "seu-project-id"' -ForegroundColor Yellow
    exit 1
}

Write-Host "Projeto: $project" -ForegroundColor Cyan
Write-Host "Criando APENAS recursos NXD (nxd-*). Slideflow nao sera tocado." -ForegroundColor Green
Write-Host ""

# 1) Configurar projeto
Write-Host "[1/4] Configurando projeto..." -ForegroundColor Cyan
gcloud config set project $project

# 2) Ativar APIs necessarias (nao altera APIs ja usadas pelo Slideflow)
Write-Host "[2/4] Ativando APIs (Run, SQL, Cloud Build)..." -ForegroundColor Cyan
gcloud services enable run.googleapis.com sqladmin.googleapis.com cloudbuild.googleapis.com --project=$project

# 3) Cloud SQL - Instancia NOVA so para NXD (nome: nxd-db)
Write-Host "[3/4] Criando instancia Cloud SQL nxd-db (so NXD)..." -ForegroundColor Cyan
$region = "southamerica-east1"
# Tier db-f1-micro = mais barato; pode subir para db-g1-small se precisar
gcloud sql instances create nxd-db `
  --database-version=POSTGRES_15 `
  --tier=db-f1-micro `
  --region=$region `
  --storage-size=10GB `
  --storage-type=SSD `
  --project=$project

Write-Host "Criando database nxd_app..." -ForegroundColor Cyan
gcloud sql databases create nxd_app --instance=nxd-db --project=$project
Write-Host "Usuario nxd_user: execute manualmente (guarde a senha):" -ForegroundColor Yellow
Write-Host '  gcloud sql users create nxd_user --instance=nxd-db --password=SUA_SENHA --project='$project -ForegroundColor Gray

# 4) Cloud Run - Servico NOVO so para NXD (nome: nxd-api). Porta 8080 no container.
Write-Host "[4/4] Deploy do NXD no Cloud Run (nxd-api)..." -ForegroundColor Cyan
Set-Location $PSScriptRoot\..
gcloud run deploy nxd-api `
  --source . `
  --region=$region `
  --platform=managed `
  --allow-unauthenticated `
  --set-env-vars="PORT=8080" `
  --project=$project

Write-Host ""
Write-Host "Concluido. Recursos NXD criados:" -ForegroundColor Green
Write-Host "  - Cloud SQL: nxd-db (database nxd_app, user nxd_user)" -ForegroundColor White
Write-Host "  - Cloud Run: nxd-api (URL propria, porta 8080 no container)" -ForegroundColor White
Write-Host "Slideflow nao foi alterado." -ForegroundColor Green
