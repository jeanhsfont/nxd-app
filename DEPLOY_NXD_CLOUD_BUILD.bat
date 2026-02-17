@echo off
echo ==========================================
echo ☁️ DEPLOY NXD CORE (Via Cloud Build)
echo ==========================================

REM Configura variaveis
set PROJECT_ID=slideflow-prod
set REGION=us-central1
set REPO=nxd-repo
set IMAGE=nxd-core
set TAG=latest
set FULL_IMAGE=%REGION%-docker.pkg.dev/%PROJECT_ID%/%REPO%/%IMAGE%:%TAG%

echo.
echo [1/2] Enviando codigo para Cloud Build...
echo Isso vai construir a imagem Docker diretamente nos servidores do Google.
gcloud builds submit --config cloudbuild.nxd.yaml .

if %ERRORLEVEL% NEQ 0 (
    echo ❌ Erro no Cloud Build. Verifique os logs.
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo ✅ Build e Push concluidos com sucesso: %FULL_IMAGE%
echo.
echo Agora aguarde a criacao do Cloud SQL para fazer o deploy do servico Cloud Run.
pause
