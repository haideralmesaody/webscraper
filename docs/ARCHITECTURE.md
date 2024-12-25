# Iraq Stock Exchange Scraper - Architecture Documentation

## Application Structure

### 1. Core Components

#### Command Layer (`cmd/`)
- **main.go**
  - Entry point of the application
  - Handles command-line arguments
  - Initializes core components (logger, config, scraper)
  - Manages the application lifecycle
  - Importance: Provides a clean separation between the entry point and business logic

#### Internal Components (`internal/`)

##### Scraper Package (`internal/scraper/`)
- **scraper.go**
  - Contains the core scraping logic
  - Manages browser automation
  - Handles data extraction and pagination
  - Implements error recovery and retry mechanisms
  - Importance: Core business logic for data extraction

##### Utils Package (`internal/utils/`)
- **config.go**
  - Handles configuration loading and parsing
  - Defines configuration structures
  - Validates configuration values
  - Importance: Centralizes application configuration management

- **logger.go**
  - Implements custom logging functionality
  - Handles both file and console logging
  - Provides different log levels (INFO, DEBUG, ERROR)
  - Importance: Ensures proper debugging and monitoring capabilities

- **utils.go**
  - Contains shared utility functions
  - Handles CSV file operations
  - Provides helper functions
  - Importance: Reduces code duplication and centralizes common functionality

### 2. Configuration

#### Configuration Files (`configs/`)
- **config.yaml**
  ```yaml
  scraper:
    timeout: 300    # Browser operation timeout
    retries: 3      # Number of retry attempts
    delay: 10       # Delay between operations
    maxPages: 4     # Maximum pages to scrape
    browser:
      headless: false  # Browser visibility
      debug: true      # Debug logging
  ```
  - Importance: 
    - Centralizes application settings
    - Allows easy configuration changes without code modification
    - Supports different environments (development, production)

### 3. Data Files

#### Input Data
- **TICKERS.csv**
  - Contains stock ticker information
  - Format: Ticker,Sector,Name
  - Used for batch processing
  - Importance: Provides structured input for batch operations

#### Output Data (`output/`)
- **{TICKER}_data.csv**
  - Generated for each processed ticker
  - Contains extracted stock data
  - Format: Date,Open,High,Low,Close,Volume,T.Shares,Trades
  - Importance: Stores extracted data in a structured format

### 4. Docker Support (`docker/`)
- **Dockerfile**
  - Defines the container image
  - Sets up the runtime environment
  - Configures dependencies
  - Importance: Ensures consistent deployment environment

- **entrypoint.sh**
  - Handles container startup
  - Manages application initialization
  - Importance: Provides proper container orchestration

### 5. Scripts (`scripts/`)
- **run-local.bat**
  - Executes application locally
  - Handles command-line arguments
  - Importance: Simplifies local development and testing

- **run-docker.bat**
  - Manages Docker-based execution
  - Handles environment variables
  - Importance: Simplifies containerized execution

### 6. Logging (`logs/`)
- Contains application logs
- Timestamp-based log files
- Tracks execution progress and errors
- Importance: Essential for monitoring and debugging

## Key Design Principles

1. **Separation of Concerns**
   - Clear separation between components
   - Modular design for easy maintenance
   - Independent package structure

2. **Configuration Management**
   - External configuration
   - Environment-specific settings
   - Easy customization

3. **Error Handling**
   - Comprehensive error logging
   - Retry mechanisms
   - Graceful failure handling

4. **Data Management**
   - Structured input/output
   - CSV format for compatibility
   - Clear data organization

5. **Deployment Flexibility**
   - Local execution support
   - Docker containerization
   - Environment independence

## Dependencies

- **chromedp**: Browser automation
- **yaml.v2**: Configuration parsing
- **Standard library**: Core functionality

## Future Considerations

1. **Scalability**
   - Parallel processing support
   - Distributed execution capability
   - Load balancing

2. **Monitoring**
   - Metrics collection
   - Performance monitoring
   - Health checks

3. **Data Management**
   - Database integration
   - Data validation
   - Historical data tracking 