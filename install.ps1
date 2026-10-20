# AxiomLang (axlc) Automated Universal Installer for Windows (PowerShell)
# Run in Administrator PowerShell: powershell -ExecutionPolicy Bypass -File install.ps1

Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  AxiomLang Compiler & Toolchain (axlc) Installer for Windows" -ForegroundColor Cyan
Write-Host "================================================================================" -ForegroundColor Cyan

$InstallDir = "$env:ProgramFiles\AxiomLang\bin"
$SourceDir = $PSScriptRoot

# 1. Check Go compiler
if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
    Write-Host " [!] Go compiler not found in PATH. Please install Go from https://go.dev/dl/ or via: winget install GoLang.Go" -ForegroundColor Yellow
    exit 1
}

# 2. Build Windows executable
Write-Host " [*] Compiling axlc.exe binary..." -ForegroundColor Green
Set-Location -Path $SourceDir
go build -ldflags="-s -w" -o axlc.exe cmd\axlc\main.go

if (-not (Test-Path "$SourceDir\axlc.exe")) {
    Write-Host " [FATAL] Compilation failed." -ForegroundColor Red
    exit 1
}

# 3. Create ProgramFiles directory & copy binary
Write-Host " [*] Installing binary to $InstallDir..." -ForegroundColor Green
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

Copy-Item "$SourceDir\axlc.exe" -Destination "$InstallDir\axlc.exe" -Force

# 4. Add to System PATH if not already present
$MachinePath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::Machine)
if ($MachinePath -notlike "*$InstallDir*") {
    Write-Host " [*] Adding $InstallDir to System PATH..." -ForegroundColor Green
    [Environment]::SetEnvironmentVariable("Path", "$MachinePath;$InstallDir", [EnvironmentVariableTarget]::Machine)
    $env:Path += ";$InstallDir"
}

# 5. Register .axl and .axlui File Associations in Windows Registry
Write-Host " [*] Registering .axl and .axlui file associations..." -ForegroundColor Green
New-Item -Path "HKCR:\.axl" -Value "AxiomLang.Source" -Force | Out-Null
New-Item -Path "HKCR:\.axlui" -Value "AxiomLang.UI" -Force | Out-Null
New-Item -Path "HKCR:\AxiomLang.Source\shell\open\command" -Value "`"$InstallDir\axlc.exe`" -run `"%1`"" -Force | Out-Null

Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  [SUCCESS] AxiomLang (axlc.exe) successfully installed on Windows!" -ForegroundColor Green
Write-Host "  Try running in CMD / PowerShell:" -ForegroundColor White
Write-Host "    axlc -run examples\01_hello_world.axl" -ForegroundColor Yellow
Write-Host "    axlc -o my_app.exe examples\09_axlui_gui_studio.axl" -ForegroundColor Yellow
Write-Host "================================================================================" -ForegroundColor Cyan
