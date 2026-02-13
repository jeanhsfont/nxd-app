@echo off
chcp 65001 > nul
cls
echo.
echo ════════════════════════════════════════════════════════════════
echo     🏭 NXD DEMO - Fábrica Vale Plast (Injetoras de Plástico)
echo ════════════════════════════════════════════════════════════════
echo.
echo   Este demo simula 1 DIA de produção em 10 MINUTOS
echo.
echo   Você verá:
echo   - 2 Injetoras (Siemens e Delta) produzindo tampas
echo   - Paradas por superaquecimento e falhas
echo   - Cálculo de prejuízo em tempo real
echo   - Comparativo de eficiência entre máquinas
echo.
echo ════════════════════════════════════════════════════════════════
echo.

set /p API_KEY="Cole sua API Key aqui: "

if "%API_KEY%"=="" (
    echo.
    echo ❌ API Key não pode ser vazia!
    pause
    exit /b 1
)

echo.
echo ✓ API Key configurada
echo ✓ Endpoint: https://nxdata-production.up.railway.app/api/ingest
echo.
echo Iniciando simulação em 3 segundos...
timeout /t 3 > nul

set NXD_ENDPOINT=https://nxdata-production.up.railway.app/api/ingest
set API_KEY=%API_KEY%

cd /d "%~dp0"
go run simulator/dx_demo.go

echo.
echo ════════════════════════════════════════════════════════════════
echo   Demo finalizado! Veja os resultados no dashboard:
echo   https://nxdata-production.up.railway.app
echo ════════════════════════════════════════════════════════════════
pause
