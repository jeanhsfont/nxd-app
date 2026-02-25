# NXD — Início limpo e deploy em um único serviço

## Sistema inicia limpo (sem histórico, sem logins)

- **Nenhum dado de seed.** Ao subir o servidor pela primeira vez, as migrações apenas criam as tabelas (API legada: `users`, `factories`, etc.; NXD: schema `nxd.*` ou SQLite `nxd.db`). Não há INSERT automático de usuários ou fábricas.
- **Primeiro uso:**  
  1. **Registrar** (POST /api/register) → cria o primeiro usuário.  
  2. **Login** (POST /api/login) → retorna JWT.  
  3. **Onboarding** (POST /api/onboarding, com JWT) → cria fábrica NXD, gera API Key e associa ao usuário.  
- A partir daí: Dashboard, Ativos, Setores, Importar histórico, Indicadores financeiros, IA — tudo usando esse usuário e essa fábrica.

## Validar localmente primeiro

1. Banco limpo: use um SQLite novo (`./nxd.db` + `./hubsystem.db` se não definir `DATABASE_URL`) ou um Postgres novo (defina `DATABASE_URL` e opcionalmente `NXD_DATABASE_URL`).
2. Inicie o backend: `go run .` ou `./server.exe`.
3. Inicie o frontend: `cd web-app && npm run dev`.
4. Ou use um único container: `docker-compose up --build nxd-server` (imagem unificada React + Go).
5. Fluxo: Registrar → Login → Onboarding → usar todas as telas e o DX (simulador ou real) enviando para POST /api/ingest.

## Depois de validar: um único serviço NXD

- O deploy deve ser **um único serviço** (ex.: um Cloud Run ou uma VM) que serve:
  - Backend Go (API + worker de importação).
  - Frontend React (build em `dist/` servido pelo mesmo processo).
- Use o **Dockerfile.nxd-unified** (já usado no workflow de deploy). Não é necessário manter um serviço separado de “frontend” ou “backend” do Hub System antigo.

## Posso deletar os serviços antigos do Hub System?

**Sim.** Assim que o NXD unificado estiver em produção e validado:

- **Serviços antigos** = qualquer instância antiga do “Hub System” que você não usa mais, por exemplo:
  - Um Cloud Run (ou outro) só de frontend.
  - Um Cloud Run (ou outro) só de backend com outro nome/serviço.
  - VMs ou outros hosts que rodavam versões antigas.
- Você **pode desligar/deletar** esses serviços. O que importa é manter **apenas o serviço único do NXD** (imagem unificada) e o banco (Cloud SQL ou o que estiver usando).
- **Código no repositório:** o que ainda é usado pelo NXD (auth, onboarding, API legada de users/factories para login e onboarding) deve permanecer. O que pode ser removido é código morto de features que não existem mais ou que foram totalmente substituídas pelo NXD. Se quiser, podemos revisar juntos o que pode ser apagado do repo sem quebrar o NXD.

Resumo: **nada do que foi pedido é “para o futuro”** — dx_http está implementado, Import e Indicadores financeiros estão prontos. O sistema inicia limpo; valide local e depois suba em um **serviço novo e limpo**, em um único NXD. Depois disso, os serviços antigos do Hub System podem ser deletados.
