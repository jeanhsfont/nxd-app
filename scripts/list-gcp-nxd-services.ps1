# Lista recursos GCP relacionados a NXD / Hub System (nao mexe em Slideflow).
# Uso: .\scripts\list-gcp-nxd-services.ps1
# Requer: gcloud config set project slideflow-prod (ou seu projeto)

$ErrorActionPreference = "Stop"
$project = "slideflow-prod"
$region = "us-central1"

Write-Host "Projeto: $project | Regiao: $region" -ForegroundColor Cyan
Write-Host ""

Write-Host "=== Cloud Run (servicos que contem nxd, hubsystem, hub-system) ===" -ForegroundColor Yellow
$services = gcloud run services list --project=$project --region=$region --format="value(SERVICE)" 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Erro ao listar Cloud Run. Faca: gcloud auth login ; gcloud config set project $project" -ForegroundColor Red
    exit 1
}
$nxdHub = $services | Where-Object { $_ -match "nxd|hubsystem|hub-system" }
foreach ($s in $nxdHub) { Write-Host "  - $s" }
if (-not $nxdHub) { Write-Host "  (nenhum encontrado)" }

Write-Host ""
Write-Host "Para apagar os servicos ANTIGOS e deixar so o 'nxd':" -ForegroundColor Yellow
Write-Host "  .\scripts\delete-old-nxd-services.ps1"
Write-Host "Ou manualmente: gcloud run services delete NOME --project=$project --region=$region --quiet"
Write-Host ""
Write-Host "Um unico servico: docs\UM_SO_SERVICO_NXD.md | Limpeza/banco: docs\LIMPEZA_GCP_NXD_HUB.md" -ForegroundColor Cyan
