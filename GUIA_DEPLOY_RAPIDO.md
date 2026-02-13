# 🚀 GUIA RÁPIDO - Deploy NXD no Google Cloud

**Projeto:** nxdata-487304  
**Região:** São Paulo (southamerica-east1)

---

## ⚡ DEPLOY EM 3 PASSOS

### **PASSO 1: Fazer Deploy no Google Cloud**

```bash
# Execute:
DEPLOY_NOW.bat
```

**O que vai acontecer:**
1. ✅ Verifica autenticação do gcloud
2. ✅ Ativa APIs necessárias (Compute Engine, IAM, etc)
3. ✅ Cria VM no Google Cloud (e2-micro - GRÁTIS)
4. ✅ Instala Docker na VM
5. ✅ Faz upload do código NXD
6. ✅ Builda e inicia os containers
7. ✅ Configura firewall
8. ✅ **Te dá a URL final**: `http://IP-DA-VM:8080`

**Tempo estimado:** 2-3 minutos

---

### **PASSO 2: Criar Fábrica e Obter API Key**

1. Acesse a URL fornecida: `http://IP-DA-VM:8080`
2. Digite o nome da fábrica (ex: "Vale Sistemas")
3. Clique em **"Criar Fábrica"**
4. **COPIE A API KEY** (formato: `NXD_xxxxx...`)

---

### **PASSO 3: Conectar Simulador Local**

```bash
# Execute:
CONECTAR_SIMULADOR.bat
```

**Vai pedir:**
1. IP do servidor (o que você recebeu no Passo 1)
2. API Key (a que você copiou no Passo 2)

**Pronto!** O simulador no seu PC vai começar a enviar dados para o NXD na nuvem!

---

## 📊 MONITORAMENTO

### Ver dados em tempo real
```
Acesse: http://IP-DA-VM:8080
```

### Ver logs do servidor
```bash
gcloud compute ssh nxd-server-vm --zone=southamerica-east1-b --project=nxdata-487304 --command='sudo docker-compose logs -f nxd-server'
```

### Ver logs do simulador (local)
```bash
docker-compose logs -f dx-simulator
```

---

## 🔧 COMANDOS ÚTEIS

### Parar simulador local
```bash
docker-compose stop dx-simulator
```

### Iniciar simulador local
```bash
docker-compose up -d dx-simulator
```

### Reiniciar servidor na nuvem
```bash
gcloud compute ssh nxd-server-vm --zone=southamerica-east1-b --project=nxdata-487304 --command='cd /opt/nxd && sudo docker-compose restart'
```

### Ver status da VM
```bash
gcloud compute instances list --project=nxdata-487304
```

---

## 🆘 TROUBLESHOOTING

### Simulador não conecta
1. Verifique se o IP está correto no `.env`
2. Teste: `curl http://IP-DA-VM:8080/api/health`
3. Verifique firewall no Google Cloud

### Dashboard não carrega
1. Aguarde 30 segundos após deploy
2. Verifique se a VM está rodando: `gcloud compute instances list`
3. Acesse os logs: `docker-compose logs nxd-server`

### API Key inválida
1. Verifique se copiou a chave completa (começa com `NXD_`)
2. Crie uma nova fábrica se necessário

---

## 💰 CUSTOS

- **VM e2-micro**: **GRÁTIS** (always free tier)
- **Disco 10GB**: ~$2/mês
- **Tráfego**: Incluído nos $300 de crédito

**Total: $0-2/mês** (com créditos = GRÁTIS por meses!)

---

## 🎯 AMANHÃ NA VALE SISTEMAS

### Preparação:
1. ✅ NXD rodando na nuvem
2. ✅ Dashboard acessível
3. ✅ API Key gerada

### No local:
1. Configure o DX real com:
   - **Endpoint**: `http://IP-DA-VM:8080/api/ingest`
   - **API Key**: (a que você gerou)
2. Acesse o dashboard para ver os dados em tempo real
3. 🎉 **SUCESSO!**

---

**Desenvolvido para NXD (Nexus Data Exchange)** 🏭
