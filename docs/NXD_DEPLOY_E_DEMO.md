# NXD: Deploy do zero + demo para investidor

Um único site (Cloud Run) com backend + frontend. Depois, rodar o simulador no seu PC para enviar dados e testar a IA e todas as funções.

---

## Parte 1 — Subir o serviço (um único site)

### 1.1 Criar banco e secrets no GCP (uma vez)

No PowerShell, na raiz do repositório:

```powershell
cd c:\HubSystem1.0
.\NXD_OPS_KIT\ps\create_nxd_sql_and_secrets.ps1 -ProjectId slideflow-prod
```

Isso vai:

- Criar a instância Cloud SQL **nxd-sql-instance** (PostgreSQL 15, tier db-f1-micro)
- Criar o database **nxd**
- Definir a senha do usuário **postgres**
- Criar/atualizar os secrets **NXD_DATABASE_URL** e **JWT_SECRET_NXD**

A primeira vez pode levar **5–10 minutos** (criação da instância). Se der erro de API não habilitada, o script tenta habilitar; se precisar, rode:

```powershell
gcloud services enable sqladmin.googleapis.com secretmanager.googleapis.com --project=slideflow-prod
```

### 1.2 Deploy (push na main)

Depois que o script terminar sem erro:

```powershell
git add .
git commit -m "Deploy NXD"
git push origin main
```

O workflow (`.github/workflows/deploy-gcp.yml`) vai:

- Fazer build da imagem unificada (React + Go)
- Fazer push para o Artifact Registry
- Fazer deploy no Cloud Run do serviço **hubsystem-backend**
- O app sobe com migrations automáticas (tabelas NXD criadas na subida)

URL do site (exemplo):

`https://hubsystem-backend-925156909645.us-central1.run.app`

(No console do GCP ou no final do workflow você vê a URL exata.)

---

## Parte 2 — Demo 100% funcional (incluindo IA e simulador no PC)

### 2.1 No site (navegador)

1. Acesse a URL do Cloud Run (ex.: acima).
2. **Registrar:** crie uma conta (e-mail e senha).
3. **Login** e conclua o **onboarding** (nome da fábrica, etc.).
4. Em **Ajustes**, copie a **API Key** (necessária para o simulador e para ingestão).

### 2.2 Simulador no seu PC (enviando dados para o NXD na nuvem)

1. Abra o `config.json` do simulador:

   `c:\HubSystem1.0\dx-simulator\config.json`

2. Ajuste:

   - **endpoint:** URL do seu serviço + `/api/ingest`  
     Ex.: `https://hubsystem-backend-925156909645.us-central1.run.app/api/ingest`
   - **api_key:** a API Key que você copiou em Ajustes

Exemplo de `config.json`:

```json
{
  "api_key": "SUA_API_KEY_COPIADA_DOS_AJUSTES",
  "endpoint": "https://hubsystem-backend-925156909645.us-central1.run.app/api/ingest",
  "update_rate_s": 10,
  "clps": [
    { "device_id": "CLP-SALA-A-01", "brand": "Siemens", "protocol": "Profinet" },
    { "device_id": "CLP-SALA-A-02", "brand": "Rockwell", "protocol": "EtherNet/IP" }
  ]
}
```

3. Rode o simulador no seu PC:

```powershell
cd c:\HubSystem1.0\dx-simulator
go run main.go
```

O simulador envia dados a cada 10 segundos (ou o valor de `update_rate_s`) para o NXD na nuvem. Você verá mensagens de sucesso/falha no terminal.

### 2.3 O que mostrar na demo (tudo no mesmo site)

- **Dashboard:** dados em tempo (quase) real dos CLPs.
- **Gestão de ativos:** ativos criados a partir do ingest.
- **NXD Intelligence:** perguntas em linguagem natural; a IA usa contexto (telemetria + agregados financeiros).
- **Indicadores financeiros:** config de negócio, mapeamento de tags, cards de resumo.
- **Importar histórico:** jobs de importação (memory ou dx_http).
- **Cobrança / Termos / Suporte:** telas mock para demonstração.
- **Ajustes:** API Key, 2FA (se habilitado).

Para a IA (Vertex) funcionar em produção, o projeto GCP precisa ter a API Vertex AI habilitada e a conta de serviço do Cloud Run com permissão para usar o Vertex (o workflow já envia `NXD_IA_PROVIDER=vertex` e variáveis do projeto/região).

---

## Resumo rápido

| Passo | Onde | Comando / Ação |
|-------|------|----------------|
| 1 | PowerShell (repo) | `.\NXD_OPS_KIT\ps\create_nxd_sql_and_secrets.ps1 -ProjectId slideflow-prod` |
| 2 | Git | `git push origin main` (dispara o deploy) |
| 3 | Navegador | Abrir URL do Cloud Run → Registrar → Login → Onboarding → Ajustes (copiar API Key) |
| 4 | dx-simulator/config.json | Colocar `endpoint` (URL + `/api/ingest`) e `api_key` |
| 5 | PC | `cd dx-simulator; go run main.go` |

Assim você sobe **um único serviço** NXD, deixa tudo ligado de novo e usa o simulador no PC para testar a IA e todas as funções do NXD em modo demonstração.
