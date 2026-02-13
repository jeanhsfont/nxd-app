@echo off
chcp 65001 >nul
title HUB System - Docker
color 0A

cls
echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║                                                               ║
echo ║              🐳 HUB SYSTEM - DOCKER                           ║
echo ║                                                               ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

REM Verifica se Docker está rodando
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker não está rodando!
    echo.
    echo Por favor, inicie o Docker Desktop e tente novamente.
    pause
    exit /b 1
)

echo ✓ Docker está rodando
echo.

REM Verifica se precisa compilar
docker images | findstr "hubsystem10" >nul 2>&1
if %errorlevel% neq 0 (
    echo 📦 Primeira execução detectada!
    echo.
    echo 🔨 Compilando imagens Docker...
    echo    (Isso pode demorar 2-5 minutos na primeira vez)
    echo.
    
    REM Gera go.sum se não existir
    if not exist "go.sum" (
        echo 📦 Preparando dependências...
        go mod tidy >nul 2>&1
    )
    
    docker-compose build
    if %errorlevel% neq 0 (
        echo.
        echo ❌ Erro ao compilar!
        echo.
        echo Tente executar: BUILD.bat
        echo Ou veja: TROUBLESHOOTING.md
        pause
        exit /b 1
    )
    echo.
    echo ✓ Imagens compiladas!
    echo.
)

REM Verifica se .env existe
if not exist ".env" (
    cls
    echo.
    echo ╔═══════════════════════════════════════════════════════════════╗
    echo ║                                                               ║
    echo ║              📋 PRIMEIRA CONFIGURAÇÃO                         ║
    echo ║                                                               ║
    echo ╚═══════════════════════════════════════════════════════════════╝
    echo.
    echo Vou iniciar apenas o servidor para você criar uma fábrica.
    echo.
    echo Pressione qualquer tecla para continuar...
    pause >nul
    
    cls
    echo.
    echo 🚀 Iniciando servidor HUB...
    echo.
    docker-compose up -d hub-server
    
    echo.
    echo ⏳ Aguardando servidor ficar pronto...
    echo    (Primeira vez pode demorar 30-60 segundos)
    echo.
    
    REM Aguarda 45 segundos
    timeout /t 45 /nobreak
    
    cls
    echo.
    echo ╔═══════════════════════════════════════════════════════════════╗
    echo ║                                                               ║
    echo ║              ✅ SERVIDOR PRONTO!                              ║
    echo ║                                                               ║
    echo ╚═══════════════════════════════════════════════════════════════╝
    echo.
    echo 🌐 Abrindo dashboard...
    start http://localhost:8080
    
    echo.
    echo ═══════════════════════════════════════════════════════════════
    echo  📋 PRÓXIMOS PASSOS:
    echo ═══════════════════════════════════════════════════════════════
    echo.
    echo 1. No navegador que abriu:
    echo    • Digite o nome da sua fábrica
    echo    • Clique em "Criar Fábrica"
    echo    • COPIE a API Key que aparecer
    echo.
    echo 2. Crie um arquivo chamado .env na raiz com:
    echo    API_KEY=cole_sua_chave_aqui
    echo.
    echo 3. Execute START.bat novamente
    echo.
    echo ═══════════════════════════════════════════════════════════════
    echo.
    echo 💡 Dica: Use o Notepad para criar o arquivo .env
    echo.
    pause
    exit /b 0
)

cls
echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║                                                               ║
echo ║              🚀 INICIANDO SISTEMA COMPLETO                    ║
echo ║                                                               ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

echo ✓ Arquivo .env encontrado
echo.
echo 🚀 Iniciando HUB System (servidor + simulador)...
echo.

docker-compose up -d

if %errorlevel% neq 0 (
    echo.
    echo ❌ Erro ao iniciar containers
    echo.
    echo Verifique os logs: docker-compose logs
    pause
    exit /b 1
)

echo.
echo ✓ Containers iniciados!
echo.
echo ⏳ Aguardando sistema ficar pronto (10 segundos)...
timeout /t 10 /nobreak >nul

cls
echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║                                                               ║
echo ║              ✅ SISTEMA RODANDO!                              ║
echo ║                                                               ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

echo 🌐 Abrindo dashboard...
start http://localhost:8080

echo.
echo ═══════════════════════════════════════════════════════════════
echo  📊 INFORMAÇÕES:
echo ═══════════════════════════════════════════════════════════════
echo.
echo  Dashboard: http://localhost:8080
echo.
echo  Comandos úteis:
echo   • docker-compose logs -f           Ver logs
echo   • docker-compose ps                Ver status
echo   • STOP.bat                         Parar tudo
echo.
echo ═══════════════════════════════════════════════════════════════
echo.
echo ⏳ Aguarde 10-15 segundos para as máquinas aparecerem
echo.
pause
