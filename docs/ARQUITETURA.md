# 🏗️ Arquitetura do HUB System

## Visão Geral

O HUB System é uma plataforma de inteligência industrial que conecta máquinas de diferentes fabricantes a um sistema centralizado de monitoramento em tempo real.

## Componentes Principais

### 1. Core (`/core`)
- **types.go**: Definições de estruturas de dados
- **security.go**: Funções de segurança e validação

### 2. API (`/api`)
- **handlers.go**: Endpoints HTTP
  - `POST /api/ingest`: Recebe dados do DX
  - `POST /api/factory/create`: Cria nova fábrica
  - `GET /api/dashboard`: Retorna dados do dashboard
  - `GET /api/health`: Health check
- **websocket.go**: Comunicação em tempo real

### 3. Data (`/data`)
- **database.go**: Inicialização e migrações do SQLite
- **repository.go**: Operações de banco de dados

### 4. Services (`/services`)
- **logger.go**: Sistema de logs e auditoria
- **websocket_broadcaster.go**: Broadcast de atualizações
- **export_service.go**: Exportação de relatórios (futuro)
- **alert_service.go**: Sistema de alertas (futuro)
- **analytics_service.go**: Analytics e OEE (futuro)

### 5. Simulator (`/simulator`)
- **dx_simulator.go**: Simula módulo DX com:
  - Múltiplos CLPs (Siemens, Delta)
  - Protocolos industriais (S7, Modbus)
  - Condições de rede 4G/LTE
  - Latência variável e perda de pacotes

### 6. Web (`/web`)
- **index.html**: Interface do dashboard
- **style.css**: Estilos modernos
- **app.js**: Lógica do frontend

## Fluxo de Dados

```
[CLP Siemens/Delta] 
    ↓ (S7/Modbus)
[Módulo DX] 
    ↓ (4G/LTE - JSON via HTTP)
[HUB System API] 
    ↓ (Auto-Discovery)
[Banco de Dados SQLite] 
    ↓ (WebSocket)
[Dashboard Web]
```

## Auto-Discovery

O sistema detecta automaticamente:
1. Novas máquinas conectadas
2. Novas tags enviadas pelos CLPs
3. Tipos de dados (float, int, bool, string)

Não é necessário configuração prévia!

## Segurança

- API Key única por fábrica (64 caracteres hex)
- Validação de origem dos dados
- Logs de auditoria completos
- Isolamento de dados por fábrica

## Banco de Dados

### Tabelas Principais

- **factories**: Fábricas cadastradas
- **machines**: Máquinas conectadas
- **tags**: Pontos de dados (auto-discovery)
- **data_points**: Valores históricos
- **alerts**: Alertas configurados
- **audit_logs**: Logs de auditoria

## Performance

- SQLite com índices otimizados
- Conexões pooling
- WebSocket para updates em tempo real
- Cache de queries frequentes

## Expansão Futura

- Exportação de relatórios (Excel, PDF)
- Sistema de alertas (Email, WhatsApp)
- Cálculo de OEE automático
- Manutenção preditiva
- Integração com ERPs
- Multi-idiomas
