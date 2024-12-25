iraq-stock-scraper/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── scraper/
│   │   └── scraper.go         # Core scraping logic
│   └── utils/
│       ├── config.go          # Configuration handling
│       ├── logger.go          # Logging functionality
│       └── utils.go           # General utilities
├── configs/
│   └── config.yaml            # Application configuration
├── docker/
│   ├── Dockerfile            # Docker image definition
│   └── entrypoint.sh         # Docker entry point script
├── scripts/
│   ├── setup.ps1             # Project setup script
│   ├── run-docker.bat        # Docker execution script
│   └── run-local.bat         # Local execution script
├── test/
│   └── testdata/            # Test data directory
├── docs/
│   ├── README.md            # Project documentation
│   └── USAGE.md            # Usage guide
├── output/                  # Generated CSV files
├── logs/                   # Application logs
├── temp_builds/           # Temporary build files
├── .dockerignore         # Docker ignore rules
├── .gitignore           # Git ignore rules
├── docker-compose.yml   # Docker compose configuration
├── go.mod              # Go module file
├── go.sum             # Go dependencies checksums
└── TICKERS.csv        # Input file with ticker symbols 