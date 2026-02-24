# Etapa 1: Build
# Usamos uma imagem oficial do Go para compilar nossa aplicação.
# Especificamos a versão para garantir builds consistentes.
FROM golang:1.24-alpine AS builder

# Define o diretório de trabalho dentro do contêiner.
WORKDIR /app

# Copia os arquivos de gerenciamento de dependências.
COPY go.mod go.sum ./

# Baixa as dependências.
RUN go mod download

# Copia todo o código fonte para o contêiner.
COPY . .

# Compila a aplicação para um binário estático para Linux.
# CGO_ENABLED=0 desabilita o CGO, necessário para compilações estáticas.
# GOOS=linux especifica o sistema operacional de destino.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# ---

# Etapa 2: Imagem Final
# Usamos uma imagem 'alpine' que é mínima mas inclui ferramentas de depuração.
FROM alpine:latest

# Define o diretório de trabalho.
WORKDIR /app

# Copia o executável compilado da etapa de build.
COPY --from=builder /app/server .

# Copia o frontend compilado (React/Vite dist).
COPY --from=builder /app/dist ./dist

# Copia os certificados CA da imagem de build para permitir comunicação HTTPS.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expõe a porta que a aplicação usará (será definida pela variável de ambiente PORT).
EXPOSE 8081

# Comando para iniciar o servidor quando o contêiner for executado.
CMD ["/app/server"]
