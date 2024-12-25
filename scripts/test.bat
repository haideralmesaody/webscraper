@echo off
echo Running system checks...

REM Check directories
if not exist "configs" mkdir configs
if not exist "output" mkdir output
if not exist "logs" mkdir logs
if not exist "temp_builds" mkdir temp_builds

REM Check config file
if not exist "configs/config.yaml" (
    echo Error: config.yaml not found
    exit /b 1
)

REM Check Go installation
go version > nul 2>&1
if errorlevel 1 (
    echo Error: Go is not installed
    exit /b 1
)

REM Run a single ticker test
echo Running test with BBOB...
.\scripts\run-local.bat BBOB

if errorlevel 1 (
    echo Test failed
    exit /b 1
)

echo All checks passed 