# HUB System v1.0.0

## 🎯 O que é o HUB System?

O **HUB System** é uma infraestrutura de inteligência industrial que atua como uma camada de software entre o chão de fábrica e a gestão. Ele não depende de marca ou modelo de máquina, pois utiliza o **Módulo DX (Delta)** como gateway universal para traduzir dados de CLPs e sensores em informações web em tempo real.

## 🚀 Início Rápido

### Requisitos

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Windows 10/11** (ou Linux/Mac com adaptações)
- **Navegador Web** moderno (Chrome, Firefox, Edge)

### Instalação e Execução

1. **Clone ou extraia o projeto** para uma pasta (ex: `C:\HubSystem1.0`)

2. **Execute o sistema** com um único comando:
   ```bash
   START.bat
   ```

3. **O que acontece automaticamente:**
   - ✅ Verifica dependências (Go)
   - ✅ Cria estrutura de diretórios
   - ✅ Baixa dependências do Go
   - ✅ Compila servidor HUB System
   - ✅ Compila simulador DX
   - ✅ Inicia servidor em `http://localhost:8080`
   - ✅ Cria fábrica de teste
   - ✅ Inicia simulador DX enviando dados
   - ✅ Abre Dashboard no navegador

4. **Pronto!** Você verá os dados das máquinas chegando em tempo real no Dashboard.

### Outros Comandos

- **Parar o sistema:**
  ```bash
  STOP.bat
  ```

- **Limpar e reiniciar do zero:**
  ```bash
  CLEAN.bat
  ```

## 📁 Estrutura do Projeto

```
HubSystem1.0/
├── core/                   # Lógica de negócio e modelos
│   ├── models.go          # Estruturas de dados
│   └── security.go        # Geração e validação de API Keys
├── api/                   # Endpoints e handlers
│   ├── handlers.go        # Rotas HTTP
│   └── websocket.go       # WebSocket para tempo real
├── data/                  # Camada de dados
│   ├── database.go        # SQLite + Auto-discovery
│   └── hubsystem.db       # Banco de dados (gerado)
├── services/              # Serviços futuros (vazios por enquanto)
│   └── .gitkeep
├── simulator/             # Simulador DX Gateway
│   └── dx_simulator.go    # Simula CLPs Siemens + Delta
├── web/                   # Dashboard Web
│   ├── index.html         # Interface principal
│   ├── style.css          # Estilos
│   └── app.js             # Lógica frontend
├── logs/                  # Logs do sistema
├── docs/                  # Documentação
│   ├── README.md          # Este arquivo
│   ├── API.md             # Documentação da API
│   └── FAQ.md             # 50+ Perguntas e Respostas
├── main.go                # Ponto de entrada do servidor
├── go.mod                 # Dependências Go
├── START.bat              # Inicialização automática
├── STOP.bat               # Parar sistema
└── CLEAN.bat              # Limpeza completa
```

## 🔑 Como Funciona

### 1. Fluxo de Dados

```
[CLP Siemens/Delta] 
    ↓ (S7/Modbus)
[DX Gateway] 
    ↓ (HTTP/JSON via 4G)
[HUB System API] 
    ↓ (Auto-discovery + SQLite)
[Dashboard Web] 
    ↓ (WebSocket)
[Usuário Final]
```

### 2. Auto-Discovery (O Segredo)

Quando o DX envia dados, o sistema:

1. **Valida a API Key** da fábrica
2. **Busca ou cria a máquina** automaticamente (pelo `device_id`)
3. **Para cada tag nova** (ex: `Pressao_Vapor`):
   - Cria um campo no banco de dados
   - Detecta o tipo de dado (int, float, bool, string)
   - Registra no log de auditoria
4. **Armazena o valor** com timestamp
5. **Envia atualização** via WebSocket para o Dashboard

**Nenhuma intervenção humana necessária!**

### 3. Simulador de Rede 4G/LTE

O simulador DX replica condições reais:

