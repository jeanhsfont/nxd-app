# API Documentation - HUB System

## Base URL

```
http://localhost:8080/api
```

Em produção, substitua por `https://seu-dominio.com/api`

---

## Endpoints

### 1. Health Check

Verifica se o servidor está funcionando.

**Request:**
```http
GET /api/health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-02-12T10:30:00Z",
  "version": "1.0.0"
}
```

**Status Codes:**
- `200 OK` - Servidor funcionando

---

### 2. Criar Fábrica

Cria uma nova fábrica e gera uma API Key única.

**Request:**
```http
POST /api/factories
Content-Type: application/json

{
  "name": "Fábrica São Paulo"
}
```

**Response:**
```json
{
  "id": 1,
  "name": "Fábrica São Paulo",
  "api_key": "HUB_a1b2c3d4e5f6789012345678901234567890123456789012345678901234",
  "created_at": "2026-02-12T10:30:00Z",
  "is_active": true
}
```

**Status Codes:**
- `200 OK` - Fábrica criada com sucesso
- `400 Bad Request` - Nome da fábrica não fornecido
- `500 Internal Server Error` - Erro ao criar fábrica

**Notas:**
- Guarde a `api_key` com segurança
- Esta chave será usada pelo DX Gateway para enviar dados

---

### 3. Ingestão de Dados

Endpoint usado pelo DX Gateway para enviar dados das máquinas.

**Request:**
```http
POST /api/ingest
Content-Type: application/json

{
  "api_key": "HUB_a1b2c3d4e5f6789012345678901234567890123456789012345678901234",
  "device_id": "DX_FACTORY_001",
  "brand": "DX_GATEWAY",
  "protocol": "MULTI_PROTOCOL",
  "timestamp": "2026-02-12T10:30:00.123Z",
  "tags": {
    "SIEMENS_Pressao_Vapor": 12.5,
    "SIEMENS_Temperatura_Motor": 75.3,
    "SIEMENS_Velocidade_Esteira": 120,
    "SIEMENS_Status_Operacional": true,
    "DELTA_Pressao_Hidraulica": 180.2,
    "DELTA_RPM_Motor_Principal": 2450,
    "DELTA_Modo_Operacao": "AUTO"
  }
}
```

**Response:**
```json
{
  "status": "success",
  "machine_id": 1,
  "tags_count": 7
}
```

**Status Codes:**
- `200 OK` - Dados processados com sucesso
- `400 Bad Request` - Payload inválido
- `401 Unauthorized` - API Key inválida ou fábrica inativa
- `500 Internal Server Error` - Erro ao processar dados

**Notas:**
- **Auto-discovery:** Se a máquina (`device_id`) não existir, será criada automaticamente
- **Tags dinâmicas:** Novas tags são detectadas e criadas automaticamente
- **Tipos de dados:** Suporta `int`, `float`, `bool`, `string`
- **Timestamp:** Opcional. Se omitido, usa o timestamp do servidor

---

### 4. Listar Máquinas

Retorna todas as máquinas de uma fábrica com seus últimos valores.

**Request:**
```http
GET /api/machines
X-API-Key: HUB_a1b2c3d4e5f6789012345678901234567890123456789012345678901234
```

**Response:**
```json
[
  {
    "id": 1,
    "factory_id": 1,
    "device_id": "DX_FACTORY_001",
    "name": "DX_GATEWAY_DX_FACTORY_001",
    "brand": "DX_GATEWAY",
    "protocol": "MULTI_PROTOCOL",
    "last_seen": "2026-02-12T10:30:00Z",
    "is_online": true,
    "created_at": "2026-02-12T10:00:00Z",
    "latest_data": [
      {
        "id": 1234,
        "machine_id": 1,
        "tag_name": "SIEMENS_Pressao_Vapor",
        "value": "12.5",
        "timestamp": "2026-02-12T10:30:00Z"
      },
      {
        "id": 1235,
        "machine_id": 1,
        "tag_name": "SIEMENS_Temperatura_Motor",
        "value": "75.3",
        "timestamp": "2026-02-12T10:30:00Z"
      }
    ]
  }
]
```

**Status Codes:**
- `200 OK` - Máquinas retornadas com sucesso
- `401 Unauthorized` - API Key não fornecida ou inválida
- `500 Internal Server Error` - Erro ao buscar máquinas

**Notas:**
- `latest_data` contém o último valor de cada tag
- `is_online` é `true` se a máquina enviou dados nos últimos 60 segundos

---

