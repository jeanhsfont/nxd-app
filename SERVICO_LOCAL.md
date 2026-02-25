# Serviço local — o que esperar e o que está operante

## 1. O que você deve esperar de funcionalidade real (local)

| Funcionalidade | Local | Observação |
|----------------|--------|------------|
| **Registro / Login / Onboarding** | ✅ | Cria usuário e fábrica, gera API Key. |
| **Dashboard** | ✅ | Mostra ativos e métricas a partir do que o DX (ou simulador) enviar. |
| **Gestão de ativos e setores** | ✅ | Criar setores, arrastar máquinas. |
| **Importar histórico** | ✅ | Jobs “memory” (colar JSON) e “dx_http” (GET na URL do DX). |
| **Indicadores financeiros** | ✅ | Config. negócio + mapeamento de tags; cards Hoje/24h/7d. |
| **NXD Intelligence (IA)** | ⚠️ Depende de credenciais | O servidor **local** chama a **nuvem** (Vertex AI / Gemini). Sem credenciais GCP, a IA falha. |
| **Simulador DX** | ✅ | Existe no repositório; você configura e roda local. |

Resumo: tudo que não depende de API externa funciona 100% local. A IA funciona local **desde que** o ambiente tenha credenciais GCP válidas (Vertex AI).

---

## 2. IA “local” — está operante?

- A IA **não roda um modelo na sua máquina**. O backend local envia o prompt para o **Gemini no Vertex AI (Google Cloud)**.
- Para a IA responder no ambiente local você precisa:
  - **Opção A:** Conta GCP com Vertex AI habilitado e uma das opções:
    - `gcloud auth application-default login` (uso com sua conta), ou
    - `GOOGLE_APPLICATION_CREDENTIALS` apontando para chave JSON de uma service account com permissão no Vertex AI.
  - **Opção B:** Ajustar o código para usar outra API (ex.: Gemini via API key), se quiser evitar GCP.
- Se **não** configurar credenciais: o resto do sistema funciona; só a tela “NXD Intelligence” (chat/relatório) pode retornar erro ao chamar o modelo.

Conclusão: **IA “local” = servidor local + Vertex na nuvem**. Está operante se as credenciais GCP estiverem corretas no ambiente onde o backend roda.

---

## 3. Tem servidor local para simular o DX?

**Sim.** O simulador está na pasta **`dx-simulator`**.

### Como usar o simulador local

1. **Subir o NXD** (backend + frontend), fazer **Registro → Login → Onboarding** e **copiar a API Key**.
2. **Configurar o simulador** em `dx-simulator/config.json`:
   - `api_key`: a API Key que você copiou.
   - `endpoint`: URL do ingest do seu NXD local, por exemplo  
     `http://localhost:8081/api/ingest` (troque a porta se o seu backend usar outra).
   - `update_rate_s`: intervalo em segundos (ex.: 10).
   - `clps`: lista de CLPs simulados (device_id, brand, protocol).
3. **Rodar o simulador** (a partir da pasta do simulador, para ele achar o `config.json`):
   ```bash
   cd dx-simulator
   go run .
   ```
   Ou, na raiz do repositório:
   ```bash
   go run ./dx-simulator
   ```
   Nesse caso, o programa procura `config.json` no **diretório de trabalho atual**; coloque lá uma cópia do `config.json` ou ajuste o código para um caminho fixo, se preferir.

4. O simulador envia POST no formato do NXD a cada `update_rate_s` segundos. No Dashboard você deve ver os ativos e as métricas atualizando.

**Docker:** O `docker-compose` tem o serviço `dx-simulator` que usa `Dockerfile.simulator` e variável `API_KEY`. O `Dockerfile.simulator` hoje referencia `simulator/` (caminho antigo). Se a sua árvore real é `dx-simulator/`, o compose precisa apontar para ela (ou você roda o simulador manualmente com `go run` como acima).

---

## 4. O sistema aceita “qualquer” dado do DX? Quais os mais comuns?

### O que o NXD aceita (contrato do ingest)

- **Endpoint:** `POST /api/ingest`
- **Corpo (JSON):**
  - `api_key` (string)
  - `device_id` (string)
  - `brand`, `protocol` (opcionais)
  - `timestamp` (opcional; se omitido, usa o horário do servidor)
  - `tags`: um objeto em que **cada chave é o nome da tag (métrica)** e **cada valor precisa ser numérico** (número ou string que converta para número).

Ou seja: **qualquer nome de tag** (ex.: `Temperatura_Molde`, `Total_Pecas`, `MyCustom_KPI`) é aceito. O que **não** é aceito é valor não numérico (ex.: string livre); essas tags são ignoradas e aparece aviso no log.

### Tipos de dados mais comuns na indústria (e que o NXD cobre)

| Categoria | Exemplos de tags / métricas | Uso no NXD |
|-----------|----------------------------|------------|
| **Contadores** | Peças OK, peças NOK, total produzido, pulsos | Número (inteiro ou decimal). Para indicadores financeiros: mapear em “tag OK” / “tag NOK” e usar delta ou absoluto. |
| **Temperatura / pressão / vazão** | Temperatura molde, temperatura motor, pressão, vazão | Valor numérico (ex.: °C, bar, L/min). |
| **Status / running** | Máquina ligada (0/1), running, falha | 0/1 ou código numérico. Para “horas parada”, mapear como tag status (ex.: 0 = parado, 1 = rodando). |
| **Energia** | Consumo kWh, corrente, potência | Número. |
| **Tempos** | Cycle time (ms), tempo de ciclo, tempo parada | Número (ex.: milissegundos ou segundos). |
| **Qualidade / saúde** | Health score, OEE parcial, fault code | Número (ex.: 0–1 para score, código numérico para falha). |
| **Custom / KPIs** | Qualquer KPI numérico que o DX envie | Qualquer chave em `tags` com valor numérico. |

Resumo: **o sistema lê qualquer dado que o DX envie em `tags` com valor numérico.** Os “mais comuns” acima são os típicos em chão de fábrica; o NXD não fixa uma lista — você pode usar os nomes que o seu DX/CLP enviar e, para relatórios financeiros, configurar o mapeamento (tag OK, NOK, status) na tela de Indicadores financeiros.

---

## 5. Checklist rápido para testar tudo local

1. [ ] Backend no ar (`go run .` ou `.\server.exe`), frontend no ar (`cd web-app && npm run dev`).
2. [ ] Registrar → Login → Onboarding; guardar API Key.
3. [ ] Colocar API Key e `endpoint` local em `dx-simulator/config.json`; rodar `go run .` em `dx-simulator` (ou `go run ./dx-simulator` na raiz com config no cwd).
4. [ ] Abrir Dashboard e conferir ativos/métricas.
5. [ ] (Opcional) Se tiver GCP: configurar credenciais e testar o chat/relatório na IA.
6. [ ] Testar Importar histórico (job memory com JSON; ou dx_http se tiver uma URL de histórico).
7. [ ] Configurar Indicadores financeiros (config. negócio + mapeamento de tags) e conferir os cards.

Com isso você cobre a funcionalidade real que deve esperar do serviço local, incluindo simulador DX e tipos de dados mais comuns.
