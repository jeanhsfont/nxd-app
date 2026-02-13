@echo off
REM ============================================
REM Criar Profiles Separados (SlideFlow e NXD)
REM ============================================

echo.
echo ========================================
echo   Configurar Profiles Separados
echo ========================================
echo.

echo [1/4] Criando profile para SlideFlow...
gcloud config configurations create slideflow 2>nul
gcloud config configurations activate slideflow
gcloud config set account ojeanhs@gmail.com
echo [OK] Profile SlideFlow criado

echo.
echo [2/4] Criando profile para NXD...
gcloud config configurations create nxd 2>nul
gcloud config configurations activate nxd
echo.
echo Uma janela vai abrir para login com jeanhhirata@gmail.com
pause
gcloud auth login jeanhhirata@gmail.com

if %ERRORLEVEL% NEQ 0 (
    echo [ERRO] Falha no login
    pause
    exit /b 1
)

echo.
echo [3/4] Configurando projeto NXD...
gcloud config set account jeanhhirata@gmail.com
gcloud config set project nxdata-487304

echo.
echo [4/4] Listando configuracoes...
gcloud config configurations list

echo.
echo ========================================
echo   CONFIGURACAO CONCLUIDA!
echo ========================================
echo.
echo Agora voce tem 2 profiles:
echo.
echo 1. slideflow (ojeanhs@gmail.com)
echo 2. nxd (jeanhhirata@gmail.com) - ATIVO
echo.
echo Para trocar, use: TROCAR_PROFILE.bat
echo.
pause
