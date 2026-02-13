@echo off
chcp 65001 >nul
title HUB System - Teste Final
color 0A

cls
echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║                                                               ║
echo ║              🧪 TESTE FINAL - TODAS CORREÇÕES                 ║
echo ║                                                               ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.
echo Correções aplicadas:
echo  ✓ gcc, musl-dev, sqlite-dev instalados
echo  ✓ CGO_ENABLED=1
echo  ✓ Tags sqlite_omit_load_extension
echo  ✓ go.sum opcional
echo.
echo Pressione qualquer tecla para começar...
pause >nul

cls
echo.
echo 🧹 Limpando tudo...
docker-compose down >nul 2>&1
docker system prune -f >nul 2>&1
echo ✓ Limpo
echo.

echo 🔨 Compilando com TODAS as correções...
echo    (Aguarde 2-4 minutos)
echo.

docker-compose build --no-cache --progress=plain

if %errorlevel% neq 0 (
    echo.
    echo ╔═══════════════════════════════════════════════════════════════╗
    echo ║                                                               ║
    echo ║              ❌ AINDA DEU ERRO!                               ║
    echo ║                                                               ║
    echo ╚═══════════════════════════════════════════════════════════════╝
    echo.
    echo Vou salvar os logs para análise...
    docker-compose build --no-cache > build_error.log 2>&1
    echo.
    echo Logs salvos em: build_error.log
    echo.
    echo Por favor, envie este arquivo para análise.
    echo.
    pause
    exit /b 1
)

cls
echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║                                                               ║
echo ║              ✅ COMPILOU COM SUCESSO!                         ║
echo ║                                                               ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

echo 🚀 Iniciando servidor...
docker-compose up -d hub-server

echo.
echo ⏳ Aguardando servidor inicializar...
echo    Tentativa 1/12 (5 segundos cada)
timeout /t 5 /nobreak >nul

REM Tenta conectar 12 vezes (60 segundos total)
set TENTATIVAS=0
:WAIT_SERVER
set /a TENTATIVAS+=1
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:8080/api/health' -TimeoutSec 2 -UseBasicParsing; if ($response.StatusCode -eq 200) { exit 0 } else { exit 1 } } catch { exit 1 }" >nul 2>&1

if %errorlevel% equ 0 goto SERVER_READY

if %TENTATIVAS% lss 12 (
    echo    Tentativa %TENTATIVAS%/12...
    timeout /t 5 /nobreak >nul
    goto WAIT_SERVER
)

echo.
echo ❌ Servidor não respondeu após 60 segundos
echo.
echo Verificando logs...
docker-compose logs hub-server
echo.
pause
exit /b 1

:SERVER_READY
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
echo 1. No navegador:
echo    • Crie uma fábrica
echo    • Copie a API Key
echo.
echo 2. Crie arquivo .env:
echo    API_KEY=sua_chave_aqui
echo.
echo 3. Execute START.bat para iniciar com simulador
echo.
echo ═══════════════════════════════════════════════════════════════
echo.
pause
