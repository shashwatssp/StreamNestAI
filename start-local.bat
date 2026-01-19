@echo off
REM StreamNestAI Local Development Startup Script for Windows
REM This script starts all services for local development

setlocal enabledelayedexpansion

echo.
echo [INFO] Starting StreamNestAI Local Development Environment...
echo.

REM Check prerequisites
echo [INFO] Checking prerequisites...

where node >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Node.js is not installed
    pause
    exit /b 1
)

where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed
    pause
    exit /b 1
)

echo [SUCCESS] Prerequisites check completed
echo.

REM Create logs directory
if not exist logs mkdir logs

REM Setup environment files if they don't exist
if not exist "Client\StreamNestAIClient\.env" (
    echo [INFO] Setting up frontend environment file...
    copy "Client\StreamNestAIClient\.env.local" "Client\StreamNestAIClient\.env" >nul
    echo [SUCCESS] Frontend environment file created
)

if not exist "Server\StreamNestAIServer\.env" (
    echo [INFO] Setting up backend environment file...
    copy "Server\StreamNestAIServer\.env.local" "Server\StreamNestAIServer\.env" >nul
    echo [SUCCESS] Backend environment file created
)

REM Install dependencies if needed
if not exist "Client\StreamNestAIClient\node_modules" (
    echo [INFO] Installing frontend dependencies...
    cd Client\StreamNestAIClient
    call npm install
    cd ..\..
    echo [SUCCESS] Frontend dependencies installed
)

if not exist "MCP-and-CDN\node_modules" (
    echo [INFO] Installing Cloudflare Worker dependencies...
    cd MCP-and-CDN
    call npm install
    cd ..
    echo [SUCCESS] Cloudflare Worker dependencies installed
)

echo.
echo [INFO] Starting services...
echo.

REM Start Backend
echo [INFO] Starting Backend server...
cd Server\StreamNestAIServer
start /B cmd /c "go run main.go > ..\..\logs\backend.log 2>&1"
cd ..\..
echo [SUCCESS] Backend started

REM Wait for Backend to be ready
echo [INFO] Waiting for Backend to start...
:wait_backend
timeout /t 2 /nobreak >nul
curl -s http://localhost:8080/health >nul 2>&1
if %errorlevel% neq 0 goto wait_backend
echo [SUCCESS] Backend is ready!
echo.

REM Start Cloudflare Worker
echo [INFO] Starting Cloudflare Worker...
cd MCP-and-CDN
start /B cmd /c "npx wrangler dev --config wrangler.local.toml > ..\logs\cloudflare.log 2>&1"
cd ..
echo [SUCCESS] Cloudflare Worker started

REM Wait for Cloudflare Worker to be ready
timeout /t 5 /nobreak >nul
echo [INFO] Waiting for Cloudflare Worker to start...
:wait_cloudflare
timeout /t 2 /nobreak >nul
curl -s http://localhost:8787 >nul 2>&1
if %errorlevel% neq 0 goto wait_cloudflare
echo [SUCCESS] Cloudflare Worker is ready!
echo.

REM Start Frontend
echo [INFO] Starting Frontend...
cd Client\StreamNestAIClient
start /B cmd /c "npm run dev > ..\..\logs\frontend.log 2>&1"
cd ..\..
echo [SUCCESS] Frontend started

REM Wait for Frontend to be ready
timeout /t 5 /nobreak >nul
echo [INFO] Waiting for Frontend to start...
:wait_frontend
timeout /t 2 /nobreak >nul
curl -s http://localhost:5173 >nul 2>&1
if %errorlevel% neq 0 goto wait_frontend
echo [SUCCESS] Frontend is ready!
echo.

REM Final status check
echo [INFO] Performing final service health check...
echo.

curl -s http://localhost:8080/health >nul 2>&1
if %errorlevel% equ 0 (
    echo [SUCCESS] Backend API is running
) else (
    echo [WARNING] Backend API is not responding
)

curl -s http://localhost:5173 >nul 2>&1
if %errorlevel% equ 0 (
    echo [SUCCESS] Frontend is running
) else (
    echo [WARNING] Frontend is not responding
)

curl -s http://localhost:8787 >nul 2>&1
if %errorlevel% equ 0 (
    echo [SUCCESS] Cloudflare Worker is running
) else (
    echo [WARNING] Cloudflare Worker is not responding
)

echo.
echo [SUCCESS] 🎉 StreamNestAI Local Development Environment is ready!
echo.
echo 📍 Service URLs:
echo    • Frontend:        http://localhost:5173
echo    • Backend API:     http://localhost:8080
echo    • Backend Health:  http://localhost:8080/health
echo    • Cloudflare:      http://localhost:8787
echo    • Metrics:         http://localhost:8080/metrics
echo.
echo 📝 Logs:
echo    • Backend:   logs\backend.log
echo    • Frontend:  logs\frontend.log
echo    • Cloudflare: logs\cloudflare.log
echo.
echo 🛑 To stop all services, run: stop-local.bat
echo.
echo [INFO] Services are running in the background. Press any key to exit this window...
pause >nul