# NXD OPS KIT — Cria Cloud SQL (nxd-sql-instance) + banco nxd + secrets para deploy no slideflow-prod
# Depois de rodar: push na main dispara o workflow e o NXD sobe em um unico site (Cloud Run).
# Uso: .\create_nxd_sql_and_secrets.ps1 -ProjectId slideflow-prod

param(
    [string]$ProjectId,
    [string]$Region = "us-central1",
    [string]$InstanceName = "nxd-sql-instance",
    [string]$DatabaseName = "nxd",
    [switch]$WhatIf
)

# Garantir gcloud no PATH (Windows)
$Script:GcloudCmd = $null
$gcloudPaths = @(
    "$env:LOCALAPPDATA\Google\Cloud SDK\google-cloud-sdk\bin",
    "${env:ProgramFiles}\Google\Cloud SDK\google-cloud-sdk\bin"
)
foreach ($gp in $gcloudPaths) {
    $gc = Join-Path $gp "gcloud.cmd"
    if (Test-Path $gc) {
        $env:Path = "$gp;$env:Path"
        $Script:GcloudCmd = $gc
        break
    }
}

. "$PSScriptRoot\_lib.ps1" -ErrorAction Stop

$cfg = Get-OpsConfig
$proj = if ($ProjectId) { $ProjectId } else { $cfg["PROJECT_ID"] }
if (-not $proj) { $proj = Get-GcloudProject }
if (-not $proj) {
    Write-OpsLog "PROJECT_ID nao definido. Use -ProjectId (ex: slideflow-prod)" "ERROR"
    exit 1
}

$gcloudExe = if ($Script:GcloudCmd) { $Script:GcloudCmd } else { "gcloud" }
$null = cmd /c "`"$gcloudExe`" version 2>nul"
if ($LASTEXITCODE -ne 0) {
    Write-OpsLog "gcloud nao encontrado." "ERROR"
    exit 1
}
if (-not (Test-GcloudAuth)) {
    Write-OpsLog "gcloud nao autenticado. Execute: gcloud auth login" "ERROR"
    exit 1
}

# Senha alfanumerica para evitar problemas na URL
function New-RandomPassword {
    $chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    -join ((1..24) | ForEach-Object { $chars[(Get-Random -Maximum $chars.Length)] })
}

Write-OpsLog "=== Criar Cloud SQL + Secrets para NXD (projeto: $proj) ===" "INFO"
Write-Host ""

# 0) APIs (evitar falha por API nao habilitada)
Write-Host "0. Habilitando APIs (se necessario) ..." -ForegroundColor Cyan
$null = cmd /c "`"$gcloudExe`" services enable sqladmin.googleapis.com secretmanager.googleapis.com run.googleapis.com --project=$proj 2>nul"
Write-Host "   OK." -ForegroundColor Green

# 1) Cloud SQL instance
Write-Host "1. Cloud SQL instance: $InstanceName ..." -ForegroundColor Cyan
$exists = cmd /c "`"$gcloudExe`" sql instances describe $InstanceName --project=$proj 2>nul"
if ($LASTEXITCODE -eq 0) {
    Write-Host "   Instancia ja existe. Pulando criacao." -ForegroundColor Yellow
} elseif ($WhatIf) {
    Write-Host "   [WhatIf] Criaria: $InstanceName (POSTGRES_15, db-f1-micro, $Region)" -ForegroundColor Magenta
} else {
    Write-Host "   Criando (pode levar 5-10 min) ..." -ForegroundColor Gray
    $null = cmd /c "`"$gcloudExe`" sql instances create $InstanceName --database-version=POSTGRES_15 --tier=db-f1-micro --region=$Region --project=$proj --quiet 2>nul"
    if ($LASTEXITCODE -ne 0) {
        Write-OpsLog "Erro ao criar instancia. Verifique: gcloud sql instances create $InstanceName --project=$proj ..." "ERROR"
        exit 1
    }
    Write-Host "   OK instancia criada." -ForegroundColor Green
}

# 2) Database
Write-Host "2. Database: $DatabaseName ..." -ForegroundColor Cyan
$dbList = cmd /c "`"$gcloudExe`" sql databases list --instance=$InstanceName --project=$proj --format=value(name) 2>nul"
$dbExists = ($dbList | Select-String -Pattern "^\s*$DatabaseName\s*$" -Quiet)
if ($dbExists) {
    Write-Host "   Database ja existe." -ForegroundColor Yellow
} elseif ($WhatIf) {
    Write-Host "   [WhatIf] Criaria database $DatabaseName" -ForegroundColor Magenta
} else {
    $null = cmd /c "`"$gcloudExe`" sql databases create $DatabaseName --instance=$InstanceName --project=$proj 2>nul"
    if ($LASTEXITCODE -ne 0) {
        Write-Host "   Aviso: criar database falhou (pode ja existir). Continuando." -ForegroundColor Yellow
    } else {
        Write-Host "   OK database criado." -ForegroundColor Green
    }
}

