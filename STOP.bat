@echo off
chcp 65001 >nul
title HUB System - Parar
color 0C

echo.
echo 🛑 Parando HUB System...
echo.

docker-compose down

echo.
echo ✓ Sistema parado
echo.
pause
