@echo off
REM ============================================
REM NXD - Conectar Simulador Local ao Cloud
REM ============================================

echo.
echo ========================================
echo   NXD - Configurar Simulador Local
echo ========================================
echo.

REM Verifica se .env existe
if not exist ".env" (
    echo [AVISO] Arquivo .env nao encontrado
    echo.
    echo Vamos criar agora...
    echo.
)

echo Digite as informacoes do servidor NXD na nuvem:
echo.

REM Pede o IP do servidor
set /p SERVER_IP="IP do servidor NXD (ex: 34.95.xxx.xxx): "
if "%SERVER_IP%"=="" (
    echo [ERRO] IP nao pode ser vazio
    pause
    exit /b 1
)

REM Pede a API Key
set /p API_KEY="API Key (ex: NXD_xxxxx...): "
if "%API_KEY%"=="" (
    echo [ERRO] API Key nao pode ser vazia
    pause
    exit /b 1
)

REM Cria/atualiza .env
echo # NXD Configuration > .env
echo API_KEY=%API_KEY% >> .env
echo NXD_ENDPOINT=http://%SERVER_IP%:8080/api/ingest >> .env
echo HUB_ENDPOINT=http://%SERVER_IP%:8080/api/ingest >> .env

echo.
echo [OK] Configuracao salva em .env
echo.
echo ========================================
echo   TESTANDO CONEXAO
echo ========================================
echo.

REM Testa conexão
powershell -Command "try { $r = Invoke-WebRequest -Uri 'http://%SERVER_IP%:8080/api/health' -UseBasicParsing -TimeoutSec 5; if ($r.StatusCode -eq 200) { Write-Host '[OK] Servidor NXD respondendo!' -ForegroundColor Green } else { Write-Host '[ERRO] Servidor nao respondeu' -ForegroundColor Red } } catch { Write-Host '[ERRO] Nao foi possivel conectar ao servidor' -ForegroundColor Red; Write-Host $_.Exception.Message -ForegroundColor Yellow }"

echo.
echo ========================================
echo   INICIAR SIMULADOR
echo ========================================
echo.
echo Deseja iniciar o simulador agora?
echo.
set /p START_SIM="(S/N): "

if /i "%START_SIM%"=="S" (
    echo.
    echo [INFO] Iniciando simulador...
    docker-compose up -d dx-simulator
    
    echo.
    echo [OK] Simulador iniciado!
    echo.
    echo Monitorar logs:
    echo   docker-compose logs -f dx-simulator
    echo.
) else (
    echo.
    echo Para iniciar o simulador depois, execute:
    echo   docker-compose up -d dx-simulator
    echo.
)

pause
