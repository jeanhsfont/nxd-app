@echo off
chcp 65001 >nul
title HUB System - Teste Simples
color 0A

cls
echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║                                                               ║
echo ║              🎯 SOLUÇÃO DEFINITIVA                            ║
echo ║                                                               ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.
echo Nova abordagem:
echo  • Build stage: Debian (SQLite funciona nativamente)
echo  • Runtime: Alpine (imagem menor)
echo  • Sem problemas de musl/glibc
echo.
pause

cls
echo.
echo 🧹 Limpando...
docker-compose down >nul 2>&1
echo ✓ Limpo
echo.

echo 🔨 Compilando (nova abordagem)...
echo.

docker-compose build hub-server

if %errorlevel% neq 0 (
    echo.
    echo ❌ Erro ao compilar
    echo.
    echo Vendo últimas linhas do erro...
    docker-compose build hub-server 2>&1 | findstr /C:"error" /C:"Error" /C:"ERROR"
    echo.
    pause
    exit /b 1
)

echo.
echo ✅ Compilado!
echo.

echo 🚀 Iniciando...
docker-compose up -d hub-server

echo.
echo ⏳ Aguardando 30 segundos...
timeout /t 30 /nobreak

echo.
echo 🌐 Abrindo navegador...
start http://localhost:8080

echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║                                                               ║
echo ║              ✅ TESTE!                                        ║
echo ║                                                               ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.
echo Se abrir o dashboard = FUNCIONOU! 🎉
echo.
echo Próximos passos:
echo  1. Crie fábrica
echo  2. Copie API Key
echo  3. Crie .env
echo  4. Execute START.bat
echo.
pause
