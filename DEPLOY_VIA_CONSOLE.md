# 🌐 DEPLOY NXD via Google Cloud Console (Navegador)

**✅ SEM precisar instalar gcloud!**  
**✅ TUDO pelo navegador!**

---

## 🎯 **PASSO A PASSO:**

### **1️⃣ Acessar Google Cloud Console**

```
https://console.cloud.google.com/
```

- Faça login com: **jeanhhirata@gmail.com**
- Selecione o projeto: **nxdata-487304**

---

### **2️⃣ Ativar APIs Necessárias**

**Acesse:** https://console.cloud.google.com/apis/library

Ative estas APIs (clique em cada uma e depois "ATIVAR"):
1. **Compute Engine API**
2. **Cloud Resource Manager API**
3. **IAM API**

⏳ **Aguarde 1-2 minutos** após ativar

---

### **3️⃣ Criar VM**

**Acesse:** https://console.cloud.google.com/compute/instances

Clique em **"CRIAR INSTÂNCIA"**

**Configurações:**

#### **Nome:**
```
nxd-server-vm
```

#### **Região:**
```
southamerica-east1 (São Paulo)
```

#### **Zona:**
```
southamerica-east1-b
```

#### **Tipo de máquina:**
```
Série E2 → e2-micro (2 vCPUs compartilhadas, 1 GB de memória)
```
✅ **GRÁTIS no free tier!**

#### **Disco de inicialização:**
- Clique em **"ALTERAR"**
- Sistema operacional: **Ubuntu**
- Versão: **Ubuntu 22.04 LTS**
- Tipo de disco: **Disco permanente padrão**
- Tamanho: **10 GB**
- Clique em **"SELECIONAR"**

#### **Firewall:**
- ✅ Marque: **"Permitir tráfego HTTP"**
- ✅ Marque: **"Permitir tráfego HTTPS"**

#### **Avançado → Rede:**
- Expanda **"Rede"**
- Em **"Tags de rede"**, adicione: `nxd-server`

Clique em **"CRIAR"**

⏳ **Aguarde 1-2 minutos** para a VM ser criada

---

### **4️⃣ Configurar Firewall para Porta 8080**

**Acesse:** https://console.cloud.google.com/networking/firewalls/list

Clique em **"CRIAR REGRA DE FIREWALL"**

**Configurações:**

#### **Nome:**
```
allow-nxd-8080
```

#### **Destinos:**
```
Tags de destino especificadas
```

#### **Tags de destino:**
```
nxd-server
```

#### **Filtro de origem:**
```
Intervalos de IPv4
```

#### **Intervalos de IPv4 de origem:**

**Opção A: Apenas seu IP (RECOMENDADO)**
```
SEU.IP.AQUI/32
```
(Descubra seu IP em: https://www.whatismyip.com/)

**Opção B: Acesso público (NÃO recomendado)**
```
0.0.0.0/0
```

#### **Protocolos e portas:**
- ✅ Marque: **"Protocolos e portas especificados"**
- Selecione: **TCP**
- Portas: `8080`

Clique em **"CRIAR"**

---

### **5️⃣ Conectar na VM e Instalar NXD**

**Acesse:** https://console.cloud.google.com/compute/instances

Encontre a VM **"nxd-server-vm"**

Clique no botão **"SSH"** (vai abrir um terminal no navegador)

---

### **6️⃣ Executar Comandos na VM**

**Cole estes comandos no terminal SSH:**

```bash
# 1. Atualizar sistema
sudo apt-get update
sudo apt-get install -y docker.io docker-compose git

# 2. Habilitar Docker
sudo systemctl enable docker
sudo systemctl start docker

# 3. Criar diretório
sudo mkdir -p /opt/nxd
cd /opt/nxd

# 4. Fazer upload do código (AGUARDE - vou te dar o comando)
```

**⏸️ PARE AQUI!** Vou criar um script para você fazer upload do código!

---

### **7️⃣ Fazer Upload do Código**

**No seu PC (PowerShell):**

```powershell
# Navegar até a pasta do NXD
cd C:\HubSystem1.0

# Criar arquivo tar
tar -czf nxd-deploy.tar.gz --exclude=node_modules --exclude=.git --exclude=data --exclude=logs --exclude=NXD_OPS_KIT .

# Fazer upload via console
# (Vou criar um script para isso)
```

**Depois, volte ao terminal SSH e execute:**

```bash
# Extrair código
cd /opt/nxd
sudo tar -xzf nxd-deploy.tar.gz

# Criar arquivo .env (IMPORTANTE!)
sudo nano .env
```

**Cole isso no .env:**
```
API_KEY=
```
(Deixe vazio por enquanto, vamos gerar depois)

**Salve:** `Ctrl+X` → `Y` → `Enter`

---

### **8️⃣ Iniciar NXD**

```bash
# Buildar e iniciar
sudo docker-compose build
sudo docker-compose up -d

# Ver logs
sudo docker-compose logs -f
```

**Aguarde ver:** `✓ Servidor rodando em http://localhost:8080`

Pressione `Ctrl+C` para sair dos logs

---

### **9️⃣ Obter IP da VM**

**Volte para:** https://console.cloud.google.com/compute/instances

Copie o **"IP externo"** da VM **nxd-server-vm**

---

### **🎉 PRONTO!**

Acesse no navegador:
```
http://SEU-IP-EXTERNO:8080
```

---

## 📊 **PRÓXIMOS PASSOS:**

1. ✅ Acesse o dashboard
2. ✅ Crie uma fábrica
3. ✅ Copie a API Key
4. ✅ Configure o simulador local

---

## 🆘 **TROUBLESHOOTING:**

### Não consigo acessar o dashboard
1. Verifique se a regra de firewall foi criada
2. Verifique se usou o IP externo correto
3. Aguarde 1-2 minutos após iniciar os containers

### Erro ao fazer upload do código
Use o **Cloud Shell** (terminal no próprio navegador):
1. Clique no ícone `>_` no topo do console
2. Execute: `git clone SEU-REPOSITORIO` (se tiver)
3. Ou faça upload manual dos arquivos

---

**Muito mais fácil que gcloud! 🎉**
