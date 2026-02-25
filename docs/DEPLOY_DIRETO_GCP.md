# Deploy direto no Google Cloud (sem GitHub)

O deploy do NXD e feito **direto no Google Cloud**, nao por push no Git.

## Como fazer deploy

Na **raiz do projeto** (onde esta `cloudbuild.yaml`), no PowerShell:

```powershell
.\scripts\deploy-nxd-gcp.ps1
```

Ou manualmente:

```powershell
gcloud config set project slideflow-prod
gcloud builds submit --config=cloudbuild.yaml --project=slideflow-prod .
```

O Cloud Build vai:
1. Construir a imagem (Dockerfile.nxd-unified)
2. Enviar para Artifact Registry (nxd-repo/nxd)
3. Fazer deploy no Cloud Run no servico **nxd**

**URL apos o deploy:** https://nxd-925156909645.us-central1.run.app

## Requisitos

- `gcloud` instalado e logado: `gcloud auth login`
- Projeto configurado: `gcloud config set project slideflow-prod`
- Permissoes no projeto para Cloud Build e Cloud Run

## GitHub

O workflow em `.github/workflows/deploy-gcp.yml` **nao** roda em push (esta desativado para push). Deploy e apenas pelo comando acima.
