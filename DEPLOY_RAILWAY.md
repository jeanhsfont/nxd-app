# 🚂 DEPLOY NXD no Railway.app

**✅ GRÁTIS ($5/mês de crédito)**  
**✅ Deploy em 2 minutos**  
**✅ Sem cartão de crédito**

---

## 🎯 **PASSO A PASSO:**

### **1️⃣ Criar conta no Railway**

```
https://railway.app/
```

- Clique em **"Start a New Project"**
- Faça login com GitHub (ou email)

---

### **2️⃣ Criar Projeto**

1. Clique em **"+ New Project"**
2. Selecione **"Empty Project"**

---

### **3️⃣ Adicionar Serviço NXD**

1. Clique em **"+ New"**
2. Selecione **"Empty Service"**
3. Nome: **nxd-server**

---

### **4️⃣ Configurar Dockerfile**

O Railway vai detectar automaticamente o `Dockerfile.nxd`!

Mas primeiro, vamos criar um Dockerfile único na raiz:

**No seu PC, crie:** `C:\HubSystem1.0\Dockerfile`

```dockerfile
# Build stage
FROM golang:1.21-bullseye AS builder

WORKDIR /build

# Copy go files
COPY go.mod ./
COPY go.sum* ./
RUN go mod download

# Copy source
COPY core/ ./core/
COPY api/ ./api/
COPY data/ ./data/
COPY services/ ./services/
COPY web/ ./web/
COPY main.go ./

# Build
RUN CGO_ENABLED=1 go build -o nxd_server .

# Runtime stage
FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/nxd_server .
COPY --from=builder /build/web ./web

RUN mkdir -p /app/data /app/logs

# Railway usa PORT como variável de ambiente
ENV PORT=8080
EXPOSE 8080

CMD ["./nxd_server"]
```

---

### **5️⃣ Fazer Deploy**

**Opção A: Via GitHub (Recomendado)**

1. Crie um repositório no GitHub
2. Faça push do código
3. No Railway, clique em **"Deploy from GitHub"**
4. Selecione o repositório

**Opção B: Via CLI do Railway**

```bash
# Instalar Railway CLI
npm i -g @railway/cli

# Fazer login
railway login

# Fazer deploy
cd C:\HubSystem1.0
railway up
```

---

### **6️⃣ Obter URL**

Após o deploy:
1. Clique no serviço **nxd-server**
2. Vá em **"Settings"**
3. Em **"Networking"**, clique em **"Generate Domain"**

**Você vai receber uma URL tipo:**
```
https://nxd-server-production-xxxx.up.railway.app
```

---

### **🎉 PRONTO!**

Acesse a URL e teste o NXD!

---

## 💰 **CUSTOS:**

- **$5 grátis/mês** (suficiente para ~500 horas de uso)
- Depois: ~$5-10/mês

---

**Muito mais simples que Google Cloud! 🚀**
