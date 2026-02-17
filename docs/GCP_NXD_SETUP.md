# NXD no Google Cloud — Setup separado (sem tocar no Slideflow)

## Regras

- **Só criamos recursos com prefixo `nxd-`.** Nenhum recurso do Slideflow é alterado.
- **Portas:** NXD usa porta **8080** dentro do container (padrão Cloud Run). A URL externa é única (ex: `nxd-xxx.run.app`). Não configuramos nenhuma porta ou rota usada pelo Slideflow.
- **Banco:** instância ou database **só para NXD** (nome `nxd-*`).

---

## Pré-requisito

Definir o **ID do projeto** (o mesmo que você usa para Slideflow, mesma cobrança):

```powershell
$env:NXD_GCP_PROJECT = "SEU_PROJECT_ID_AQUI"
```

(Substitua `SEU_PROJECT_ID_AQUI` pelo ID do projeto no Console do Google Cloud.)

---

## 1. Ativar APIs (só se ainda não estiver)

```powershell
gcloud config set project $env:NXD_GCP_PROJECT
gcloud services enable run.googleapis.com sqladmin.googleapis.com cloudbuild.googleapis.com --project=$env:NXD_GCP_PROJECT
```

---

## 2. Cloud SQL — Banco só do NXD

Cria uma instância **nova** (só para NXD). Não usa a instância do Slideflow.

### Via Console (feito em 13/02/2026)

- **Instância:** `nxd-db` — PostgreSQL 18, Edição Enterprise Sandbox (2 vCPU, 8 GB RAM, 10 GB SSD), região **us-central1**.
- **Conexão:** `slideflow-prod:us-central1:nxd-db` — porta **5432**, IP público ativado.
- **Senha do usuário `postgres`:** gerada no Console no momento da criação (botão "Gerar senha"). **Guarde em local seguro** (ex.: cofre de senhas); será usada para conectar ao banco e para o Cloud Run quando migrarmos para PostgreSQL.

Quando a instância estiver **pronta** (status "Em execução"), opcionalmente crie um database e usuário dedicados:

```powershell
gcloud sql databases create nxd_app --instance=nxd-db --project=$env:NXD_GCP_PROJECT
gcloud sql users create nxd_user --instance=nxd-db --password=SUA_SENHA_SEGURA_AQUI --project=$env:NXD_GCP_PROJECT
```

*(Ou use o usuário `postgres` com a senha gerada; aí não precisa criar `nxd_user` até querer segregar.)*

### Via gcloud (alternativa)

```powershell
# Ajuste a região se quiser (ex: southamerica-east1 para SP)
gcloud sql instances create nxd-db `
  --database-version=POSTGRES_15 `
  --tier=db-f1-micro `
  --region=southamerica-east1 `
  --storage-size=10GB `
  --storage-type=SSD `
  --project=$env:NXD_GCP_PROJECT
```

Depois, criar o database e usuário (comandos acima).

---

## 3. Cloud Run — Serviço só do NXD

Nome do serviço: **nxd-api**. URL será algo como `https://nxd-api-xxxxx.run.app` (não conflita com Slideflow). **Não altera nenhum serviço do Slideflow.**

```powershell
# Na pasta do NXD (onde está main.go, Dockerfile, etc.)
cd c:\HubSystem1.0

gcloud run deploy nxd-api `
  --source . `
  --region=us-central1 `
  --platform=managed `
  --allow-unauthenticated `
  --set-env-vars="PORT=8080" `
  --project=slideflow-prod
```

- A primeira build pode levar **5–15 minutos** (compile Go + push da imagem).
- Quando o NXD usar PostgreSQL, adicione as env vars de conexão (ex: `DATABASE_URL`) no deploy.
- O `.gcloudignore` na raiz evita enviar simuladores e docs; só sobe o necessário para o build.

---

## 4. O que NÃO fazemos

- Não alteramos projeto, VPC ou billing do Slideflow.
- Não usamos nomes de serviço, database ou instância do Slideflow.
- Não configuramos portas iguais às do Slideflow (NXD usa 8080 no container; a URL é a do Cloud Run).

---

## 5. Conferir depois

- **Projeto:** mesmo do Slideflow (`gcloud config get-value project`).
- **Recursos NXD:** só os que têm nome `nxd-*` (ex: `nxd-db`, `nxd-api`).
- **Slideflow:** permanece intocado.
