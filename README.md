# 🏭 NXD v1.0 - Nexus Data Exchange

**Inteligência Industrial em Tempo Real com Docker**

Sistema de monitoramento industrial que conecta máquinas de qualquer marca a um dashboard centralizado.

## 🚀 Início Rápido

### Pré-requisitos
- Docker Desktop instalado e **rodando** (ícone verde)
- Navegador moderno

### Primeira Execução (4 passos)

#### 1. Iniciar Sistema
```bash
# Windows: Clique duas vezes em
START.bat

# Linux/Mac:
docker-compose build
docker-compose up -d nxd-server
```

⏳ **Aguarde 30-60 segundos** (primeira vez demora mais)

#### 2. Criar Fábrica
O navegador abrirá automaticamente em http://localhost:8080

1. Digite o nome da sua fábrica
2. Clique em "Criar Fábrica"
3. **COPIE** a API Key gerada (68 caracteres)

#### 3. Configurar Simulador
Crie arquivo `.env` na raiz do projeto:
```env
API_KEY=HUB_abc123def456...
```

💡 **Dica:** Use o Notepad ou qualquer editor de texto

#### 4. Reiniciar com Simulador
```bash
# Execute START.bat novamente
START.bat
```

⏳ **Aguarde 10-15 segundos** para as máquinas aparecerem

---

### Execuções Seguintes (1 passo)

Depois da primeira configuração:
```bash
START.bat  # Só isso!
```

## 📊 O Que Você Vai Ver

Após 5-10 segundos:
- 2 máquinas (Siemens S7 + Delta Modbus)
- 6 tags por máquina
- Valores atualizando em tempo real

## 🐳 Comandos Docker

```bash
# Iniciar tudo
docker-compose up -d

# Ver logs
docker-compose logs -f

# Ver status
docker-compose ps

# Parar tudo
docker-compose down

# Reconstruir
docker-compose build --no-cache
```

## 📁 Estrutura

```
HubSystem1.0/
├── docker-compose.yml          # Orquestração
├── Dockerfile.hub              # Container servidor
├── Dockerfile.simulator        # Container simulador
├── START.bat                   # Inicia sistema
├── STOP.bat                    # Para sistema
├── core/                       # Lógica de negócio
├── api/                        # Endpoints REST
├── data/                       # Banco de dados
├── services/                   # Serviços
├── simulator/                  # Simulador DX
├── web/                        # Dashboard
└── docs/                       # Documentação
```

## 🔧 Características

- ✅ Auto-Discovery de máquinas e tags
- ✅ Multi-marca (Siemens, Delta, Mitsubishi...)
- ✅ Tempo real (2-3 segundos)
- ✅ Simulação de rede 4G/LTE
- ✅ API REST completa
- ✅ WebSocket para updates
- ✅ Isolamento com Docker

## 📡 API Endpoints

- `GET  /api/health` - Health check
- `POST /api/factory/create` - Criar fábrica
- `POST /api/ingest` - Receber dados do DX
- `GET  /api/dashboard?api_key=XXX` - Dados do dashboard

## 🆘 Troubleshooting

### ❌ "Unable to connect" no navegador
**Causa:** Servidor ainda não terminou de inicializar  
**Solução:** Aguarde mais 30 segundos e atualize a página (F5)

### ❌ Docker não está rodando
**Solução:** Inicie o Docker Desktop e aguarde o ícone ficar verde

### ❌ Porta 8080 em uso
```bash
docker-compose down
# Ou mude a porta em docker-compose.yml
```

### ❌ Máquinas não aparecem
1. Verifique o arquivo `.env` com API Key correta
2. Aguarde 15 segundos e atualize (F5)
3. Veja os logs: `docker-compose logs dx-simulator`

### ❌ Compilação demora muito
**É normal na primeira vez!** Pode demorar 2-5 minutos.

---

📚 **Guia completo:** [TROUBLESHOOTING.md](TROUBLESHOOTING.md)

## 📚 Documentação

- [Arquitetura](docs/ARQUITETURA.md)
- [Manual de Instalação](docs/MANUAL_INSTALACAO.md)
- [FAQ](docs/FAQ.md)

## 📄 Licença

Copyright © 2026 HUB System. Todos os direitos reservados.
