# ☁️ DEPLOY NXD no Microsoft Azure

**✅ $200 de crédito grátis**  
**✅ 12 meses de VM grátis**  
**✅ Datacenter no Brasil**

---

## 🎯 **PASSO A PASSO:**

### **1️⃣ Criar conta no Azure**

```
https://azure.microsoft.com/free/
```

- Clique em **"Começar gratuitamente"**
- Faça login com sua conta Microsoft
- Preencha os dados (vai pedir cartão, mas não cobra)
- **$200 de crédito grátis por 30 dias!**

---

### **2️⃣ Criar Resource Group**

**Acesse:** https://portal.azure.com/

1. No menu, clique em **"Resource groups"**
2. Clique em **"+ Create"**
3. **Subscription:** Selecione sua assinatura
4. **Resource group name:** `nxd-production`
5. **Region:** `Brazil South` (São Paulo)
6. Clique em **"Review + create"** → **"Create"**

---

### **3️⃣ Criar Virtual Machine**

1. No menu, clique em **"Virtual machines"**
2. Clique em **"+ Create"** → **"Azure virtual machine"**

**Configurações:**

#### **Basics:**
- **Subscription:** Sua assinatura
- **Resource group:** `nxd-production`
- **Virtual machine name:** `nxd-server-vm`
- **Region:** `Brazil South`
- **Image:** `Ubuntu Server 22.04 LTS`
- **Size:** `Standard_B1s` (1 vCPU, 1GB RAM - **GRÁTIS por 12 meses!**)
- **Authentication type:** `Password`
- **Username:** `azureuser`
- **Password:** (escolha uma senha forte)

#### **Disks:**
- **OS disk type:** `Standard SSD`
- **Size:** `30 GB` (incluído no free tier)

#### **Networking:**
- **Public IP:** `Yes`
- **Inbound ports:** Selecione `HTTP (80)`, `HTTPS (443)`, `SSH (22)`

Clique em **"Review + create"** → **"Create"**

⏳ **Aguarde 2-3 minutos**

---

### **4️⃣ Configurar Porta 8080 (Firewall)**

1. Vá em **"Virtual machines"** → **"nxd-server-vm"**
2. No menu lateral, clique em **"Networking"**
3. Clique em **"Add inbound port rule"**
4. **Destination port ranges:** `8080`
5. **Protocol:** `TCP`
6. **Name:** `Allow-NXD-8080`
7. **Source:** 
   - **Opção A (Seguro):** `IP Addresses` → Cole seu IP
   - **Opção B (Público):** `Any`
8. Clique em **"Add"**

---

### **5️⃣ Conectar na VM**

**Opção A: SSH pelo navegador (mais fácil)**

1. Na página da VM, clique em **"Connect"**
2. Selecione **"SSH"**
3. Clique em **"Go to Bastion"** (ou use SSH direto)

**Opção B: SSH pelo terminal**

```bash
ssh azureuser@SEU-IP-PUBLICO
# Digite a senha que você criou
```

---

### **6️⃣ Instalar Docker na VM**

**Cole estes comandos:**

```bash
# Atualizar sistema
sudo apt-get update
sudo apt-get upgrade -y

# Instalar Docker
sudo apt-get install -y docker.io docker-compose

# Habilitar Docker
sudo systemctl enable docker
sudo systemctl start docker

# Adicionar usuário ao grupo docker
sudo usermod -aG docker $USER

# Criar diretório
sudo mkdir -p /opt/nxd
sudo chown -R $USER:$USER /opt/nxd
cd /opt/nxd
```

---

### **7️⃣ Fazer Upload do Código**

**Opção A: Via SCP (do seu PC)**

```powershell
# No PowerShell do seu PC:
cd C:\HubSystem1.0

# Criar arquivo tar
tar -czf nxd-deploy.tar.gz --exclude=node_modules --exclude=.git --exclude=data --exclude=logs --exclude=NXD_OPS_KIT .

# Copiar para Azure VM
scp nxd-deploy.tar.gz azureuser@SEU-IP:/tmp/
```

**Opção B: Via Git (se tiver repositório)**

```bash
# Na VM:
cd /opt/nxd
git clone https://github.com/SEU-USUARIO/nxd.git .
```

---

### **8️⃣ Extrair e Iniciar**

**Na VM, execute:**

```bash
# Extrair código
cd /opt/nxd
tar -xzf /tmp/nxd-deploy.tar.gz
rm /tmp/nxd-deploy.tar.gz

# Criar .env
echo "API_KEY=" > .env

# Buildar e iniciar
docker-compose build
docker-compose up -d

# Ver logs
docker-compose logs -f
```

**Aguarde ver:** `✓ Servidor rodando`

Pressione `Ctrl+C` para sair

---

### **9️⃣ Obter IP Público**

**No portal Azure:**
1. Vá em **"Virtual machines"** → **"nxd-server-vm"**
2. Copie o **"Public IP address"**

---

### **🎉 PRONTO!**

Acesse no navegador:
```
http://SEU-IP-PUBLICO:8080
```

---

## 💰 **CUSTOS AZURE:**

| Recurso | Especificação | Custo/Mês |
|---------|---------------|-----------|
| VM B1s | 1 vCPU, 1GB RAM | **GRÁTIS** (12 meses) |
| Disco 30GB | Standard SSD | **GRÁTIS** (12 meses) |
| IP Público | 1 endereço | ~$3 USD |
| Tráfego | 15GB/mês | **GRÁTIS** |

**Total: $0-3/mês** (coberto pelos $200!)

---

## 🎯 **COMANDOS ÚTEIS:**

### Reiniciar VM
```bash
# No portal Azure
Virtual machines → nxd-server-vm → Restart
```

### Ver logs
```bash
ssh azureuser@SEU-IP
cd /opt/nxd
docker-compose logs -f
```

### Parar VM (economizar créditos)
```bash
# No portal Azure
Virtual machines → nxd-server-vm → Stop
```

---

**Azure é excelente! Quer seguir com ele?** 🚀