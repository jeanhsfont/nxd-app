# Limpeza Google Cloud — apenas NXD / Hub System (não tocar em Slideflow)

Use este guia para remover **somente** recursos antigos relacionados a **NXD** e **Hub System** no Google Cloud, sem alterar nada do **Slideflow**.

## Pré-requisitos

1. **Login no GCP**
   ```bash
   gcloud auth login
   gcloud config set project slideflow-prod
   ```

2. Confirme que está no projeto certo:
   ```bash
   gcloud config get-value project
   ```

## 1. Listar serviços Cloud Run (NXD/Hub)

No projeto `slideflow-prod` (ou no projeto onde o NXD está):

```bash
gcloud run services list --project=slideflow-prod --region=us-central1 --format="table(SERVICE,REGION)"
```

**Exclua somente** serviços cujo nome contenha `nxd`, `hubsystem` ou `hub-system`.  
**NÃO** exclua serviços com `slideflow` no nome.

Exemplo de exclusão (substitua `NOME_DO_SERVICO` pelo nome real, ex.: `hubsystem-backend`). **Não apague o serviço `nxd`** — é o único em uso:

```bash
gcloud run services delete NOME_DO_SERVICO --project=slideflow-prod --region=us-central1 --quiet
```

## 2. Banco de dados (Cloud SQL)

- **Instância usada pelo NXD no workflow:** `nxd-sql-instance` (conexão: `slideflow-prod:us-central1:nxd-sql-instance`).

**Opção A – Só limpar dados (recomendado para “subir do zero”):**

Conectar na instância e apagar apenas os esquemas/dados do NXD (por exemplo esquema `nxd` ou tabelas do app), deixando a instância no ar. Assim o próximo deploy roda as migrations do zero.

Exemplo (via Cloud SQL Proxy ou conexão autorizada):

```sql
-- Conecte no banco usado pelo NXD e execute (ajuste o nome do schema se for diferente):
DROP SCHEMA IF EXISTS nxd CASCADE;
CREATE SCHEMA nxd;
```

Ou, se o app usar apenas o schema `public` e tabelas com prefixo conhecido, dropar apenas essas tabelas.

**Opção B – Apagar a instância (drástico):**

Só faça se quiser recriar a instância do zero. O workflow atual espera `nxd-sql-instance`; se apagar, será preciso criar de novo com o mesmo nome ou alterar o workflow.

```bash
# NÃO execute se quiser manter a instância e só limpar dados
# gcloud sql instances delete nxd-sql-instance --project=slideflow-prod --quiet
```

## 3. Artifact Registry (imagens Docker)

Imagem usada no workflow: `us-central1-docker.pkg.dev/slideflow-prod/nxd-repo/nxd:latest`.

Para limpar imagens antigas do repositório NXD (sem tocar em repositórios do Slideflow):

```bash
# Listar imagens no repositório nxd-repo
gcloud artifacts docker images list us-central1-docker.pkg.dev/slideflow-prod/nxd-repo --format="table(package,version,tags)"

# Opcional: deletar imagens antigas (mantenha 'latest' se for usar no próximo deploy)
# gcloud artifacts docker images delete IMAGEM --delete-tags
```

Se quiser apagar **todo** o repositório `nxd-repo` (e recriar depois):

```bash
# Cuidado: isso remove todo o repositório nxd-repo
# gcloud artifacts repositories delete nxd-repo --location=us-central1 --project=slideflow-prod
```

## 4. Resumo do que NÃO fazer

- **Não** alterar ou excluir projetos, serviços ou recursos com **Slideflow** no nome.
- **Não** excluir recursos de outros produtos que não sejam NXD/Hub System.
- **Não** alterar o workflow em `.github/workflows/deploy-gcp.yml` para nada relacionado ao Slideflow.

## 5. Depois da limpeza

1. Confirme que não há serviços Cloud Run com nome NXD/Hub (ou que os que restaram são os desejados).
2. Se limpou só o banco (Opção A), faça um novo deploy (push na `main` ou disparo do workflow); o app sobe e as migrations criam as tabelas do zero.
3. Se removeu a instância Cloud SQL (Opção B), crie novamente a instância (e o banco) antes do próximo deploy e atualize o workflow se o nome mudar.

---

**Referência:** deploy atual em `.github/workflows/deploy-gcp.yml`  
- Projeto: `slideflow-prod`  
- Serviço Cloud Run: **`nxd`** (único). Apagar antigos: `.\scripts\delete-old-nxd-services.ps1` ou ver `docs/UM_SO_SERVICO_NXD.md`  
- Cloud SQL: `nxd-sql-instance`  
- Imagem: `us-central1-docker.pkg.dev/slideflow-prod/nxd-repo/nxd:latest`
