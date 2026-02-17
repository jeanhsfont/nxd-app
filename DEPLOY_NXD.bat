@echo off
echo ==========================================
echo 🚀 DEPLOY NXD CORE (Google Cloud)
echo ==========================================

REM Configura variaveis
set PROJECT_ID=slideflow-prod
set REGION=us-central1
set REPO=nxd-repo
set IMAGE=nxd-core
set TAG=latest
set FULL_IMAGE=%REGION%-docker.pkg.dev/%PROJECT_ID%/%REPO%/%IMAGE%:%TAG%

echo.
echo [1/3] Construindo imagem Docker Unificada...
docker build -f Dockerfile.nxd-unified -t %FULL_IMAGE% .
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Erro no build. Verifique o Dockerfile.
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo [2/3] Autenticando Docker no GCloud...
gcloud auth configure-docker %REGION%-docker.pkg.dev --quiet

echo.
echo [3/3] Enviando imagem para o Artifact Registry...
docker push %FULL_IMAGE%
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Erro no push. Verifique permissoes ou rede.
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo ✅ Imagem enviada com sucesso: %FULL_IMAGE%
echo.
echo Agora aguarde a criacao do Cloud SQL para fazer o deploy do servico Cloud Run.
pause
