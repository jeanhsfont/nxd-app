@echo off
title Delta DX-2100 Simulator
echo.
echo  ====================================================
echo   DELTA DX-2100 - SIMULADOR DE GATEWAY INDUSTRIAL
echo  ====================================================
echo.
echo   Este simulador replica o comportamento do modulo
echo   DX Delta para testes do sistema NXD.
echo.
echo   Abrindo interface no navegador...
echo.
start "" "%~dp0dx-simulator\index.html"
echo.
echo   Interface aberta! 
echo.
echo   INSTRUCOES:
echo   1. Acesse o NXD e crie uma fabrica
echo   2. Copie a API Key gerada
echo   3. Cole no campo "API Key" do simulador
echo   4. Clique em "Conectar"
echo   5. Veja os dados fluindo em tempo real!
echo.
echo   Pressione qualquer tecla para fechar...
pause >nul
