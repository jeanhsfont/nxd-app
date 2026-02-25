# NXD OPS KIT — Lista e remove recursos GCP de NXD/Hub System (NAO mexe em Slideflow)
# Uso:
#   .\list_and_clean_nxd_hub.ps1                    # só lista
#   .\list_and_clean_nxd_hub.ps1 -DeleteRun        # remove servicos Cloud Run NXD/Hub
#   .\list_and_clean_nxd_hub.ps1 -ProjectId slideflow-prod

param(
    [string]$ProjectId,
    [string]$Region = "us-central1",
    [switch]$DeleteRun,
    [switch]$DeleteSql,
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
    Write-OpsLog "PROJECT_ID nao definido. Use -ProjectId ou config\ops.env (PROJECT_ID) ou gcloud config set project" "ERROR"
    exit 1
}

# Usar gcloud pelo caminho encontrado (evitar erro de stderr no PowerShell)
$gcloudExe = if ($Script:GcloudCmd) { $Script:GcloudCmd } else { "gcloud" }
$gcloudTest = cmd /c "`"$gcloudExe`" version 2>nul"
$gcloudOk = ($LASTEXITCODE -eq 0)
if (-not $gcloudOk) {
    Write-OpsLog "gcloud nao encontrado. Instale o Google Cloud SDK." "ERROR"
    exit 1
}

if (-not (Test-GcloudAuth)) {
    Write-OpsLog "gcloud nao autenticado. Execute: gcloud auth login" "ERROR"
    exit 1
}

# Nao tocar em nada que contenha "slideflow"
$excludePattern = "slideflow"
# Incluir apenas recursos que parecem NXD/Hub
$includePatterns = @("nxd", "hubsystem", "hub-system", "hub_system")

Write-OpsLog "=== Recursos NXD / Hub System (projeto: $proj, regiao: $Region) ===" "INFO"
Write-OpsLog "NAO sera alterado nada que contenha: $excludePattern" "INFO"
Write-Host ""

# --- Cloud Run ---
Write-Host "--- Cloud Run ---" -ForegroundColor Cyan
$runList = & $gcloudExe run services list --project=$proj --region=$Region --format="value(SERVICE)" 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "  Erro ao listar Cloud Run (ou regiao sem servicos): $runList" -ForegroundColor Yellow
} else {
    $runLines = $runList | Where-Object { $_.Trim() -ne "" }
    $toDelete = @()
    foreach ($name in $runLines) {
        $nameLower = $name.ToLowerInvariant()
        if ($nameLower -match $excludePattern) {
            Write-Host "  [IGNORADO - Slideflow] $name" -ForegroundColor DarkGray
            continue
        }
        $match = $false
        foreach ($p in $includePatterns) {
            if ($nameLower -match $p) { $match = $true; break }
        }
        if ($match) {
            Write-Host "  NXD/Hub: $name" -ForegroundColor Yellow
            $toDelete += $name
        }
    }
    if ($toDelete.Count -eq 0 -and $runLines) {
        Write-Host "  (nenhum servico Cloud Run com nome NXD/Hub nesta regiao)" -ForegroundColor Gray
    }

    if ($DeleteRun -and $toDelete.Count -gt 0) {
        foreach ($svc in $toDelete) {
            if ($WhatIf) {
                Write-Host "  [WhatIf] Deletaria: $svc" -ForegroundColor Magenta
            } else {
                Write-Host "  Deletando Cloud Run: $svc ..." -ForegroundColor Red
                $null = cmd /c "`"$gcloudExe`" run services delete $svc --project=$proj --region=$Region --quiet 2>nul"
                if ($LASTEXITCODE -eq 0) {
                    Write-Host "  OK removido: $svc" -ForegroundColor Green
                } else {
                    Write-Host "  Erro ao remover $svc" -ForegroundColor Red
                }
            }
        }
    }
}

# --- Compute Engine VMs (nome nxd/hub) ---
Write-Host ""
Write-Host "--- Compute Engine (VMs) ---" -ForegroundColor Cyan
$zones = @("us-central1-a", "us-central1-b", "southamerica-east1-b", "southamerica-east1-a")
foreach ($z in $zones) {
    $vmList = cmd /c "`"$gcloudExe`" compute instances list --project=$proj --filter=`"zone=$z`" --format=value(name) 2>nul"
    if ($LASTEXITCODE -ne 0) { continue }
    foreach ($name in ($vmList | Where-Object { $_.Trim() -ne "" })) {
        $nameLower = $name.ToLowerInvariant()
        if ($nameLower -match $excludePattern) {
            Write-Host "  [IGNORADO - Slideflow] $name ($z)" -ForegroundColor DarkGray
            continue
        }
        foreach ($p in $includePatterns) {
            if ($nameLower -match $p) {
                Write-Host "  NXD/Hub VM: $name (zona: $z)" -ForegroundColor Yellow
                break
            }
        }
    }
}

# --- Cloud SQL (instancias com nxd/hub no nome) ---
Write-Host ""
Write-Host "--- Cloud SQL ---" -ForegroundColor Cyan
$sqlToDelete = @()
$sqlList = & $gcloudExe sql instances list --project=$proj --format="value(name)" 2>&1
if ($LASTEXITCODE -eq 0) {
    foreach ($name in ($sqlList | Where-Object { $_.Trim() -ne "" })) {
        $nameLower = $name.ToLowerInvariant()
        if ($nameLower -match $excludePattern) {
            Write-Host "  [IGNORADO - Slideflow] $name" -ForegroundColor DarkGray
            continue
        }
        if ($nameLower -match "nxd|hub") {
            Write-Host "  NXD/Hub SQL: $name" -ForegroundColor Yellow
            $sqlToDelete += $name
        }
    }
    if ($sqlToDelete.Count -eq 0 -and $sqlList) {
        Write-Host "  (nenhuma instancia SQL com nome NXD/Hub)" -ForegroundColor Gray
    }
} else {
    Write-Host "  (nao foi possivel listar ou sem instancias)" -ForegroundColor Gray
}

if ($DeleteSql -and $sqlToDelete.Count -gt 0) {
    foreach ($inst in $sqlToDelete) {
        if ($WhatIf) {
            Write-Host "  [WhatIf] Deletaria instancia SQL: $inst" -ForegroundColor Magenta
        } else {
            Write-Host "  Deletando instancia Cloud SQL: $inst (pode levar varios minutos) ..." -ForegroundColor Red
            $null = cmd /c "`"$gcloudExe`" sql instances delete $inst --project=$proj --quiet 2>nul"
            if ($LASTEXITCODE -eq 0) {
                Write-Host "  OK removida: $inst" -ForegroundColor Green
            } else {
                Write-Host "  Erro ao remover $inst" -ForegroundColor Red
            }
        }
    }
} elseif (-not $DeleteSql -and $sqlToDelete.Count -gt 0) {
    Write-Host "  Para remover instancias NXD/Hub use: -DeleteSql" -ForegroundColor Gray
}

Write-Host ""
if ($DeleteRun -or $DeleteSql) {
    Write-OpsLog "Concluido. (Slideflow intacto.)" "INFO"
} else {
    Write-OpsLog "Modo listagem. Remover Cloud Run: -DeleteRun | Remover Cloud SQL NXD/Hub: -DeleteSql" "INFO"
}
Write-Host ""
