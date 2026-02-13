@echo off
REM ============================================
REM NXD - Deploy COMPLETO para Google Cloud
REM ============================================

REM Ativa automaticamente o profile NXD
gcloud config configurations activate nxd 2>nul

echo.
echo ========================================
echo   NXD - Deploy para Google Cloud
echo   Profile: nxd (jeanhhirata@gmail.com)
echo   Project: nxdata-487304
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

echo [1/5] Verificando profile NXD...
gcloud config configurations activate nxd 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [AVISO] Profile NXD nao encontrado!
    echo.
    echo Execute primeiro: SETUP_PROFILES.bat
    echo.
    pause
    exit /b 1
)
echo [OK] Profile NXD ativo

echo.
echo [2/5] Verificando autenticacao...
for /f %%i in ('gcloud config get-value account') do set ACCOUNT=%%i
if "%ACCOUNT%"=="" (
    echo [ERRO] Conta nao configurada
    pause
    exit /b 1
)
echo [OK] Conta: %ACCOUNT%
echo [OK] Projeto: nxdata-487304

echo.
echo [3/5] Ativando APIs necessarias...
powershell -ExecutionPolicy Bypass -File "NXD_OPS_KIT\ps\setup_gcloud_apis.ps1" -ProjectId "nxdata-487304"
if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] Falha ao ativar APIs
    pause
    exit /b 1
)
echo [OK] APIs ativadas

echo.
echo [4/5] Obtendo seu IP publico...
for /f %%i in ('powershell -Command "(Invoke-WebRequest -Uri 'https://api.ipify.org' -UseBasicParsing).Content"') do set MY_IP=%%i
echo [OK] Seu IP: %MY_IP%

echo.
echo ========================================
echo   CONFIGURACAO DE SEGURANCA
echo ========================================
echo.
echo Seu IP publico: %MY_IP%
echo.
echo Deseja restringir o acesso APENAS ao seu IP?
echo (Recomendado para seguranca)
echo.
echo 1. SIM - Apenas meu IP pode acessar (RECOMENDADO)
echo 2. NAO - Acesso publico (qualquer IP)
echo.
set /p SECURITY_CHOICE="Escolha (1 ou 2): "

if "%SECURITY_CHOICE%"=="1" (
    set ALLOWED_IP=%MY_IP%
    echo.
    echo [OK] Firewall configurado para IP: %MY_IP%
) else (
    set ALLOWED_IP=
    echo.
    echo [AVISO] Servidor sera acessivel publicamente!
)

echo.
echo [5/5] Iniciando deploy na VM...
echo.
echo ========================================
echo   AGUARDE: Deploy em andamento...
echo   Isso pode levar 2-3 minutos
echo ========================================
echo.

if "%ALLOWED_IP%"=="" (
    powershell -ExecutionPolicy Bypass -File "NXD_OPS_KIT\ps\deploy_nxd_vm.ps1" -ProjectId "nxdata-487304" -Zone "southamerica-east1-b"
) else (
    powershell -ExecutionPolicy Bypass -File "NXD_OPS_KIT\ps\deploy_nxd_vm.ps1" -ProjectId "nxdata-487304" -Zone "southamerica-east1-b" -AllowedIP "%ALLOWED_IP%"
)

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERRO] Deploy falhou!
    pause
    exit /b 1
)

echo.
echo ========================================
echo   DEPLOY CONCLUIDO COM SUCESSO!
echo ========================================
echo.
echo Proximos passos:
echo 1. Acesse a URL fornecida acima
echo 2. Crie uma fabrica e copie a API Key
echo 3. Configure o simulador DX no seu PC
echo.
pause
