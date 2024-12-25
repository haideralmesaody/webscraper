@echo off
set TEMP_DIR=C:\GoProjects\webscraper\temp_builds

REM Create temp directory if it doesn't exist
mkdir "%TEMP_DIR%" 2>nul

set GOTMPDIR=%TEMP_DIR%

IF "%1"=="" (
    echo Please provide either a ticker symbol or -file flag
    exit /b 1
)

IF "%1"=="-file" (
    go run cmd/main.go -file TICKERS.csv
) ELSE (
    go run cmd/main.go -ticker %1
) 