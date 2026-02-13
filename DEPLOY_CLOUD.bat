@echo off
REM ============================================
REM NXD - Deploy para Google Cloud
REM ============================================

echo.
echo ========================================
echo   NXD - Deploy para Google Cloud
echo ========================================
echo.

REM Verifica se gcloud está instalado
where gcloud >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] gcloud CLI nao encontrado!
    echo.
    echo Instale o Google Cloud SDK:
    echo https://cloud.google.com/sdk/docs/install
    echo.
    pause
    exit /b 1
)

REM Verifica autenticação
gcloud auth list --filter="status:ACTIVE" --format="value(account)" >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] gcloud nao autenticado!
    echo.
    echo Execute: gcloud auth login
    echo.
    pause
    exit /b 1
)

echo [OK] gcloud instalado e autenticado
echo.

REM Pergunta o Project ID
set /p PROJECT_ID="Digite o Project ID do Google Cloud: "
if "%PROJECT_ID%"=="" (
    echo [ERRO] Project ID nao pode ser vazio
    pause
    exit /b 1
)

echo.
echo Configurando projeto: %PROJECT_ID%
gcloud config set project %PROJECT_ID%

echo.
echo ========================================
echo   Opcoes de Deploy
echo ========================================
echo.
echo 1. Deploy em VM (Compute Engine) - Recomendado
echo 2. Deploy em Cloud Run (Serverless)
echo.
set /p DEPLOY_TYPE="Escolha (1 ou 2): "

if "%DEPLOY_TYPE%"=="1" (
    echo.
    echo [INFO] Deploy em VM selecionado
    echo.
    
    REM Pergunta o IP autorizado (opcional)
    echo Para maior seguranca, restrinja o acesso apenas ao seu IP.
    echo Seu IP publico: 
    powershell -Command "(Invoke-WebRequest -Uri 'https://api.ipify.org' -UseBasicParsing).Content"
    echo.
    set /p ALLOWED_IP="Digite seu IP (ou deixe vazio para acesso publico): "
    
    echo.
    echo [INFO] Iniciando deploy...
    echo.
    
    if "%ALLOWED_IP%"=="" (
        powershell -ExecutionPolicy Bypass -File "NXD_OPS_KIT\ps\deploy_nxd_vm.ps1" -ProjectId "%PROJECT_ID%"
    ) else (
        powershell -ExecutionPolicy Bypass -File "NXD_OPS_KIT\ps\deploy_nxd_vm.ps1" -ProjectId "%PROJECT_ID%" -AllowedIP "%ALLOWED_IP%"
    )
) else if "%DEPLOY_TYPE%"=="2" (
    echo.
    echo [INFO] Cloud Run ainda nao implementado para NXD
    echo Use a opcao 1 (VM) por enquanto
    pause
    exit /b 1
) else (
    echo [ERRO] Opcao invalida
    pause
    exit /b 1
)

echo.
echo ========================================
echo   Deploy Concluido!
echo ========================================
echo.
pause
