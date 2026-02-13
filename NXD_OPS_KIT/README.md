# 🚀 NXD OPS KIT

Kit de operações para deploy e gerenciamento do **NXD (Nexus Data Exchange)** no Google Cloud.

---

## 📂 Estrutura

```
NXD_OPS_KIT/
├── config/
│   └── ops.env.example    # Template de configuração
├── ps/
│   ├── _lib.ps1           # Biblioteca comum
│   └── deploy_nxd_vm.ps1  # Deploy em VM (Compute Engine)
└── README.md              # Este arquivo
```

---

## 🎯 Quick Start

### 1️⃣ Pré-requisitos

- **Google Cloud SDK** instalado: https://cloud.google.com/sdk/docs/install
- **Conta no Google Cloud** com projeto criado
- **Docker** (para testes locais)

### 2️⃣ Configuração

```bash
# Copie o template de configuração
cd NXD_OPS_KIT/config
copy ops.env.example ops.env

# Edite ops.env com suas configurações:
# - PROJECT_ID: ID do seu projeto no Google Cloud
# - REGION: Região (recomendado: southamerica-east1 para São Paulo)
# - ALLOWED_CLIENT_IP: Seu IP público (para segurança)
```

### 3️⃣ Deploy

**Opção A: Script Simplificado (Recomendado)**
```bash
# Na raiz do projeto
DEPLOY_CLOUD.bat
```

**Opção B: PowerShell Direto**
```powershell
# Autenticar no Google Cloud
gcloud auth login

# Deploy em VM
.\NXD_OPS_KIT\ps\deploy_nxd_vm.ps1 -ProjectId "seu-projeto-id"

# Com IP restrito (recomendado)
.\NXD_OPS_KIT\ps\deploy_nxd_vm.ps1 -ProjectId "seu-projeto-id" -AllowedIP "seu.ip.publico"
```

---

## 🔧 Scripts Disponíveis

### `deploy_nxd_vm.ps1`
Deploy completo do NXD em VM do Google Compute Engine.

**Parâmetros:**
- `-ProjectId`: ID do projeto Google Cloud
- `-Zone`: Zona da VM (padrão: southamerica-east1-b)
- `-MachineType`: Tipo de máquina (padrão: e2-micro - grátis)
- `-DiskSize`: Tamanho do disco (padrão: 10GB)
- `-AllowedIP`: IP autorizado no firewall (opcional, recomendado)
- `-SkipTests`: Pula testes locais

**Exemplo:**
```powershell
.\ps\deploy_nxd_vm.ps1 `
    -ProjectId "nxd-production" `
    -Zone "southamerica-east1-b" `
    -AllowedIP "200.123.45.67"
```

---

## 💰 Custos Estimados

| Recurso | Especificação | Custo/Mês |
|---------|---------------|-----------|
| VM e2-micro | 0.25 vCPU, 1GB RAM | **GRÁTIS** (always free) |
| VM e2-small | 0.5 vCPU, 2GB RAM | ~$13 USD |
| Disco 10GB | SSD persistente | ~$2 USD |
| IP Estático | 1 endereço | ~$3 USD |

**Com $300 de crédito grátis = 6+ meses sem custo!**

---

## 🔒 Segurança

### Firewall
O script configura automaticamente uma regra de firewall:
- **Com `-AllowedIP`**: Apenas o IP especificado pode acessar
- **Sem `-AllowedIP`**: Acesso público (⚠️ NÃO recomendado)

### API Key
- Gerada automaticamente no primeiro acesso
- Formato: `NXD_[64 caracteres hexadecimais]`
- Necessária para enviar dados ao NXD

---

## 📊 Monitoramento

### Ver logs em tempo real
```bash
gcloud compute ssh nxd-server-vm \
    --zone=southamerica-east1-b \
    --project=seu-projeto-id \
    --command='sudo docker-compose logs -f'
```

### Ver status dos containers
```bash
gcloud compute ssh nxd-server-vm \
    --zone=southamerica-east1-b \
    --project=seu-projeto-id \
    --command='sudo docker-compose ps'
```

### Reiniciar serviço
```bash
gcloud compute ssh nxd-server-vm \
    --zone=southamerica-east1-b \
    --project=seu-projeto-id \
    --command='cd /opt/nxd && sudo docker-compose restart'
```

---

## 🆘 Troubleshooting

### Erro: "gcloud not found"
```bash
# Instale o Google Cloud SDK
https://cloud.google.com/sdk/docs/install
```

### Erro: "gcloud not authenticated"
```bash
gcloud auth login
```

### Erro: "Project not found"
```bash
# Verifique se o projeto existe
gcloud projects list

# Configure o projeto padrão
gcloud config set project SEU-PROJECT-ID
```

### VM não responde
```bash
# Verifique se a VM está rodando
gcloud compute instances list --project=seu-projeto-id

# Reinicie a VM
gcloud compute instances stop nxd-server-vm --zone=southamerica-east1-b --project=seu-projeto-id
gcloud compute instances start nxd-server-vm --zone=southamerica-east1-b --project=seu-projeto-id
```

---

## 📝 Notas

- **Primeira execução**: Pode demorar 2-3 minutos (download de imagens Docker)
- **Atualizações**: Execute o script novamente para atualizar o código
- **Dados persistentes**: Armazenados em `/opt/nxd/data` na VM
- **Logs**: Armazenados em `/opt/nxd/logs` na VM

---

## 🔗 Links Úteis

- **Google Cloud Console**: https://console.cloud.google.com/
- **Documentação NXD**: ../README.md
- **Suporte**: Entre em contato com o time de desenvolvimento

---

**Desenvolvido para NXD (Nexus Data Exchange)** 🏭
