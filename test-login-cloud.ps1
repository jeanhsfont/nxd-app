# Testa registro e login na URL do Cloud Run (unico servico nxd).
# Uso: .\test-login-cloud.ps1
#      .\test-login-cloud.ps1 -BaseUrl "https://nxd-925156909645.us-central1.run.app"

param(
    [string]$BaseUrl = "https://nxd-925156909645.us-central1.run.app"
)

$ErrorActionPreference = "Stop"
$email = "test-login-" + [Guid]::NewGuid().ToString("N").Substring(0,8) + "@test.local"
$password = "TestPass123!"
$name = "TestUser"

Write-Host "Base URL: $BaseUrl" -ForegroundColor Cyan
Write-Host ""

# 1) Health
Write-Host "[1/3] GET /api/health ..." -ForegroundColor Yellow
try {
    $r = Invoke-RestMethod -Uri "$BaseUrl/api/health" -Method Get -TimeoutSec 60
    Write-Host "  OK: $r" -ForegroundColor Green
} catch {
    Write-Host "  ERRO: $_" -ForegroundColor Red
    exit 1
}

# 2) Register
Write-Host "[2/3] POST /api/register ..." -ForegroundColor Yellow
$bodyRegister = @{ name = $name; email = $email; password = $password } | ConvertTo-Json
try {
    $r = Invoke-RestMethod -Uri "$BaseUrl/api/register" -Method Post -Body $bodyRegister -ContentType "application/json" -TimeoutSec 60
    Write-Host "  OK: $($r.message)" -ForegroundColor Green
} catch {
    $status = $_.Exception.Response.StatusCode.value__
    $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
    $reader.BaseStream.Position = 0
    $errBody = $reader.ReadToEnd()
    Write-Host "  ERRO ($status): $errBody" -ForegroundColor Red
    exit 1
}

# 3) Login
Write-Host "[3/3] POST /api/login ..." -ForegroundColor Yellow
$bodyLogin = @{ email = $email; password = $password } | ConvertTo-Json
try {
    $r = Invoke-RestMethod -Uri "$BaseUrl/api/login" -Method Post -Body $bodyLogin -ContentType "application/json" -TimeoutSec 60
    if ($r.token) {
        Write-Host "  OK: token recebido (length $($r.token.Length))" -ForegroundColor Green
        Write-Host ""
        Write-Host "=== LOGIN FUNCIONOU ===" -ForegroundColor Green
        exit 0
    }
    Write-Host "  ERRO: resposta sem token" -ForegroundColor Red
    exit 1
} catch {
    $status = $_.Exception.Response.StatusCode.value__
    $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
    $reader.BaseStream.Position = 0
    $errBody = $reader.ReadToEnd()
    Write-Host "  ERRO ($status): $errBody" -ForegroundColor Red
    exit 1
}
