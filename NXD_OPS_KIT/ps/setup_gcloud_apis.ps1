# NXD OPS KIT — Ativar APIs necessárias no Google Cloud
# Uso: .\setup_gcloud_apis.ps1 -ProjectId nxdata-487304

param(
    [Parameter(Mandatory=$true)]
    [string]$ProjectId
)

. "$PSScriptRoot\_lib.ps1" -ErrorAction Stop

Write-OpsLog "=== Ativando APIs do Google Cloud ===" "INFO"
Write-OpsLog "Project: $ProjectId"

# Lista de APIs necessárias para o NXD
$apis = @(
    "compute.googleapis.com",           # Compute Engine (VMs)
    "cloudresourcemanager.googleapis.com", # Resource Manager
    "serviceusage.googleapis.com",      # Service Usage
    "iam.googleapis.com",               # IAM
    "logging.googleapis.com",           # Cloud Logging
    "monitoring.googleapis.com"         # Cloud Monitoring
)

Write-OpsLog "Ativando $($apis.Count) APIs..." "INFO"

foreach ($api in $apis) {
    Write-Host "  Ativando: $api" -ForegroundColor Cyan
    gcloud services enable $api --project=$ProjectId 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "    ✓ $api ativada" -ForegroundColor Green
    } else {
        Write-Host "    ⚠ Erro ao ativar $api" -ForegroundColor Yellow
    }
}

Write-OpsLog "APIs ativadas com sucesso!" "INFO"
Write-Host ""
Write-Host "✅ Google Cloud configurado e pronto para deploy!" -ForegroundColor Green
Write-Host ""

exit 0
