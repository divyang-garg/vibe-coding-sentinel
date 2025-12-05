# PowerShell wrapper for Sentinel
# Usage: .\sentinel.ps1 [command] [args]

param(
    [Parameter(Position=0)]
    [string]$Command = "help",
    
    [Parameter(ValueFromRemainingArguments=$true)]
    [string[]]$Args
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$SentinelExe = Join-Path $ScriptDir "sentinel.exe"

if (-not (Test-Path $SentinelExe)) {
    Write-Host "❌ Sentinel binary not found at $SentinelExe" -ForegroundColor Red
    Write-Host "   Run ./synapsevibsentinel.sh first to build the binary" -ForegroundColor Yellow
    exit 1
}

& $SentinelExe $Command $Args
exit $LASTEXITCODE



