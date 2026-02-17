# Configuração Vertex AI no Google Cloud para o HubSystem/NXD
# Execute com: .\scripts\setup-vertex-gcloud.ps1
# Requer: gcloud instalado e já autenticado (gcloud auth login + gcloud auth application-default login)

$ErrorActionPreference = "Stop"
$ProjectId = "slideflow-prod"
$SaName = "hub-nxd-vertex"
$SaEmail = "${SaName}@${ProjectId}.iam.gserviceaccount.com"
$KeyFile = "vertex-sa-key.json"

Write-Host "Projeto: $ProjectId"
Write-Host "Conta de servico: $SaEmail"
Write-Host ""

# 1) Habilitar API Vertex AI
Write-Host "[1/4] Habilitando API Vertex AI..."
gcloud services enable aiplatform.googleapis.com --project=$ProjectId
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

# 2) Criar conta de servico (ignora erro se ja existir)
Write-Host "[2/4] Criando conta de servico $SaName..."
$createOut = gcloud iam service-accounts create $SaName --display-name="HubSystem NXD Vertex AI" --project=$ProjectId 2>&1
if ($LASTEXITCODE -ne 0) {
    if ($createOut -match "already exists") {
        Write-Host "      (conta ja existe, continuando)"
    } else {
        Write-Host $createOut
        exit 1
    }
}

# 3) Conceder papel Vertex AI User
Write-Host "[3/4] Concedendo roles/aiplatform.user..."
gcloud projects add-iam-policy-binding $ProjectId `
  --member="serviceAccount:$SaEmail" `
  --role="roles/aiplatform.user" `
  --quiet
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

# 4) Criar chave JSON (salva na raiz do projeto; esta no .gitignore)
Write-Host "[4/4] Criando chave em $KeyFile..."
$KeyPath = Join-Path (Get-Location) $KeyFile
gcloud iam service-accounts keys create $KeyPath `
  --iam-account=$SaEmail `
  --project=$ProjectId
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "Concluido. Configure no ambiente (PowerShell):"
Write-Host "  `$env:NXD_IA_PROVIDER = 'vertex'"
Write-Host "  `$env:VERTEX_AI_PROJECT = '$ProjectId'"
Write-Host "  `$env:VERTEX_AI_LOCATION = 'us-central1'"
Write-Host "  `$env:GOOGLE_APPLICATION_CREDENTIALS = (Resolve-Path '$KeyFile').Path"
Write-Host ""
Write-Host "Ou adicione ao seu .env (nao commitar):"
Write-Host "  NXD_IA_PROVIDER=vertex"
Write-Host "  VERTEX_AI_PROJECT=$ProjectId"
Write-Host "  VERTEX_AI_LOCATION=us-central1"
Write-Host "  GOOGLE_APPLICATION_CREDENTIALS=$KeyPath"
