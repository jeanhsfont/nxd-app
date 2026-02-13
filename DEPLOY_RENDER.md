# 🎨 DEPLOY NXD no Render.com

**✅ GRÁTIS (750 horas/mês)**  
**✅ Deploy automático**  
**✅ Sem cartão de crédito**

---

## 🎯 **PASSO A PASSO:**

### **1️⃣ Criar conta no Render**

```
https://render.com/
```

- Clique em **"Get Started"**
- Faça login com GitHub

---

### **2️⃣ Fazer Push para GitHub**

**No seu PC:**

```bash
cd C:\HubSystem1.0

# Inicializar Git (se ainda não tem)
git init
git add .
git commit -m "Initial NXD commit"

# Criar repositório no GitHub e fazer push
git remote add origin https://github.com/SEU-USUARIO/nxd.git
git push -u origin main
```

---

### **3️⃣ Criar Web Service no Render**

1. No Render, clique em **"New +"**
2. Selecione **"Web Service"**
3. Conecte seu repositório GitHub
4. Selecione o repositório **nxd**

---

### **4️⃣ Configurar Serviço**

**Nome:**
```
nxd-server
```

**Environment:**
```
Docker
```

**Dockerfile Path:**
```
Dockerfile.nxd
```

**Plan:**
```
Free
```

Clique em **"Create Web Service"**

---

### **5️⃣ Aguardar Deploy**

⏳ O Render vai:
1. Detectar o Dockerfile
2. Buildar a imagem
3. Fazer deploy
4. Te dar uma URL

**Tempo:** 3-5 minutos

---

### **6️⃣ Obter URL**

Após o deploy, você vai receber uma URL tipo:
```
https://nxd-server.onrender.com
```

---

### **🎉 PRONTO!**

Acesse a URL e teste o NXD!

---

## ⚠️ **LIMITAÇÕES DO PLANO GRÁTIS:**

- ⏸️ **Sleep após 15 min de inatividade** (acorda em 30s quando acessar)
- 📊 **750 horas/mês** (suficiente para testes)
- 💾 **Dados não persistem** (SQLite é efêmero)

**Para produção, upgrade para $7/mês** (dados persistentes)

---

## 💡 **ALTERNATIVA: Persistir Dados**

Use **Render Disk** (adicional $1/mês):
1. No serviço, vá em **"Disks"**
2. Clique em **"Add Disk"**
3. Mount Path: `/app/data`
4. Size: 1GB

---

**Render é perfeito para testes! 🚀**
