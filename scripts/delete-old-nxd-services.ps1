# Apaga os servicos antigos do Cloud Run (hubsystem-backend, hubsystem-frontend, hubsystem-nxd).
# Mantem apenas o servico "nxd" (unico servidor NXD).
# Uso: .\scripts\delete-old-nxd-services.ps1
# Requer: gcloud config set project slideflow-prod

$ErrorActionPreference = "Stop"
$project = "slideflow-prod"
$region = "us-central1"

$oldServices = @("hubsystem-backend", "hubsystem-frontend", "hubsystem-nxd")

Write-Host "Projeto: $project | Regiao: $region" -ForegroundColor Cyan
Write-Host "Servicos antigos a remover: $($oldServices -join ', ')" -ForegroundColor Yellow
Write-Host "O servico 'nxd' NAO sera apagado (e o unico que o repo usa)." -ForegroundColor Green
Write-Host ""

foreach ($name in $oldServices) {
    Write-Host "Deletando $name ..." -ForegroundColor Gray
    gcloud run services delete $name --project=$project --region=$region --quiet 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "  (servico nao existe ou ja foi removido)" -ForegroundColor DarkGray
    } else {
        Write-Host "  OK" -ForegroundColor Green
    }
}

Write-Host ""
Write-Host "Listando servicos restantes:" -ForegroundColor Cyan
gcloud run services list --project=$project --region=$region --format="table(SERVICE,REGION)"
Write-Host ""
Write-Host "Deve restar apenas: nxd" -ForegroundColor Green
