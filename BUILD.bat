@echo off
chcp 65001 >nul
title HUB System - Build
color 0E

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║                                                               ║
echo ║              🔨 HUB SYSTEM - BUILD                            ║
echo ║                                                               ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

REM Verifica Docker
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker não está rodando!
    pause
    exit /b 1
)

echo ✓ Docker está rodando
echo.

REM Gera go.sum se não existir
if not exist "go.sum" (
    echo 📦 Gerando go.sum...
    go mod tidy
    echo ✓ go.sum gerado
    echo.
)

REM Para containers antigos
echo 🛑 Parando containers antigos...
docker-compose down >nul 2>&1
echo.

REM Remove imagens antigas
echo 🗑️  Removendo imagens antigas...
docker rmi hubsystem10-hub-server hubsystem10-dx-simulator >nul 2>&1
echo.

REM Compila imagens
echo 🔨 Compilando imagens Docker...
echo    (Isso pode demorar 2-5 minutos)
echo.

docker-compose build --no-cache

if %errorlevel% neq 0 (
    echo.
    echo ❌ Erro ao compilar!
    echo.
    echo Verifique:
    echo  - Todos os arquivos estão presentes
    echo  - Docker tem espaço em disco
    echo  - Conexão com internet está funcionando
    echo.
    pause
    exit /b 1
)

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║                                                               ║
echo ║              ✅ BUILD CONCLUÍDO!                              ║
echo ║                                                               ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.
echo Agora execute: START.bat
echo.
pause
