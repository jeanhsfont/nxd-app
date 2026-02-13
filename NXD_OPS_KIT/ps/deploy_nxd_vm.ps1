# NXD OPS KIT — Deploy NXD em VM do Google Compute Engine
# Uso: .\deploy_nxd_vm.ps1 [-ProjectId xxx] [-Zone southamerica-east1-b] [-SkipTests]
param(
    [string]$ProjectId,
    [string]$Zone,
    [string]$MachineType = "e2-micro",
    [string]$DiskSize = "10GB",
    [switch]$SkipTests,
    [string]$AllowedIP
)

. "$PSScriptRoot\_lib.ps1" -ErrorAction Stop

$cfg = Get-OpsConfig
$proj = if ($ProjectId) { $ProjectId } else { $cfg["PROJECT_ID"] }
if (-not $proj) { $proj = Get-GcloudProject }
$zone = if ($Zone) { $Zone } else { $cfg["VM_ZONE"] }
if (-not $zone) { $zone = "southamerica-east1-b" }

if (-not $proj) {
    Write-OpsLog "PROJECT_ID não definido. Defina em config/nxd.env, -ProjectId ou gcloud config set project" "ERROR"
    exit 1
}

# Repo root (um nível acima de NXD_OPS_KIT)
$kitRoot = Get-OpsKitRoot
$repoRoot = Split-Path $kitRoot -Parent

Write-OpsLog "=== Deploy NXD em VM ==="
Write-OpsLog "Project: $proj | Zone: $zone | Machine: $MachineType"

# Verifica se gcloud está instalado e autenticado
if (-not (Test-GcloudInstalled)) {
    Write-OpsLog "gcloud CLI não encontrado. Instale: https://cloud.google.com/sdk/docs/install" "ERROR"
    exit 1
}

if (-not (Test-GcloudAuth)) {
    Write-OpsLog "gcloud não autenticado. Execute: gcloud auth login" "ERROR"
    exit 1
}

# Nome da VM
$vmName = "nxd-server-vm"

# Verifica se a VM já existe
Write-OpsLog "Verificando se VM $vmName já existe..."
$vmExists = gcloud compute instances describe $vmName --zone=$zone --project=$proj 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-OpsLog "VM $vmName já existe. Atualizando código..." "INFO"
    
    # Copia código para a VM
    Write-OpsLog "Copiando código para VM..."
    Push-Location $repoRoot
    try {
        # Cria arquivo tar com o código (exclui node_modules, .git, etc)
        $tarFile = "nxd-deploy.tar.gz"
        tar -czf $tarFile --exclude=node_modules --exclude=.git --exclude=data --exclude=logs --exclude=NXD_OPS_KIT .
        
        # Copia para VM
        gcloud compute scp $tarFile ${vmName}:/tmp/ --zone=$zone --project=$proj
        
        # Extrai e reinicia na VM
        $setupScript = @"
cd /opt/nxd
sudo tar -xzf /tmp/nxd-deploy.tar.gz
sudo docker-compose down
sudo docker-compose build
sudo docker-compose up -d
rm /tmp/nxd-deploy.tar.gz
"@
        
        $setupScript | gcloud compute ssh $vmName --zone=$zone --project=$proj --command="bash -s"
        
        Remove-Item $tarFile -Force
        
        Write-OpsLog "Código atualizado e containers reiniciados!" "INFO"
    } finally {
        Pop-Location
    }
} else {
    Write-OpsLog "VM $vmName não existe. Criando nova VM..." "INFO"
    
    # Script de inicialização da VM
    $startupScript = @"
#!/bin/bash
set -e

# Atualiza sistema
apt-get update
apt-get install -y docker.io docker-compose git

# Habilita Docker
systemctl enable docker
systemctl start docker

# Cria diretório para NXD
mkdir -p /opt/nxd
cd /opt/nxd

# Aguarda código ser copiado (será feito via gcloud compute scp)
echo "VM pronta para receber código NXD"
"@
    
    $startupScriptFile = Join-Path $env:TEMP "nxd-startup.sh"
    $startupScript | Out-File -FilePath $startupScriptFile -Encoding UTF8
    
    # Cria a VM
    Write-OpsLog "Criando VM $vmName..."
    gcloud compute instances create $vmName `
        --project=$proj `
        --zone=$zone `
        --machine-type=$MachineType `
        --boot-disk-size=$DiskSize `
        --boot-disk-type=pd-standard `
        --image-family=ubuntu-2204-lts `
        --image-project=ubuntu-os-cloud `
        --metadata-from-file=startup-script=$startupScriptFile `
        --tags=nxd-server,http-server `
        --scopes=https://www.googleapis.com/auth/cloud-platform
    
    if ($LASTEXITCODE -ne 0) {
        Write-OpsLog "Falha ao criar VM" "ERROR"
        exit 1
    }
    
    Remove-Item $startupScriptFile -Force
    
    Write-OpsLog "Aguardando VM inicializar (60 segundos)..."
    Start-Sleep -Seconds 60
    
    # Copia código para a VM
    Write-OpsLog "Copiando código para VM..."
    Push-Location $repoRoot
    try {
        $tarFile = "nxd-deploy.tar.gz"
        tar -czf $tarFile --exclude=node_modules --exclude=.git --exclude=data --exclude=logs --exclude=NXD_OPS_KIT .
        
        gcloud compute scp $tarFile ${vmName}:/tmp/ --zone=$zone --project=$proj
        
        $setupScript = @"
sudo mkdir -p /opt/nxd
cd /opt/nxd
sudo tar -xzf /tmp/nxd-deploy.tar.gz
sudo docker-compose build
sudo docker-compose up -d
rm /tmp/nxd-deploy.tar.gz
"@
        
        $setupScript | gcloud compute ssh $vmName --zone=$zone --project=$proj --command="bash -s"
        
        Remove-Item $tarFile -Force
    } finally {
        Pop-Location
    }
}

