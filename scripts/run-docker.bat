@echo off
IF "%1"=="" (
    echo Please provide either a ticker symbol or -file flag
    exit /b 1
)

IF "%1"=="-file" (
    docker-compose run --rm -e FILE=TICKERS.csv scraper
) ELSE (
    docker-compose run --rm -e TICKER=%1 scraper
) 