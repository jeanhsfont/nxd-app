@echo off
chcp 65001 >nul
title HUB System - Validação
color 0E

cls
echo.
echo ╔═══════════════════════════════════════════════════════════════╗
echo ║                                                               ║
echo ║              🔍 VALIDAÇÃO PRÉ-BUILD                           ║
echo ║                                                               ║
echo ╚═══════════════════════════════════════════════════════════════╝
echo.

set ERROS=0

REM ============================================================
REM 1. Verifica Docker
REM ============================================================
echo [1/10] Verificando Docker...
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo     ❌ Docker não está rodando
    set /a ERROS+=1
) else (
    echo     ✓ Docker OK
)

REM ============================================================
REM 2. Verifica Go (se for compilar local)
REM ============================================================
echo [2/10] Verificando Go...
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo     ⚠️  Go não instalado (OK se usar só Docker)
) else (
    go version
    echo     ✓ Go OK
)

REM ============================================================
REM 3. Verifica arquivos essenciais
REM ============================================================
echo [3/10] Verificando arquivos essenciais...

set ARQUIVOS_OK=1

if not exist "main.go" (
    echo     ❌ main.go não encontrado
    set /a ERROS+=1
    set ARQUIVOS_OK=0
)

if not exist "go.mod" (
    echo     ❌ go.mod não encontrado
    set /a ERROS+=1
    set ARQUIVOS_OK=0
)

if not exist "docker-compose.yml" (
    echo     ❌ docker-compose.yml não encontrado
    set /a ERROS+=1
    set ARQUIVOS_OK=0
)

if not exist "Dockerfile.hub" (
    echo     ❌ Dockerfile.hub não encontrado
    set /a ERROS+=1
    set ARQUIVOS_OK=0
)

if not exist "Dockerfile.simulator" (
    echo     ❌ Dockerfile.simulator não encontrado
    set /a ERROS+=1
    set ARQUIVOS_OK=0
)

if %ARQUIVOS_OK%==1 echo     ✓ Arquivos essenciais OK

REM ============================================================
REM 4. Verifica pastas de código
REM ============================================================
echo [4/10] Verificando estrutura de pastas...

set PASTAS_OK=1

if not exist "core\" (
    echo     ❌ Pasta core/ não encontrada
    set /a ERROS+=1
    set PASTAS_OK=0
)

if not exist "api\" (
    echo     ❌ Pasta api/ não encontrada
    set /a ERROS+=1
    set PASTAS_OK=0
)

if not exist "data\" (
    echo     ❌ Pasta data/ não encontrada
    set /a ERROS+=1
    set PASTAS_OK=0
)

if not exist "services\" (
    echo     ❌ Pasta services/ não encontrada
    set /a ERROS+=1
    set PASTAS_OK=0
)

if not exist "simulator\" (
    echo     ❌ Pasta simulator/ não encontrada
    set /a ERROS+=1
    set PASTAS_OK=0
)

if not exist "web\" (
    echo     ❌ Pasta web/ não encontrada
    set /a ERROS+=1
    set PASTAS_OK=0
)

if %PASTAS_OK%==1 echo     ✓ Estrutura de pastas OK

REM ============================================================
REM 5. Verifica arquivos Go principais
REM ============================================================
echo [5/10] Verificando arquivos Go...

set GO_OK=1

if not exist "core\types.go" (
    echo     ❌ core/types.go não encontrado
    set /a ERROS+=1
    set GO_OK=0
)

if not exist "api\handlers.go" (
    echo     ❌ api/handlers.go não encontrado
    set /a ERROS+=1
    set GO_OK=0
)

if not exist "data\database.go" (
    echo     ❌ data/database.go não encontrado
    set /a ERROS+=1
    set GO_OK=0
)

if not exist "simulator\dx_simulator.go" (
    echo     ❌ simulator/dx_simulator.go não encontrado
    set /a ERROS+=1
    set GO_OK=0
)

if %GO_OK%==1 echo     ✓ Arquivos Go OK

REM ============================================================
REM 6. Verifica arquivos web
REM ============================================================
echo [6/10] Verificando arquivos web...

set WEB_OK=1

if not exist "web\index.html" (
    echo     ❌ web/index.html não encontrado
    set /a ERROS+=1
    set WEB_OK=0
)

if not exist "web\style.css" (
    echo     ❌ web/style.css não encontrado
    set /a ERROS+=1
    set WEB_OK=0
)

if not exist "web\app.js" (
    echo     ❌ web/app.js não encontrado
    set /a ERROS+=1
    set WEB_OK=0
)

if %WEB_OK%==1 echo     ✓ Arquivos web OK

REM ============================================================
REM 7. Testa sintaxe do docker-compose
REM ============================================================
echo [7/10] Validando docker-compose.yml...
docker-compose config >nul 2>&1
if %errorlevel% neq 0 (
    echo     ❌ docker-compose.yml tem erros de sintaxe
    set /a ERROS+=1
) else (
    echo     ✓ docker-compose.yml OK
)

REM ============================================================
REM 8. Verifica espaço em disco
REM ============================================================
echo [8/10] Verificando espaço em disco...
for /f "tokens=3" %%a in ('dir /-c ^| find "bytes free"') do set FREE=%%a
if defined FREE (
    echo     ✓ Espaço disponível
) else (
    echo     ⚠️  Não foi possível verificar espaço
)

REM ============================================================
REM 9. Testa compilação Go local (rápido)
REM ============================================================
echo [9/10] Testando compilação Go local...
where go >nul 2>&1
if %errorlevel% equ 0 (
    go build -o test_build.exe main.go >nul 2>&1
    if %errorlevel% neq 0 (
        echo     ❌ Código Go tem erros de compilação
        echo     Execute: go build main.go
        echo     Para ver os erros detalhados
        set /a ERROS+=1
    ) else (
        echo     ✓ Código Go compila OK
        del test_build.exe >nul 2>&1
    )
) else (
    echo     ⚠️  Go não instalado, pulando teste
)

REM ============================================================
REM 10. Verifica portas em uso
REM ============================================================
echo [10/10] Verificando porta 8080...
netstat -ano | findstr ":8080" >nul 2>&1
if %errorlevel% equ 0 (
    echo     ⚠️  Porta 8080 está em uso
    echo     Execute: docker-compose down
) else (
    echo     ✓ Porta 8080 disponível
)

echo.
echo ═══════════════════════════════════════════════════════════════
echo.

if %ERROS% equ 0 (
    echo ╔═══════════════════════════════════════════════════════════════╗
    echo ║                                                               ║
    echo ║              ✅ TUDO OK! PRONTO PARA BUILD                    ║
    echo ║                                                               ║
    echo ╚═══════════════════════════════════════════════════════════════╝
    echo.
    echo Próximo passo:
    echo  1. Execute: BUILD.bat
    echo  2. Ou execute: TESTAR_AGORA.bat
    echo.
) else (
    echo ╔═══════════════════════════════════════════════════════════════╗
    echo ║                                                               ║
    echo ║              ❌ ENCONTRADOS %ERROS% ERRO(S)                          ║
    echo ║                                                               ║
    echo ╚═══════════════════════════════════════════════════════════════╝
    echo.
    echo Corrija os erros acima antes de continuar.
    echo.
)

pause
