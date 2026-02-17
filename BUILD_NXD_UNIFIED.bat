@echo off
echo ==========================================
echo 🏭 BUILD UNIFICADO NXD (React + Go)
echo ==========================================

echo.
echo [1/2] Construindo imagem Docker (pode demorar na primeira vez)...
docker build -f Dockerfile.nxd-unified -t nxd-core:latest .

if %ERRORLEVEL% NEQ 0 (
    echo ❌ Erro no build do Docker. Verifique os logs acima.
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo ✅ Build concluido com sucesso! Imagem: nxd-core:latest
echo.
echo Para testar localmente, execute:
echo docker run -p 8080:8080 --env-file .env nxd-core:latest
echo.
pause
