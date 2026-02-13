# NXD_OPS_KIT — biblioteca comum (PowerShell puro, sem módulos externos)
# Uso: . "$PSScriptRoot/_lib.ps1"

$ErrorActionPreference = "Stop"
$Script:OPS_KIT_ROOT = $null
$Script:OPS_CONFIG = @{}

function Get-OpsKitRoot {
    if ($Script:OPS_KIT_ROOT) { return $Script:OPS_KIT_ROOT }
    $dir = $PSScriptRoot
    while ($dir) {
        if (Test-Path (Join-Path $dir "config")) {
            $Script:OPS_KIT_ROOT = $dir
            return $dir
        }
        $dir = Split-Path $dir -Parent
    }
    $Script:OPS_KIT_ROOT = Split-Path $PSScriptRoot -Parent
    return $Script:OPS_KIT_ROOT
}

function Get-OpsConfig {
    $root = Get-OpsKitRoot
    $envPath = Join-Path $root "config\ops.env"
    if (-not (Test-Path $envPath)) {
        $example = Join-Path $root "config\ops.env.example"
        if (Test-Path $example) {
            Write-Warning "Arquivo config\ops.env nao encontrado. Copie config\ops.env.example para config\ops.env e ajuste."
        }
        return @{
            PROJECT_ID = $env:GCLOUD_PROJECT
            REGION = "southamerica-east1"
            SERVICE_NXD = "nxd-server"
            DOMAIN_NXD = "nxd.yourdomain.com"
            VM_MACHINE_TYPE = "e2-micro"
            VM_DISK_SIZE = "10GB"
            VM_ZONE = "southamerica-east1-b"
            NXD_PORT = "8080"
            REQUIRE_GCLOUD_AUTH = "true"
        }
    }
    if ($Script:OPS_CONFIG.Count -gt 0) { return $Script:OPS_CONFIG }
    Get-Content $envPath | ForEach-Object {
        $line = $_.Trim()
        if ($line -and -not $line.StartsWith("#")) {
            $idx = $line.IndexOf("=")
            if ($idx -gt 0) {
                $key = $line.Substring(0, $idx).Trim()
                $val = $line.Substring($idx + 1).Trim().Trim('"').Trim("'")
                $Script:OPS_CONFIG[$key] = $val
            }
        }
    }
    return $Script:OPS_CONFIG
}

function Test-GcloudInstalled {
    try {
        $v = gcloud version 2>&1
        return $LASTEXITCODE -eq 0
    } catch {
        return $false
    }
}

function Test-GcloudAuth {
    try {
        $out = gcloud auth list --filter="status:ACTIVE" --format="value(account)" 2>&1
        return ($out -and $out.Trim().Length -gt 0)
    } catch {
        return $false
    }
}

function Get-GcloudProject {
    try {
        $p = gcloud config get-value project 2>&1
        if ($p -and $p -ne "(unset)") { return $p.Trim() }
    } catch {}
    return $null
}

function Write-OpsLog {
    param([string]$Message, [string]$Level = "INFO")
    $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $line = "[$ts] [$Level] $Message"
    Write-Host $line
    return $line
}

function Get-ReportPath {
    $root = Get-OpsKitRoot
    $reportsDir = Join-Path $root "reports"
    if (-not (Test-Path $reportsDir)) { New-Item -ItemType Directory -Path $reportsDir -Force | Out-Null }
    $name = "report_" + (Get-Date -Format "yyyyMMdd_HHmm") + ".md"
    return Join-Path $reportsDir $name
}

function Add-ReportContent {
    param([string]$Path, [string]$Content)
    if (-not $Path) { return }
    Add-Content -Path $Path -Value $Content -Encoding UTF8
}

function Invoke-WebRequestSafe {
    param([string]$Uri, [string]$Method = "GET", [hashtable]$Headers = @{}, [int]$TimeoutSec = 15)
    try {
        $params = @{
            Uri = $Uri
            Method = $Method
            UseBasicParsing = $true
            TimeoutSec = $TimeoutSec
            ErrorAction = "Stop"
        }
        if ($Headers.Count -gt 0) { $params["Headers"] = $Headers }
        $r = Invoke-WebRequest @params
        return @{ StatusCode = $r.StatusCode; Headers = $r.Headers; Content = $r.Content }
    } catch {
        $status = $null
        if ($_.Exception.Response) { try { $status = [int]$_.Exception.Response.StatusCode } catch {} }
        return @{ StatusCode = $status; Error = $_.Exception.Message }
    }
}

function Resolve-Dns {
    param([string]$HostName)
    try {
        $r = Resolve-DnsName -Name $HostName -Type A -ErrorAction SilentlyContinue
        if ($r) { return ($r | Where-Object { $_.Type -eq "A" } | ForEach-Object { $_.IPAddress }) }
    } catch {}
    return @()
}

# Dot-source: . "$PSScriptRoot/_lib.ps1" — funcoes ficam no escopo do chamador