# Configura regra de firewall
Write-OpsLog "Configurando firewall..."
$firewallRule = "allow-nxd-8080"
$firewallExists = gcloud compute firewall-rules describe $firewallRule --project=$proj 2>&1

if ($LASTEXITCODE -ne 0) {
    if ($AllowedIP) {
        Write-OpsLog "Criando regra de firewall (apenas IP $AllowedIP)..."
        gcloud compute firewall-rules create $firewallRule `
            --project=$proj `
            --allow=tcp:8080 `
            --source-ranges=$AllowedIP `
            --target-tags=nxd-server `
            --description="Permite acesso ao NXD apenas do IP autorizado"
    } else {
        Write-OpsLog "Criando regra de firewall (acesso público - NÃO RECOMENDADO)..." "WARN"
        gcloud compute firewall-rules create $firewallRule `
            --project=$proj `
            --allow=tcp:8080 `
            --source-ranges=0.0.0.0/0 `
            --target-tags=nxd-server `
            --description="Permite acesso ao NXD (público)"
    }
}

# Obtém IP externo da VM
Write-OpsLog "Obtendo IP externo da VM..."
$vmIP = gcloud compute instances describe $vmName --zone=$zone --project=$proj --format="get(networkInterfaces[0].accessConfigs[0].natIP)"

if ($vmIP) {
    Write-OpsLog "=== DEPLOY CONCLUÍDO ===" "INFO"
    Write-Host ""
    Write-Host "✅ NXD está rodando em: http://${vmIP}:8080" -ForegroundColor Green
    Write-Host ""
    Write-Host "📋 Próximos passos:" -ForegroundColor Cyan
    Write-Host "1. Acesse http://${vmIP}:8080 no navegador"
    Write-Host "2. Crie uma fábrica e copie a API Key"
    Write-Host "3. Configure o DX com:"
    Write-Host "   - Endpoint: http://${vmIP}:8080/api/ingest"
    Write-Host "   - API Key: (cole a chave gerada)"
    Write-Host ""
    Write-Host "🔒 Segurança:" -ForegroundColor Yellow
    if ($AllowedIP) {
        Write-Host "   ✓ Firewall configurado para IP: $AllowedIP"
    } else {
        Write-Host "   ⚠️  ATENÇÃO: Servidor acessível publicamente!"
        Write-Host "   Recomendado: Execute novamente com -AllowedIP seu.ip.aqui"
    }
    Write-Host ""
    Write-Host "📊 Monitoramento:" -ForegroundColor Cyan
    Write-Host "   - Logs: gcloud compute ssh $vmName --zone=$zone --project=$proj --command='sudo docker-compose logs -f'"
    Write-Host "   - Status: gcloud compute ssh $vmName --zone=$zone --project=$proj --command='sudo docker-compose ps'"
    Write-Host ""
} else {
    Write-OpsLog "Não foi possível obter IP da VM" "ERROR"
    exit 1
}

exit 0
