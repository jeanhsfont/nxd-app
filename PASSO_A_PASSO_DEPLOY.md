# 🚀 DEPLOY NXD - PASSO A PASSO

**⚠️ IMPORTANTE:** Preciso que você execute estes comandos porque requer autenticação no navegador.

---

## 📋 **EXECUTE ESTES COMANDOS EM ORDEM:**

### **1️⃣ Abrir PowerShell como Administrador**

Clique com botão direito no PowerShell → "Executar como Administrador"

---

### **2️⃣ Navegar até a pasta do NXD**

```powershell
cd C:\HubSystem1.0
```

---

### **3️⃣ Criar e ativar profile NXD**

```powershell
gcloud config configurations create nxd
gcloud config configurations activate nxd
```

---

### **4️⃣ Fazer login com a conta NXD**

```powershell
gcloud auth login jeanhhirata@gmail.com
```

**➡️ Uma janela do navegador vai abrir**  
**➡️ Faça login com: jeanhhirata@gmail.com**  
**➡️ Autorize o acesso**

---

### **5️⃣ Configurar projeto**

```powershell
gcloud config set project nxdata-487304
```

---

### **6️⃣ Ativar APIs necessárias**

```powershell
gcloud services enable compute.googleapis.com
gcloud services enable cloudresourcemanager.googleapis.com
gcloud services enable iam.googleapis.com
gcloud services enable logging.googleapis.com
gcloud services enable monitoring.googleapis.com
```

**⏳ Aguarde 30-60 segundos** (as APIs levam um tempo para ativar)

---

### **7️⃣ Obter seu IP público**

```powershell
$MY_IP = (Invoke-WebRequest -Uri 'https://api.ipify.org' -UseBasicParsing).Content
Write-Host "Seu IP: $MY_IP"
```

**📝 ANOTE SEU IP!**

---

### **8️⃣ Fazer o Deploy**

**Opção A: Com firewall restrito (RECOMENDADO)**
```powershell
.\NXD_OPS_KIT\ps\deploy_nxd_vm.ps1 -ProjectId "nxdata-487304" -Zone "southamerica-east1-b" -AllowedIP $MY_IP
```

**Opção B: Sem restrição de IP (NÃO recomendado)**
```powershell
.\NXD_OPS_KIT\ps\deploy_nxd_vm.ps1 -ProjectId "nxdata-487304" -Zone "southamerica-east1-b"
```

---

## ⏳ **AGUARDE O DEPLOY (2-3 minutos)**

O script vai:
1. ✅ Criar VM no Google Cloud (São Paulo)
2. ✅ Instalar Docker
3. ✅ Fazer upload do código NXD
4. ✅ Buildar containers
5. ✅ Iniciar servidor
6. ✅ Configurar firewall
7. ✅ **Te dar a URL**: `http://SEU-IP:8080`

---

## 🎯 **DEPOIS DO DEPLOY:**

### **Acessar Dashboard**
```
http://SEU-IP-DA-VM:8080
```

### **Criar Fábrica**
1. Digite: "Vale Sistemas"
2. Clique: "Criar Fábrica"
3. **COPIE A API KEY** (formato: `NXD_xxxxx...`)

### **Conectar Simulador Local**
```powershell
.\CONECTAR_SIMULADOR.bat
```

---

## 📊 **MONITORAMENTO**

### Ver logs do servidor
```powershell
gcloud compute ssh nxd-server-vm --zone=southamerica-east1-b --project=nxdata-487304 --command='sudo docker-compose logs -f'
```

### Ver status da VM
```powershell
gcloud compute instances list --project=nxdata-487304
```

---

## 🆘 **SE DER ERRO:**

### Erro: "You do not appear to have access to project"
**Solução:** Verifique se está logado com a conta correta
```powershell
gcloud auth list
gcloud config configurations activate nxd
```

### Erro: "API not enabled"
**Solução:** Aguarde 1-2 minutos e tente novamente. As APIs levam tempo para ativar.

### Erro: "Insufficient permissions"
**Solução:** Verifique se a conta tem permissões de Owner ou Editor no projeto

---

## 🎬 **COMECE AGORA!**

Abra o PowerShell como Administrador e execute os comandos acima em ordem!

**Boa sorte! 🚀**
