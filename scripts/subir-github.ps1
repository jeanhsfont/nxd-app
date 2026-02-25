# Subir HubSystem (servidor + site) para um repo novo no GitHub
# 1) Autentique primeiro: gh auth login
# 2) Depois rode: .\scripts\subir-github.ps1

Set-Location $PSScriptRoot\..

# Criar repo no GitHub e adicionar remote "hub" (se ainda nao existir)
$repo = "HubSystem"
$exists = gh repo view $repo 2>$null
if (-not $exists) {
    gh repo create $repo --public --description "NXD - servidor e site completo (API Go, web-app React, deploy GCP)" --source=. --remote=hub
}
if (-not (git remote get-url hub 2>$null)) {
    git remote add hub "https://github.com/jeanhsfont/HubSystem.git"
}

# Enviar tudo
git push -u hub main

Write-Host "`nRepo: https://github.com/jeanhsfont/HubSystem" -ForegroundColor Green