### 5. WebSocket (Tempo Real)

Conexão WebSocket para receber atualizações em tempo real.

**Connection:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('Conectado ao HUB System');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Atualização recebida:', data);
};
```

**Message Format:**
```json
{
  "type": "data_update",
  "factory_id": 1,
  "machine_id": 1,
  "device_id": "DX_FACTORY_001",
  "timestamp": "2026-02-12T10:30:00Z",
  "tags": {
    "SIEMENS_Pressao_Vapor": 12.5,
    "SIEMENS_Temperatura_Motor": 75.3
  }
}
```

**Notas:**
- Todos os clientes conectados recebem atualizações quando dados chegam
- Reconexão automática recomendada (veja `web/app.js` para exemplo)

---

## Autenticação

### API Key

A API Key deve ser incluída de duas formas, dependendo do endpoint:

1. **No corpo da requisição** (para `/api/ingest`):
   ```json
   {
     "api_key": "HUB_...",
     ...
   }
   ```

2. **No header** (para `/api/machines`):
   ```http
   X-API-Key: HUB_...
   ```

### Formato da API Key

- Prefixo: `HUB_`
- Comprimento: 68 caracteres (4 + 64 hex)
- Exemplo: `HUB_a1b2c3d4e5f6789012345678901234567890123456789012345678901234`

---

## Rate Limiting

Atualmente não há rate limiting, mas será implementado em produção:

- **Ingest:** 1000 requisições/minuto por API Key
- **Queries:** 100 requisições/minuto por API Key

---

## Logs de Auditoria

Todas as ações são registradas no banco de dados:

```sql
SELECT * FROM audit_logs ORDER BY timestamp DESC LIMIT 10;
```

**Campos:**
- `action`: Tipo de ação (ex: `INGEST`)
- `api_key`: API Key usada
- `device_id`: ID do dispositivo
- `status`: `SUCCESS` ou `FAIL`
- `message`: Detalhes da ação
- `ip_address`: IP de origem
- `timestamp`: Quando ocorreu

---

## Erros Comuns

### 401 Unauthorized
```json
{
  "error": "API Key inválida ou fábrica inativa"
}
```
**Solução:** Verifique se a API Key está correta e a fábrica está ativa.

### 400 Bad Request
```json
{
  "error": "Payload inválido"
}
```
**Solução:** Verifique o formato JSON e campos obrigatórios.

### 500 Internal Server Error
```json
{
  "error": "Erro ao processar dados"
}
```
**Solução:** Verifique os logs do servidor em `logs/` ou console.

---

## Exemplos de Uso

### cURL

**Criar fábrica:**
```bash
curl -X POST http://localhost:8080/api/factories \
  -H "Content-Type: application/json" \
  -d '{"name":"Fábrica Teste"}'
```

**Enviar dados:**
```bash
curl -X POST http://localhost:8080/api/ingest \
  -H "Content-Type: application/json" \
  -d '{
    "api_key": "HUB_...",
    "device_id": "DX_001",
    "brand": "Siemens",
    "protocol": "S7",
    "tags": {
      "Temperatura": 75.5,
      "Pressao": 120
    }
  }'
```

**Listar máquinas:**
```bash
curl -X GET http://localhost:8080/api/machines \
  -H "X-API-Key: HUB_..."
```

### PowerShell

**Criar fábrica:**
```powershell
$body = @{
    name = "Fábrica Teste"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/factories" `
  -Method Post `
  -ContentType "application/json" `
  -Body $body
```

### Python

**Enviar dados:**
```python
import requests
import json

payload = {
    "api_key": "HUB_...",
    "device_id": "DX_001",
    "brand": "Siemens",
    "protocol": "S7",
    "tags": {
        "Temperatura": 75.5,
        "Pressao": 120
    }
}

response = requests.post(
    "http://localhost:8080/api/ingest",
    json=payload
)

print(response.json())
```

---

## Próximas Versões da API

### v1.1 (Fase 2)
- `POST /api/alerts` - Criar alertas
- `GET /api/alerts` - Listar alertas
- `PUT /api/machines/{id}` - Atualizar nome/configuração de máquina

### v1.2 (Fase 3)
- `GET /api/reports/oee` - Calcular OEE
- `GET /api/export/csv` - Exportar dados para CSV
- `GET /api/export/xlsx` - Exportar dados para Excel

### v2.0 (Fase 4)
- GraphQL endpoint
- Webhooks para eventos
- API de manutenção preditiva

---

**Última atualização:** 12/02/2026
