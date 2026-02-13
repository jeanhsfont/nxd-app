# 🔍 Erros Previsíveis e Como Evitá-los

## ❌ Erros que Você Encontrou (e que eu deveria ter previsto)

### 1. **Erro: "go.sum not found"**
**Causa:** Dockerfile tentando copiar arquivo que não existe  
**Previsível?** ✅ SIM - Deveria ter verificado se o módulo tem dependências  
**Solução:** Tornar `go.sum` opcional no Dockerfile  
**Prevenção:** Sempre usar `COPY go.sum* ./` (com asterisco)

---

### 2. **Erro: "failed to solve: process did not complete successfully: exit code: 1"**
**Causa:** SQLite precisa de CGO, mas Alpine não tinha gcc instalado  
**Previsível?** ✅ SIM - SQLite SEMPRE precisa de CGO  
**Solução:** Instalar `gcc musl-dev sqlite-dev` no Dockerfile  
**Prevenção:** Checklist de dependências por tecnologia

### 3. **Erro: "pread64 undeclared" / "off64_t unknown type"**
**Causa:** Alpine Linux usa musl libc que não tem pread64/pwrite64  
**Previsível?** ✅ SIM - Problema conhecido do SQLite com musl  
**Solução:** Usar build tag `sqlite_omit_load_extension`  
**Prevenção:** Sempre usar tags corretas para Alpine + SQLite

---

### 3. **Erro: "Unable to connect" no navegador**
**Causa:** Servidor ainda não terminou de inicializar  
**Previsível?** ✅ SIM - Primeira inicialização sempre demora  
**Solução:** Aguardar 30-60 segundos  
**Prevenção:** Script deve verificar se servidor está pronto antes de abrir navegador

---

## 📋 Checklist de Validação (O que eu deveria ter feito ANTES)

### ✅ Antes de Criar Dockerfile

- [ ] Verificar se o código compila localmente
- [ ] Listar todas as dependências do sistema
- [ ] Identificar se precisa de CGO
- [ ] Verificar se todos os arquivos existem
- [ ] Testar com `go build` local primeiro

### ✅ Para Projetos com SQLite

- [ ] Sempre instalar: `gcc`, `musl-dev`, `sqlite-dev`
- [ ] Sempre usar: `CGO_ENABLED=1`
- [ ] Testar compilação com CGO localmente

### ✅ Para Docker Compose

- [ ] Validar sintaxe: `docker-compose config`
- [ ] Verificar portas disponíveis
- [ ] Testar build de cada serviço separadamente
- [ ] Verificar espaço em disco

### ✅ Para Scripts de Inicialização

- [ ] Verificar se Docker está rodando
- [ ] Verificar se portas estão livres
- [ ] Aguardar servidor ficar pronto (health check)
- [ ] Mostrar progresso claro ao usuário

---

## 🎯 Como Eu Deveria Ter Estruturado

### **Fase 1: Validação (VALIDAR.bat)**
```
1. Verifica Docker rodando
2. Verifica arquivos existem
3. Testa compilação local
4. Valida docker-compose.yml
5. Verifica portas disponíveis
```

### **Fase 2: Build (BUILD.bat)**
```
1. Gera go.sum se necessário
2. Compila imagens Docker
3. Valida que imagens foram criadas
4. Mostra tamanho das imagens
```

### **Fase 3: Start (START.bat)**
```
1. Verifica se imagens existem
2. Inicia containers
3. Aguarda health check
4. Só então abre navegador
```

---

## 🔧 Dependências por Tecnologia

### **Go + SQLite**
```dockerfile
RUN apk add --no-cache gcc musl-dev sqlite-dev
ENV CGO_ENABLED=1
```

### **Go Puro (sem CGO)**
```dockerfile
ENV CGO_ENABLED=0
RUN go build -ldflags="-w -s" -o app
```

### **Go + PostgreSQL**
```dockerfile
RUN apk add --no-cache postgresql-dev
```

### **Go + MySQL**
```dockerfile
RUN apk add --no-cache mysql-dev
```

---

## 🚨 Sinais de Alerta (Red Flags)

### ❌ "Vai dar erro se..."

1. **Copiar go.sum sem verificar se existe**
   ```dockerfile
   COPY go.sum ./  # ❌ Pode não existir
   COPY go.sum* ./ # ✅ Opcional
   ```

