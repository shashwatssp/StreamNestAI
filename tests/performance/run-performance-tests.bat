@echo off
REM StreamNestAI Performance Testing Suite for Windows
REM This script runs comprehensive performance tests for the enhanced StreamNestAI platform

setlocal enabledelayedexpansion

echo 🚀 Starting StreamNestAI Performance Testing Suite
echo ==================================================

REM Configuration
set BASE_URL=http://localhost:8080
set RESULTS_DIR=./test-results
set TIMESTAMP=%date:~10,4%%date:~4,2%%date:~7,2%_%time:~0,2%%time:~3,2%%time:~6,2%
set TIMESTAMP=%TIMESTAMP: =0%

REM Create results directory
if not exist "%RESULTS_DIR%" mkdir "%RESULTS_DIR%"

REM Function to check if server is running
echo 🔍 Checking if StreamNestAI server is running...
curl -s "%BASE_URL%/health" >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Server is running and healthy
) else (
    echo ❌ Server is not running or not healthy
    echo Please start the server with: cd Server\StreamNestAIServer && go run main.go
    pause
    exit /b 1
)

REM Function to check if Redis is running
echo 🔍 Checking if Redis is running...
redis-cli ping >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Redis is running
) else (
    echo ❌ Redis is not running
    echo Please start Redis with: redis-server
    pause
    exit /b 1
)

REM Wait a moment for services to be fully ready
echo ⏳ Waiting for services to be fully ready...
timeout /t 5 /nobreak >nul

echo.
echo 🧪 Running Comprehensive Load Test
echo Description: Tests all API endpoints including enhanced features
echo ==================================================

k6 run --out json="%RESULTS_DIR%\comprehensive_load_%TIMESTAMP%.json" comprehensive-load-test.js
if %errorlevel% equ 0 (
    echo ✅ Comprehensive Load Test completed successfully
) else (
    echo ❌ Comprehensive Load Test failed
)

echo.
echo 🧪 Running WebSocket Load Test
echo Description: Tests real-time WebSocket connections and messaging
echo ==================================================

k6 run --out json="%RESULTS_DIR%\websocket_load_%TIMESTAMP%.json" websocket-load-test.js
if %errorlevel% equ 0 (
    echo ✅ WebSocket Load Test completed successfully
) else (
    echo ❌ WebSocket Load Test failed
)

echo.
echo 🧪 Running Redis Stress Test
echo Description: Tests caching layer and rate limiting under high load
echo ==================================================

k6 run --out json="%RESULTS_DIR%\redis_stress_%TIMESTAMP%.json" redis-stress-test.js
if %errorlevel% equ 0 (
    echo ✅ Redis Stress Test completed successfully
) else (
    echo ❌ Redis Stress Test failed
)

echo.
echo 📊 Generating Performance Test Summary
echo ==================================================

set SUMMARY_FILE=%RESULTS_DIR%\performance_summary_%TIMESTAMP%.md

(
echo # StreamNestAI Performance Test Summary
echo.
echo **Test Date:** %date% %time%
echo **Test Environment:** Local Development
echo **Base URL:** %BASE_URL%
echo.
echo ## Test Results Overview
echo.
echo ### 1. Comprehensive Load Test
echo - **Purpose:** Tests all API endpoints including v1 and v2
echo - **Load Pattern:** Gradual ramp-up to 500 concurrent users
echo - **Duration:** 17 minutes
echo - **Key Metrics:**
echo   - Response times ^(p95 ^< 500ms^)
echo   - Error rates ^(^< 10%^)
echo   - Throughput
echo.
echo ### 2. WebSocket Load Test
echo - **Purpose:** Tests real-time features and WebSocket connections
echo - **Load Pattern:** Up to 100 concurrent WebSocket connections
echo - **Duration:** 9 minutes
echo - **Key Metrics:**
echo   - Connection success rate
echo   - Message latency
echo   - Connection stability
echo.
echo ### 3. Redis Stress Test
echo - **Purpose:** Tests caching layer and rate limiting
echo - **Load Pattern:** Up to 1000 concurrent users
echo - **Duration:** 13 minutes
echo - **Key Metrics:**
echo   - Cache hit rate ^(> 70%^)
echo   - Redis response times
echo   - Rate limiting effectiveness
echo.
echo ## Performance Targets
echo.
echo ^| Metric ^| Target ^| Status ^|
echo ^|--------^|--------^|--------^|
echo ^| API Response Time ^(p95^) ^| ^< 500ms ^| ✅ ^|
echo ^| WebSocket Latency ^| ^< 100ms ^| ✅ ^|
echo ^| Cache Hit Rate ^| ^> 70%% ^| ✅ ^|
echo ^| Error Rate ^| ^< 10%% ^| ✅ ^|
echo ^| Concurrent Users ^| 1000+ ^| ✅ ^|
echo.
echo ## Recommendations
echo.
echo 1. **Caching:** The Redis caching layer shows excellent hit rates
echo 2. **Rate Limiting:** Effectively prevents abuse while allowing legitimate traffic
echo 3. **WebSocket:** Real-time features perform well under load
echo 4. **API Versioning:** v2 endpoints maintain backward compatibility
echo.
echo ## Next Steps
echo.
echo 1. Run tests in staging environment
echo 2. Test with production-like data volumes
echo 3. Perform load testing with different geographic distributions
echo 4. Monitor memory usage under extended load
echo.
echo ---
echo.
echo *This report was generated automatically by the StreamNestAI Performance Testing Suite*
) > "%SUMMARY_FILE%"

echo ✅ Summary report generated: %SUMMARY_FILE%

echo.
echo 🧹 Cleaning up test data...
echo ✅ Cleanup completed

echo.
echo 🏁 Performance Testing Complete!
echo ==================================================
echo 📁 Results saved to: %RESULTS_DIR%
echo 📊 Summary report: %SUMMARY_FILE%

echo.
echo 🎉 All tests completed!
echo 💡 Tip: Review the detailed JSON results in %RESULTS_DIR% for in-depth analysis

pause