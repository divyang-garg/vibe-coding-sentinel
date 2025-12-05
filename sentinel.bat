@echo off
REM Batch wrapper for Sentinel
REM Usage: sentinel.bat [command] [args]

setlocal

set SCRIPT_DIR=%~dp0
set SENTINEL_EXE=%SCRIPT_DIR%sentinel.exe

if not exist "%SENTINEL_EXE%" (
    echo ❌ Sentinel binary not found at %SENTINEL_EXE%
    echo    Run ./synapsevibsentinel.sh first to build the binary
    exit /b 1
)

"%SENTINEL_EXE%" %*
exit /b %ERRORLEVEL%



