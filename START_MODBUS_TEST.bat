@echo off
title NXD - Teste Modbus Real
color 0A

echo.
echo  ================================================================
echo          NXD - AMBIENTE DE TESTE MODBUS REAL
echo  ================================================================
echo.
echo   Este script inicia 3 terminais:
echo.
echo   1. CLP SIEMENS (Modbus TCP porta 502)
echo   2. CLP DELTA   (Modbus TCP porta 503)  
echo   3. DX GATEWAY  (Le Modbus, envia para NXD)
echo.
echo   ARQUITETURA:
echo.
echo   [CLP Siemens:502] ---+
echo                        +---[DX Gateway]---HTTP---[NXD Cloud]
echo   [CLP Delta:503]   ---+
echo.
echo  ================================================================
echo.

cd /d "%~dp0modbus-simulator"

echo  [1/3] Iniciando CLP Siemens (porta 502)...
start "CLP SIEMENS S7-1200" cmd /k "node clp-siemens.js"

timeout /t 2 /nobreak >nul

echo  [2/3] Iniciando CLP Delta (porta 503)...
start "CLP DELTA DVP-28SV" cmd /k "node clp-delta.js"

timeout /t 2 /nobreak >nul

echo  [3/3] Iniciando DX Gateway...
start "DX-2100 GATEWAY" cmd /k "node dx-gateway.js"

echo.
echo  ================================================================
echo   TODOS OS COMPONENTES INICIADOS!
echo  ================================================================
echo.
echo   Agora no terminal do DX Gateway:
echo.
echo   1. Digite: api SUA_API_KEY_AQUI
echo   2. Digite: start
echo.
echo   O gateway vai ler os CLPs via Modbus e enviar para o NXD!
echo.
echo   Comandos nos CLPs:
echo   - P = Parar maquina
echo   - R = Retomar producao
echo   - S = Superaquecer (Siemens)
echo   - C = Falha comunicacao (Delta)
echo.
echo   Pressione qualquer tecla para fechar esta janela...
pause >nul