- ✅ **Latência variável** (20-800ms baseada na qualidade do sinal)
- ✅ **Quedas de conexão** aleatórias (5% de chance)
- ✅ **Sistema de retry** automático (até 3 tentativas)
- ✅ **Monitoramento de sinal** (Excelente, Bom, Regular, Fraco)
- ✅ **Buffering** (aguarda reconexão antes de descartar dados)

## 🌐 API Endpoints

### Health Check
```http
GET /api/health
```

### Criar Fábrica
```http
POST /api/factories
Content-Type: application/json

{
  "name": "Fábrica São Paulo"
}
```

**Resposta:**
```json
{
  "id": 1,
  "name": "Fábrica São Paulo",
  "api_key": "HUB_a1b2c3d4e5f6...",
  "created_at": "2026-02-12T10:30:00Z",
  "is_active": true
}
```

### Ingestão de Dados (DX → HUB)
```http
POST /api/ingest
Content-Type: application/json

{
  "api_key": "HUB_a1b2c3d4e5f6...",
  "device_id": "DX_FACTORY_001",
  "brand": "DX_GATEWAY",
  "protocol": "MULTI_PROTOCOL",
  "timestamp": "2026-02-12T10:30:00Z",
  "tags": {
    "SIEMENS_Pressao_Vapor": 12.5,
    "SIEMENS_Temperatura_Motor": 75.3,
    "DELTA_Pressao_Hidraulica": 180.2,
    "DELTA_RPM_Motor_Principal": 2450
  }
}
```

### Listar Máquinas
```http
GET /api/machines
X-API-Key: HUB_a1b2c3d4e5f6...
```

### WebSocket (Tempo Real)
```
ws://localhost:8080/ws
```

## 🔐 Segurança

- **API Keys** geradas com 256 bits de entropia
- **Validação** em todas as requisições de ingestão
- **Isolamento** por fábrica (uma fábrica não vê dados de outra)
- **Logs de auditoria** de todas as ações
- **IP tracking** para rastreabilidade

## 📊 Dashboard

O Dashboard oferece:

- ✅ **Visão em tempo real** de todas as máquinas
- ✅ **Status online/offline** com indicador visual
- ✅ **Últimos valores** de todas as tags
- ✅ **Atualização automática** via WebSocket
- ✅ **Estatísticas** (máquinas conectadas, tags monitoradas)
- ✅ **Responsivo** (funciona em desktop e mobile)

## 🛠️ Tecnologias

- **Backend:** Go 1.21+ (Gorilla Mux, WebSocket)
- **Banco de Dados:** SQLite (embedded, sem servidor externo)
- **Frontend:** HTML5 + CSS3 + JavaScript (Vanilla)
- **Protocolos:** HTTP/REST + WebSocket
- **Simulação:** Go (CLPs Siemens S7 + Delta Modbus)

## 📈 Próximos Passos (Roadmap)

### Fase 2 - Alertas e Notificações
- [ ] Sistema de alertas por threshold
- [ ] Notificações via Email
- [ ] Notificações via WhatsApp (API)
- [ ] Alertas sonoros no Dashboard

### Fase 3 - Relatórios e Exportação
- [ ] Exportar dados para Excel/CSV
- [ ] Relatórios de OEE (Eficiência Global)
- [ ] Comparação de turnos
- [ ] Análise de micro-paradas

### Fase 4 - Integração e Expansão
- [ ] API para PowerBI
- [ ] Integração com ERPs
- [ ] Manutenção preditiva (ML)
- [ ] Suporte a mais protocolos (OPC UA, MQTT)

## ❓ FAQ

Consulte [FAQ.md](FAQ.md) para as **50+ perguntas mais frequentes** e suas respostas.

## 📞 Suporte

- **Logs:** Verifique `logs/` para diagnóstico
- **Banco de Dados:** `data/hubsystem.db` (use SQLite Browser)
- **Console:** Janelas do CMD mostram logs em tempo real

## 📄 Licença

Projeto proprietário - HUB System v1.0.0

---

**Desenvolvido com ❤️ para a Indústria 4.0**
