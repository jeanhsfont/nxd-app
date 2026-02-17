# NXD v2.0 - Especificação Técnica e Funcional

## 🏗️ Arquitetura (Serviço Único)

O NXD v2.0 opera como um monolito modular containerizado (`nxd-core`), combinando:
1.  **Backend:** Go (Golang) com Gorilla Mux.
2.  **Frontend:** React (Vite) servido estaticamente pelo Go.
3.  **Banco de Dados:** PostgreSQL (Cloud SQL) com suporte a JSONB.
4.  **IA:** Vertex AI (Gemini Pro) integrado via API.

## 📱 Mapa de Telas (Total: 12)

### 🔐 Grupo 1: Acesso e Segurança
1.  **Login:** Autenticação via E-mail/Senha (Firebase Auth).
2.  **Recuperação de Senha:** Fluxo de "Esqueci minha senha".
3.  **Nova Senha:** Redefinição segura.
4.  **Configuração MFA:** Setup de 2FA (Google Authenticator).

### ⚙️ Grupo 2: Setup
5.  **Minha Fábrica:** Cadastro inicial e geração de **API Key** (exibida uma única vez).
6.  **Configuração Financeira:** Inputs de custos (Energia/kWh, Hora/Homem).

### 🏭 Grupo 3: Operação
7.  **Dashboard Principal:** Visão geral de KPIs, status das máquinas e alertas críticos.
8.  **Gestão de Ativos (Implementada):** Drag & Drop para organizar máquinas em Setores.
9.  **Detalhes da Máquina:** Histórico de telemetria e gráficos de uma máquina específica.
10. **Logs de Auditoria:** Tabela imutável de ações de usuários e IA.

### 🧠 Grupo 4: Inteligência
11. **Chat IA & Reports:** Interface conversacional (RAG) para perguntas sobre dados da fábrica.
12. **Simulação de Cenários:** Ferramenta "What-If" para projeções financeiras.

## 🔌 API Endpoints (Principais)

### Gestão de Ativos
- `GET /api/groups`: Lista setores.
- `POST /api/groups`: Cria setor (suporta metadata: cor, ícone).
- `GET /api/assets`: Lista máquinas.
- `POST /api/assets/{id}/move`: Move máquina para um setor.

### Inteligência
- `POST /api/report/ia`: Envia prompt do usuário + contexto de dados para Vertex AI.

## 🗄️ Modelo de Dados (Destaques)

### Tabela `nxd.groups`
- `id`: UUID
- `name`: String
- `metadata`: JSONB (Ex: `{"color": "blue", "icon": "factory"}`)
- `parent_id`: UUID (Suporte a hierarquia)

### Tabela `nxd.assets`
- `group_id`: FK para `nxd.groups`
- `annotations`: JSONB (Metadados flexíveis da máquina)

## 🚀 Deploy

O sistema é entregue via imagem Docker única (`nxd-core`), hospedada no Google Cloud Run.
O frontend é compilado (`npm run build`) e os arquivos estáticos são embutidos na imagem final.