# 3) Senha do usuario postgres e connection string
$password = New-RandomPassword
# URL para Cloud Run (socket): postgres://user:pass@/dbname?host=/cloudsql/PROJECT:REGION:INSTANCE
$connStr = "postgres://postgres:$password@/$DatabaseName?host=/cloudsql/${proj}:${Region}:${InstanceName}"

if (-not $WhatIf) {
    Write-Host "3. Definir senha do usuario postgres ..." -ForegroundColor Cyan
    $null = cmd /c "`"$gcloudExe`" sql users set-password postgres --instance=$InstanceName --project=$proj --password=$password 2>nul"
    if ($LASTEXITCODE -ne 0) {
        Write-Host "   Aviso: set-password falhou. Se a instancia e nova, tente rodar de novo em 1 min." -ForegroundColor Yellow
    } else {
        Write-Host "   OK senha definida." -ForegroundColor Green
    }
}

# 4) Secret NXD_DATABASE_URL (usar arquivo temp para nao adicionar newline)
Write-Host "4. Secret Manager: NXD_DATABASE_URL ..." -ForegroundColor Cyan
$tmpConn = [System.IO.Path]::GetTempFileName()
$connStr | Out-File -FilePath $tmpConn -NoNewline -Encoding utf8
$secretExists = cmd /c "`"$gcloudExe`" secrets describe NXD_DATABASE_URL --project=$proj 2>nul"
if ($LASTEXITCODE -ne 0) {
    if ($WhatIf) {
        Write-Host "   [WhatIf] Criaria secret NXD_DATABASE_URL" -ForegroundColor Magenta
    } else {
        & $gcloudExe secrets create NXD_DATABASE_URL --project=$proj --replication-policy=automatic --data-file=$tmpConn 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "   OK secret criado." -ForegroundColor Green
        } else {
            Write-Host "   Erro. Habilite API: gcloud services enable secretmanager.googleapis.com sqladmin.googleapis.com --project=$proj" -ForegroundColor Red
        }
    }
} else {
    if (-not $WhatIf) {
        & $gcloudExe secrets versions add NXD_DATABASE_URL --project=$proj --data-file=$tmpConn 2>$null
        if ($LASTEXITCODE -eq 0) { Write-Host "   OK nova versao." -ForegroundColor Green } else { Write-Host "   Erro ao adicionar versao." -ForegroundColor Red }
    }
}
Remove-Item $tmpConn -Force -ErrorAction SilentlyContinue

# 5) JWT_SECRET_NXD (32 bytes hex)
$jwtSecret = -join ((1..32) | ForEach-Object { "{0:x2}" -f (Get-Random -Maximum 256) })
Write-Host "5. Secret Manager: JWT_SECRET_NXD ..." -ForegroundColor Cyan
$tmpJwt = [System.IO.Path]::GetTempFileName()
$jwtSecret | Out-File -FilePath $tmpJwt -NoNewline -Encoding utf8
$jwtExists = cmd /c "`"$gcloudExe`" secrets describe JWT_SECRET_NXD --project=$proj 2>nul"
if ($LASTEXITCODE -ne 0) {
    if (-not $WhatIf) {
        & $gcloudExe secrets create JWT_SECRET_NXD --project=$proj --replication-policy=automatic --data-file=$tmpJwt 2>$null
        if ($LASTEXITCODE -eq 0) { Write-Host "   OK secret JWT criado." -ForegroundColor Green } else { Write-Host "   Aviso: criar JWT falhou." -ForegroundColor Yellow }
    }
} else {
    if (-not $WhatIf) {
        & $gcloudExe secrets versions add JWT_SECRET_NXD --project=$proj --data-file=$tmpJwt 2>$null
        if ($LASTEXITCODE -eq 0) { Write-Host "   OK nova versao JWT." -ForegroundColor Green }
    }
}
Remove-Item $tmpJwt -Force -ErrorAction SilentlyContinue

Write-Host ""
Write-OpsLog "Pronto. Proximo passo: push na branch main para disparar o deploy (Cloud Run + NXD em um unico site)." "INFO"
Write-Host "  URL do servico apos deploy: https://nxd-925156909645.us-central1.run.app" -ForegroundColor Gray
Write-Host "  Para demo: registre no site, pegue a API Key em Ajustes, configure dx-simulator/config.json e rode o simulador no seu PC." -ForegroundColor Gray
Write-Host ""
