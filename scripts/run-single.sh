#!/bin/bash
ticker=$1
if [ -z "$ticker" ]; then
    echo "Please provide a ticker symbol"
    exit 1
fi

docker-compose run --rm -e TICKER=$ticker scraper 