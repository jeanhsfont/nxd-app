# ☁️ DEPLOY NXD via Cloud Shell (Terminal no Navegador)

**✅ Sem instalar NADA no PC!**  
**✅ Tudo roda no navegador!**

---

## 🎯 **PASSO A PASSO SUPER SIMPLES:**

### **1️⃣ Abrir Cloud Shell**

```
https://console.cloud.google.com/
```

- Faça login com: **jeanhhirata@gmail.com**
- Selecione projeto: **nxdata-487304**
- Clique no ícone **`>_`** no canto superior direito

**Um terminal vai abrir no navegador!** 🎉

---

### **2️⃣ Fazer Upload do Código**

**No Cloud Shell, execute:**

```bash
# Criar diretório
mkdir -p ~/nxd
cd ~/nxd
```

**Agora faça upload dos arquivos:**

1. Clique nos **3 pontinhos** (⋮) no Cloud Shell
2. Clique em **"Upload"**
3. Selecione **TODOS os arquivos** da pasta `C:\HubSystem1.0`
   - (Exceto: `node_modules`, `.git`, `data`, `logs`, `NXD_OPS_KIT`)

**OU use este comando para clonar (se tiver Git):**
```bash
# Se o código estiver no GitHub
git clone https://github.com/SEU-USUARIO/nxd.git
cd nxd
```

---

### **3️⃣ Ativar APIs**

```bash
gcloud services enable compute.googleapis.com
gcloud services enable cloudresourcemanager.googleapis.com
gcloud services enable iam.googleapis.com
```

⏳ Aguarde 30 segundos

---

### **4️⃣ Criar VM**

```bash
gcloud compute instances create nxd-server-vm \
  --project=nxdata-487304 \
  --zone=southamerica-east1-b \
  --machine-type=e2-micro \
  --boot-disk-size=10GB \
  --boot-disk-type=pd-standard \
  --image-family=ubuntu-2204-lts \
  --image-project=ubuntu-os-cloud \
  --tags=nxd-server \
  --metadata=startup-script='#!/bin/bash
apt-get update
apt-get install -y docker.io docker-compose
systemctl enable docker
systemctl start docker
mkdir -p /opt/nxd'
```

⏳ Aguarde 1-2 minutos

---

### **5️⃣ Configurar Firewall**

**Obter seu IP:**
```bash
curl https://api.ipify.org
```

**Criar regra de firewall:**

**Opção A: Apenas seu IP (RECOMENDADO)**
```bash
gcloud compute firewall-rules create allow-nxd-8080 \
  --project=nxdata-487304 \
  --allow=tcp:8080 \
  --source-ranges=SEU.IP.AQUI/32 \
  --target-tags=nxd-server
```

**Opção B: Acesso público**
```bash
gcloud compute firewall-rules create allow-nxd-8080 \
  --project=nxdata-487304 \
  --allow=tcp:8080 \
  --source-ranges=0.0.0.0/0 \
  --target-tags=nxd-server
```

---

### **6️⃣ Fazer Upload do Código para VM**

```bash
# Criar arquivo tar
cd ~/nxd
tar -czf nxd-deploy.tar.gz --exclude=node_modules --exclude=.git --exclude=data --exclude=logs --exclude=NXD_OPS_KIT .

# Copiar para VM
gcloud compute scp nxd-deploy.tar.gz nxd-server-vm:/tmp/ \
  --zone=southamerica-east1-b \
  --project=nxdata-487304
```

---

### **7️⃣ Conectar na VM e Instalar**

```bash
gcloud compute ssh nxd-server-vm \
  --zone=southamerica-east1-b \
  --project=nxdata-487304
```

**Dentro da VM, execute:**

```bash
# Extrair código
cd /opt/nxd
sudo tar -xzf /tmp/nxd-deploy.tar.gz
rm /tmp/nxd-deploy.tar.gz

# Criar .env vazio (vamos gerar API Key depois)
echo "API_KEY=" | sudo tee .env

# Buildar e iniciar
sudo docker-compose build
sudo docker-compose up -d

# Ver logs
sudo docker-compose logs -f
```

**Aguarde ver:** `✓ Servidor rodando em http://localhost:8080`

Pressione `Ctrl+C` para sair

Digite `exit` para sair da VM

---

### **8️⃣ Obter IP da VM**

```bash
gcloud compute instances describe nxd-server-vm \
  --zone=southamerica-east1-b \
  --project=nxdata-487304 \
  --format="get(networkInterfaces[0].accessConfigs[0].natIP)"
```

**Copie o IP!**

---

### **🎉 PRONTO!**

Acesse no navegador:
```
http://SEU-IP:8080
```

---

## 📊 **COMANDOS ÚTEIS:**

### Ver logs
```bash
gcloud compute ssh nxd-server-vm \
  --zone=southamerica-east1-b \
  --project=nxdata-487304 \
  --command='sudo docker-compose logs -f'
```

### Reiniciar
```bash
gcloud compute ssh nxd-server-vm \
  --zone=southamerica-east1-b \
  --project=nxdata-487304 \
  --command='cd /opt/nxd && sudo docker-compose restart'
```

### Parar VM
```bash
gcloud compute instances stop nxd-server-vm \
  --zone=southamerica-east1-b \
  --project=nxdata-487304
```

### Iniciar VM
```bash
gcloud compute instances start nxd-server-vm \
  --zone=southamerica-east1-b \
  --project=nxdata-487304
```

---

**Tudo pelo navegador! Sem instalar nada! 🚀**
