# Deploy direto para o Google Cloud Run (sem passar pelo GitHub).
# Uso: na raiz do projeto, execute:
#   .\scripts\deploy-nxd-gcp.ps1
#
# Requer: gcloud instalado e configurado (gcloud auth login, gcloud config set project slideflow-prod)

$ErrorActionPreference = "Stop"
$project = "slideflow-prod"
$root = Split-Path $PSScriptRoot -Parent
if (-not (Test-Path (Join-Path $root "cloudbuild.yaml"))) {
    $root = Get-Location
}
Set-Location $root

Write-Host "Projeto GCP: $project" -ForegroundColor Cyan
Write-Host "Pasta: $root" -ForegroundColor Gray
Write-Host "Enviando para Cloud Build (build + deploy no Cloud Run servico 'nxd')..." -ForegroundColor Yellow
Write-Host ""

gcloud builds submit --config=cloudbuild.yaml --project=$project .

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "Deploy concluido. URL: https://nxd-925156909645.us-central1.run.app" -ForegroundColor Green
} else {
    Write-Host "Falha no deploy." -ForegroundColor Red
    exit 1
}
