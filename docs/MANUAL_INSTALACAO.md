# 📘 Manual de Instalação - HUB System

## Pré-requisitos

### Windows
- Windows 10/11
- Go 1.21 ou superior
- Navegador moderno (Chrome, Firefox, Edge)

### Instalação do Go

1. Baixe o instalador: https://go.dev/dl/
2. Execute o instalador
3. Verifique a instalação:
```bash
go version
```

## Instalação Rápida

### 1. Clone ou Baixe o Projeto

```bash
cd C:\
git clone [URL_DO_REPOSITORIO] HubSystem1.0
cd HubSystem1.0
```

### 2. Execute o Script de Inicialização

```bash
START_HUB.bat
```

O script irá:
- ✓ Verificar dependências
- ✓ Criar estrutura de pastas
- ✓ Baixar bibliotecas Go
- ✓ Compilar servidor e simulador
- ✓ Iniciar servidor HUB
- ✓ Abrir dashboard no navegador

## Primeiro Uso

### 1. Criar Fábrica

No dashboard aberto automaticamente:
1. Digite o nome da sua fábrica
2. Clique em "Criar Fábrica"
3. **COPIE A API KEY** gerada (você precisará dela!)

### 2. Iniciar Simulador DX

Em um novo terminal:
```bash
START_DX_SIMULATOR.bat [SUA_API_KEY]
```

Exemplo:
```bash
START_DX_SIMULATOR.bat HUB_abc123def456...
```

### 3. Visualizar Dados

Volte ao dashboard e veja os dados chegando em tempo real!

## Estrutura de Pastas

```
HubSystem1.0/
├── core/              # Lógica de negócio
├── api/               # Endpoints HTTP
├── data/              # Banco de dados
├── services/          # Serviços auxiliares
├── simulator/         # Simulador DX
├── web/               # Dashboard
├── logs/              # Logs do sistema
├── docs/              # Documentação
├── START_HUB.bat      # Inicia sistema
├── START_DX_SIMULATOR.bat  # Inicia simulador
└── STOP_ALL.bat       # Encerra tudo
```

## Endpoints da API

- `http://localhost:8080` - Dashboard
- `http://localhost:8080/api/health` - Health check
- `http://localhost:8080/api/ingest` - Ingestão de dados
- `http://localhost:8080/api/factory/create` - Criar fábrica
- `http://localhost:8080/api/dashboard?api_key=XXX` - Dados do dashboard

## Troubleshooting

### Erro: "Go não está instalado"
- Instale o Go: https://go.dev/dl/
- Reinicie o terminal após instalação

### Erro: "Porta 8080 em uso"
- Feche outras aplicações usando a porta
- Ou edite `main.go` para mudar a porta

### Dashboard não carrega
- Verifique se o servidor está rodando
- Acesse: http://localhost:8080/api/health
- Verifique logs em `logs/`

### Simulador não conecta
- Verifique se a API Key está correta
- Verifique se o servidor HUB está rodando
- Veja logs do simulador

## Parar o Sistema

```bash
STOP_ALL.bat
```

Ou feche as janelas dos processos.

## Logs

Logs são salvos em:
- `logs/hub_YYYY-MM-DD.log`

## Backup

Para fazer backup dos dados:
1. Pare o sistema (`STOP_ALL.bat`)
2. Copie a pasta `data/`
3. Copie a pasta `logs/` (opcional)

## Próximos Passos

- Configure alertas personalizados
- Exporte relatórios
- Integre com seu ERP
- Configure múltiplas fábricas
