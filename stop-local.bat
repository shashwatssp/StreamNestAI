@echo off
REM StreamNestAI Local Development Stop Script for Windows
REM This script stops all running services

setlocal enabledelayedexpansion

echo.
echo [INFO] Stopping StreamNestAI Local Development Environment...
echo.

REM Function to kill processes by port
:kill_by_port
set port=%1
set service_name=%2

for /f "tokens=5" %%a in ('netstat -aon ^| find ":%port%" ^| find "LISTENING"') do (
    echo [INFO] Killing %service_name% process on port %port% (PID: %%a)...
    taskkill /F /PID %%a >nul 2>&1
    echo [SUCCESS] %service_name% on port %port% killed
)
goto :eof

REM Kill processes by name
echo [INFO] Stopping Backend processes...
taskkill /F /IM "go.exe" /FI "WINDOWTITLE eq *main.go*" >nul 2>&1
taskkill /F /IM "streamnest-server.exe" >nul 2>&1

echo [INFO] Stopping Frontend processes...
taskkill /F /IM "node.exe" /FI "WINDOWTITLE eq *vite*" >nul 2>&1

echo [INFO] Stopping Cloudflare Worker processes...
taskkill /F /IM "node.exe" /FI "WINDOWTITLE eq *wrangler*" >nul 2>&1

echo.

REM Kill any remaining processes on specific ports
echo [INFO] Checking for remaining processes on ports...

call :kill_by_port 8080 "Backend"
call :kill_by_port 5173 "Frontend"
call :kill_by_port 8787 "Cloudflare Worker"
call :kill_by_port 9090 "Metrics"

echo.

REM Clean up any remaining processes
echo [INFO] Cleaning up any remaining processes...

REM Kill any Go processes related to our project
for /f "tokens=2" %%i in ('tasklist /FI "IMAGENAME eq go.exe" /FO CSV ^| find "go.exe"') do (
    taskkill /F /PID %%i >nul 2>&1
)

REM Kill any Node processes related to vite or wrangler
for /f "tokens=2" %%i in ('tasklist /FI "IMAGENAME eq node.exe" /FO CSV ^| find "node.exe"') do (
    taskkill /F /PID %%i >nul 2>&1
)

echo.
echo [SUCCESS] 🛑 All StreamNestAI services have been stopped
echo.

REM Show final status
echo [INFO] Final port check:

for %%p in (8080 5173 8787 9090) do (
    netstat -an | find ":%%p " | find "LISTENING" >nul
    if !errorlevel! equ 0 (
        echo [WARNING] Port %%p is still in use
    ) else (
        echo [SUCCESS] Port %%p is free
    )
)

echo.
echo [SUCCESS] Cleanup completed!
echo.
pause