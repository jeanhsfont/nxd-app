#!/bin/sh

echo "🔧 Aguardando servidor HUB ficar pronto..."

# Wait for server to be ready
until wget -q --spider http://hub-server:8080/api/health 2>/dev/null; do
    echo "  Tentando conectar..."
    sleep 2
done

echo "✓ Servidor HUB pronto!"
echo "🚀 Iniciando simulador DX..."

if [ -z "$API_KEY" ]; then
    echo "❌ Erro: API_KEY não definida"
    echo "Execute: docker-compose up com a variável API_KEY"
    exit 1
fi

echo "✓ API Key: ${API_KEY:0:20}..."
echo "📡 Simulando CLPs Siemens e Delta..."

exec ./dx_simulator "$API_KEY"
