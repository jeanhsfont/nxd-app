# 🆘 Troubleshooting - HUB System

## ❌ Problema: "Unable to connect" no navegador

### Causa
O servidor ainda não terminou de inicializar.

### Solução
```bash
# 1. Aguarde mais tempo (primeira vez pode demorar 60 segundos)

# 2. Verifique se o container está rodando:
docker-compose ps

# 3. Veja os logs do servidor:
docker-compose logs hub-server

# 4. Se aparecer "Servidor rodando em http://localhost:8080", aguarde mais 10 segundos
```

---

## ❌ Problema: "Docker não está rodando"

### Solução
1. Abra o Docker Desktop
2. Aguarde o ícone ficar verde
3. Execute `START.bat` novamente

---

## ❌ Problema: Compilação demora muito

### Causa
Na primeira vez, o Docker precisa baixar imagens base e compilar o código.

### Solução
**É normal!** Pode demorar 2-5 minutos dependendo da sua internet.

Aguarde até ver:
```
✓ Imagens compiladas!
```

---

## ❌ Problema: "Porta 8080 em uso"

### Solução
```bash
# Pare tudo:
docker-compose down

# Ou mude a porta em docker-compose.yml:
ports:
  - "8081:8080"  # Mude 8080 para 8081
```

---

## ❌ Problema: Máquinas não aparecem no dashboard

### Possíveis Causas

#### 1. Simulador não está rodando
```bash
# Verifique:
docker-compose ps

# Deve mostrar:
# hubsystem-simulator   running
```

#### 2. API Key incorreta no .env
```bash
# Verifique o arquivo .env:
API_KEY=HUB_abc123...  # Deve ter 68 caracteres
```

#### 3. Simulador não conseguiu conectar
```bash
# Veja os logs:
docker-compose logs dx-simulator

# Deve mostrar:
# ✓ Servidor HUB pronto!
# 🚀 Iniciando simulador DX...
```

---

## ❌ Problema: Erro ao compilar imagens

### Solução 1: Limpar cache do Docker
```bash
docker-compose down
docker system prune -a
docker-compose build --no-cache
```

### Solução 2: Verificar arquivos
```bash
# Certifique-se que estes arquivos existem:
# - Dockerfile.hub
# - Dockerfile.simulator
# - docker-compose.yml
# - go.mod
# - main.go
```

---

## ❌ Problema: "Error response from daemon"

### Solução
```bash
# Reinicie o Docker Desktop
# Depois:
docker-compose down
docker-compose up -d
```

---

## 🔍 Comandos de Diagnóstico

### Ver status dos containers
```bash
docker-compose ps
```

### Ver logs em tempo real
```bash
# Todos os logs:
docker-compose logs -f

# Só servidor:
docker-compose logs -f hub-server

# Só simulador:
docker-compose logs -f dx-simulator
```

### Ver últimas 50 linhas de log
```bash
docker-compose logs --tail=50 hub-server
```

### Verificar se servidor está respondendo
Abra no navegador:
```
http://localhost:8080/api/health
```

Deve retornar:
```json
{
  "status": "online",
  "time": "2026-02-13T..."
}
```

### Entrar dentro do container
```bash
# Servidor:
docker exec -it hubsystem-server sh

# Simulador:
docker exec -it hubsystem-simulator sh
```

---

## 🔄 Resetar Tudo

Se nada funcionar, reset completo:

```bash
# 1. Para tudo
docker-compose down

# 2. Remove containers e volumes
docker-compose down -v

# 3. Remove imagens
docker rmi hubsystem1.0-hub-server hubsystem1.0-dx-simulator

# 4. Limpa cache
docker system prune -a

# 5. Reconstrói tudo
docker-compose build --no-cache

# 6. Inicia
docker-compose up -d
```

---

## 📞 Ainda com Problemas?

1. Copie a saída de:
   ```bash
   docker-compose logs
   ```

2. Tire screenshot do erro

3. Volte aqui com as informações!

---

## ✅ Checklist de Verificação

Antes de reportar erro, verifique:

- [ ] Docker Desktop está rodando
- [ ] Executou `START.bat` e aguardou completar
- [ ] Arquivo `.env` existe com API Key válida
- [ ] `docker-compose ps` mostra containers rodando
- [ ] Aguardou pelo menos 60 segundos após iniciar
- [ ] Tentou acessar http://localhost:8080/api/health
- [ ] Verificou os logs com `docker-compose logs`