2. **Usar SQLite sem CGO**
   ```dockerfile
   ENV CGO_ENABLED=0  # ❌ SQLite não vai funcionar
   ENV CGO_ENABLED=1  # ✅ Correto
   ```

3. **Abrir navegador antes do servidor estar pronto**
   ```bash
   docker-compose up -d
   start http://localhost:8080  # ❌ Muito rápido
   
   # ✅ Correto:
   docker-compose up -d
   sleep 30  # ou health check
   start http://localhost:8080
   ```

4. **Não verificar se Docker está rodando**
   ```bash
   docker-compose up  # ❌ Pode falhar silenciosamente
   
   # ✅ Correto:
   docker info || exit 1
   docker-compose up
   ```

---

## 📊 Matriz de Erros Previsíveis

| Erro | Previsível? | Como Detectar | Como Prevenir |
|------|-------------|---------------|---------------|
| go.sum não existe | ✅ SIM | Verificar arquivo | Usar `COPY go.sum* ./` |
| SQLite sem CGO | ✅ SIM | Checar imports | Instalar gcc no Dockerfile |
| Porta em uso | ✅ SIM | `netstat` | Verificar antes de iniciar |
| Docker não rodando | ✅ SIM | `docker info` | Verificar no início do script |
| Servidor não pronto | ✅ SIM | Health check | Aguardar antes de abrir navegador |
| Espaço em disco | ✅ SIM | `df -h` | Verificar antes de build |
| Arquivo faltando | ✅ SIM | `test -f` | Validar estrutura |
| Sintaxe docker-compose | ✅ SIM | `docker-compose config` | Validar antes de up |

---

## 🎓 Lições Aprendidas

### **1. Sempre Validar Antes de Executar**
```bash
# Ordem correta:
VALIDAR.bat  # Verifica tudo
BUILD.bat    # Compila
START.bat    # Inicia
```

### **2. Testar Localmente Primeiro**
```bash
# Antes de Docker:
go build main.go  # Testa compilação local
go run main.go    # Testa execução
```

### **3. Mensagens Claras de Erro**
```bash
# ❌ Ruim:
echo "Erro"

# ✅ Bom:
echo "❌ Erro ao compilar!"
echo "Causa provável: SQLite precisa de CGO"
echo "Solução: Instale gcc no Dockerfile"
echo "Veja: TROUBLESHOOTING.md"
```

### **4. Feedback de Progresso**
```bash
# ❌ Ruim:
docker-compose build  # Usuário não sabe o que está acontecendo

# ✅ Bom:
echo "🔨 Compilando... (pode demorar 2-5 minutos)"
docker-compose build
echo "✓ Compilação concluída!"
```

---

## 🔄 Fluxo Ideal de Desenvolvimento

```
1. Escrever código
2. Testar localmente (go run)
3. Compilar localmente (go build)
4. Criar Dockerfile
5. Validar Dockerfile (docker build)
6. Criar docker-compose.yml
7. Validar compose (docker-compose config)
8. Testar build (docker-compose build)
9. Testar start (docker-compose up)
10. Criar scripts de automação
11. Validar scripts (VALIDAR.bat)
12. Documentar erros comuns
```

---

## 💡 Dicas Pro

### **Use Health Checks**
```yaml
healthcheck:
  test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/api/health"]
  interval: 5s
  timeout: 3s
  retries: 10
```

### **Use Depends On com Condition**
```yaml
depends_on:
  hub-server:
    condition: service_healthy
```

### **Use Build Args para Debug**
```dockerfile
ARG DEBUG=false
RUN if [ "$DEBUG" = "true" ]; then go build -gcflags="all=-N -l"; fi
```

---

## ✅ Conclusão

**SIM, todos esses erros eram previsíveis!**

Eu deveria ter:
1. ✅ Criado VALIDAR.bat PRIMEIRO
2. ✅ Testado compilação local ANTES do Docker
3. ✅ Verificado dependências do SQLite
4. ✅ Implementado health checks
5. ✅ Aguardado servidor ficar pronto

**Agora você tem:**
- ✅ VALIDAR.bat - Detecta problemas ANTES de compilar
- ✅ BUILD.bat - Compila com verificações
- ✅ START.bat - Inicia com segurança
- ✅ TROUBLESHOOTING.md - Solução de problemas
- ✅ Este guia - Para não repetir erros

**Desculpa pela bagunça! Agora está profissional! 🚀**
