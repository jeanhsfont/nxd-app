@echo off
REM ============================================
REM Trocar entre Profiles (SlideFlow e NXD)
REM ============================================

echo.
echo ========================================
echo   Trocar Profile do gcloud
echo ========================================
echo.

echo Profiles disponiveis:
gcloud config configurations list

echo.
echo ========================================
echo   Escolha o profile:
echo ========================================
echo.
echo 1. slideflow (ojeanhs@gmail.com - SlideFlow)
echo 2. nxd (jeanhhirata@gmail.com - NXD)
echo.
set /p CHOICE="Escolha (1 ou 2): "

if "%CHOICE%"=="1" (
    echo.
    echo [INFO] Ativando profile: slideflow
    gcloud config configurations activate slideflow
    echo [OK] Profile ativo: slideflow
    echo [OK] Conta: ojeanhs@gmail.com
) else if "%CHOICE%"=="2" (
    echo.
    echo [INFO] Ativando profile: nxd
    gcloud config configurations activate nxd
    echo [OK] Profile ativo: nxd
    echo [OK] Conta: jeanhhirata@gmail.com
    echo [OK] Projeto: nxdata-487304
) else (
    echo [ERRO] Opcao invalida
    pause
    exit /b 1
)

echo.
echo ========================================
echo   Configuracao Atual:
echo ========================================
gcloud config list

echo.
pause
