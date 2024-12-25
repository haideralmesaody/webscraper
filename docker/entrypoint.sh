#!/bin/bash
if [ -n "$TICKER" ]; then
    ./webscraper -ticker "$TICKER"
elif [ -n "$FILE" ]; then
    ./webscraper -file "$FILE"
else
    echo "Please provide either TICKER or FILE environment variable"
    exit 1
fi 