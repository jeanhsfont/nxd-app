# NXD no GCP — site funcional pronto para vender

Um único fluxo: **limpa NXD antigo**, **reconfigura banco/secrets** e **sobe** o site (frontend novo + backend Go) no Cloud Run, com **Postgres** para registro, login e persistência.

## O que sobe

- **Um único serviço Cloud Run:** `nxd`  
  - URL: **https://nxd-925156909645.us-central1.run.app**
- **Backend Go** (API, auth, ingest, IA, etc.)
- **Frontend React** (casca nova, profissional)
- **Cloud SQL** (PostgreSQL): usuários, fábricas, setores, ativos, telemetria, config de negócio

## Passo único (recomendado)

Na **raiz do projeto**, no PowerShell:

```powershell
.\scripts\SOBE_NXD_GCP_COMPLETO.ps1
```

Isso faz, em ordem:

1. **Remove** só os serviços antigos: `hubsystem-backend`, `hubsystem-frontend`, `hubsystem-nxd` (o serviço **`nxd`** é mantido/atualizado).
2. **Garante** Cloud SQL (`nxd-sql-instance`), banco `nxd` e secrets `NXD_DATABASE_URL` e `JWT_SECRET_NXD` (cria se não existir).
3. **Build + deploy** via Cloud Build (imagem unificada → Cloud Run `nxd`).

Tempo estimado: 5–15 min (mais rápido se a instância SQL e os secrets já existirem).

## Opções do script

```powershell
# Só limpar antigos e fazer deploy (não mexe em SQL/secrets)
.\scripts\SOBE_NXD_GCP_COMPLETO.ps1 -SkipSqlAndSecrets

# Só deploy (não deleta nada, não roda create_nxd_sql)
.\scripts\SOBE_NXD_GCP_COMPLETO.ps1 -SkipDeleteOld -SkipSqlAndSecrets

# Simular (não executa nada)
.\scripts\SOBE_NXD_GCP_COMPLETO.ps1 -WhatIf
```

## Pré-requisitos

- **gcloud** instalado e logado: `gcloud auth login`
- Projeto: `gcloud config set project slideflow-prod`
- Permissões no projeto: Cloud Build, Cloud Run, Secret Manager, Cloud SQL (como já usado hoje)

## Se o deploy reclamar de permissão nos secrets

A conta de serviço do Cloud Run (`925156909645-compute@developer.gserviceaccount.com`) precisa poder **ler** os secrets. Uma vez no projeto:

```powershell
$proj = "slideflow-prod"
$sa = "925156909645-compute@developer.gserviceaccount.com"
gcloud secrets add-iam-policy-binding NXD_DATABASE_URL --project=$proj --member="serviceAccount:$sa" --role="roles/secretmanager.secretAccessor"
gcloud secrets add-iam-policy-binding JWT_SECRET_NXD --project=$proj --member="serviceAccount:$sa" --role="roles/secretmanager.secretAccessor"
```

## Depois do deploy

1. Abra **https://nxd-925156909645.us-central1.run.app**
2. **Registrar** → **Login** → **Onboarding** (dados da fábrica)
3. Use o site normalmente: setores, ativos, indicadores, IA, cobrança, suporte. Tudo persiste no Postgres.

## Deploy só da aplicação (sem limpar nem criar SQL)

Se o banco e os secrets já estiverem ok:

```powershell
.\scripts\deploy-nxd-gcp.ps1
```

ou:

```powershell
gcloud builds submit --config=cloudbuild.yaml --project=slideflow-prod .
```

---

**Resumo:** `.\scripts\SOBE_NXD_GCP_COMPLETO.ps1` deixa o NXD no GCP limpo (sem serviços antigos), com banco linkado e site funcional, pronto para criar login e ter persistência — e para vender.
