@echo off
REM ============================================
REM Abrir terminal com profile NXD ativo
REM ============================================

gcloud config configurations activate nxd 2>nul

echo.
echo ========================================
echo   Terminal NXD
echo ========================================
echo.
echo Profile: nxd
echo Conta: jeanhhirata@gmail.com
echo Projeto: nxdata-487304
echo.
echo ========================================
echo.

REM Mantém o terminal aberto
cmd /k
