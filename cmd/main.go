// Package main provides the entry point for the web scraper application.
// It supports scraping stock data from the Iraq Stock Exchange (ISX) for either
// a single ticker or multiple tickers from a CSV file.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"webscraper/internal/scraper"
	"webscraper/internal/utils"

	"github.com/chromedp/chromedp"
)

// processSingleTicker handles the scraping process for a single stock ticker.
// It fetches the stock data and saves it to a CSV file.
//
// Parameters:
//   - s: The scraper instance
//   - logger: Logger for tracking the process
//   - ticker: The stock ticker symbol to process
//
// Returns:
//   - error: Any error that occurred during processing
func processSingleTicker(s *scraper.Scraper, logger *utils.Logger, ticker string) error {
	logger.Info("Processing ticker: %s", ticker)

	// Fetch stock data from the website
	stockDataList, err := s.GetStockData(ticker)
	if err != nil {
		logger.Error("Error processing %s: %v", ticker, err)
		return err
	}

	// Save the fetched data to a CSV file
	err = s.SaveToCSV(ticker, stockDataList)
	if err != nil {
		logger.Error("Error saving data for %s: %v", ticker, err)
		return err
	}

	logger.Info("Successfully processed %s. Data saved to output/%s_data.csv", ticker, ticker)
	return nil
}

// processTickerList handles the scraping process for multiple stock tickers.
// It processes each ticker sequentially with a delay between requests.
//
// Parameters:
//   - s: The scraper instance
//   - logger: Logger for tracking the process
//   - tickers: Slice of ticker symbols to process
//
// Returns:
//   - error: Any error that occurred during processing
func processTickerList(s *scraper.Scraper, logger *utils.Logger, tickers []string) error {
	totalTickers := len(tickers)
	logger.Info("Starting to process %d tickers", totalTickers)

	for i, ticker := range tickers {
		logger.Info("Processing ticker %d/%d: %s", i+1, totalTickers, ticker)

		err := processSingleTicker(s, logger, ticker)
		if err != nil {
			logger.Error("Failed to process ticker %s: %v", ticker, err)
			time.Sleep(10 * time.Second)
			continue
		}

		if i < totalTickers-1 {
			logger.Debug("Waiting 10 seconds before next ticker")
			time.Sleep(10 * time.Second)
		}
	}

	// Generate and log aggregate performance report
	report := s.GetPerformanceTracker().GenerateAggregateReport()
	logger.Info("Aggregate Performance Report:\n%s", report)

	logger.Info("Completed processing %d tickers", totalTickers)
	return nil
}

// initializeScraper sets up the Chrome browser and creates necessary directories.
// It configures the browser with Arabic language support and creates the screenshots directory.
//
// Parameters:
//   - logger: Logger for tracking the initialization process
//   - config: Configuration for the scraper
//
// Returns:
//   - *scraper.Scraper: Configured scraper instance
//   - context.CancelFunc: Function to cancel the browser context
//   - error: Any error that occurred during initialization
func initializeScraper(logger *utils.Logger, config *utils.Config) (*scraper.Scraper, context.CancelFunc, error) {
	logger.Debug("Initializing Chrome with Arabic support")
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("lang", "ar"),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.NoSandbox,
		chromedp.Flag("headless", config.Scraper.Browser.Headless),
		chromedp.Flag("start-maximized", true),
		chromedp.Flag("enable-logging", config.Scraper.Browser.Debug),
		chromedp.Flag("v", "1"),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, _ := chromedp.NewContext(allocCtx, chromedp.WithLogf(logger.Debug))

	// Test browser launch
	if err := chromedp.Run(ctx, chromedp.Navigate("about:blank")); err != nil {
		logger.Error("Failed to launch browser: %v", err)
		return nil, cancel, err
	}

	// Create screenshots directory
	if err := os.MkdirAll("logs/screenshots", 0755); err != nil {
		logger.Error("Failed to create screenshots directory: %v", err)
		return nil, cancel, err
	}

	return scraper.NewScraper(logger, ctx, cancel, config), cancel, nil
}

func main() {
	startTime := time.Now()

	// Set custom temp directory before any other operations
	tempDir := "C:/GoProjects/webscraper/temp_builds"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("Warning: Failed to create temp directory: %v", err)
		// Try to use system temp directory as fallback
		tempDir = os.TempDir()
	}
	os.Setenv("GOTMPDIR", tempDir)

	// Define and parse command-line flags
	singleTicker := flag.String("ticker", "", "Single ticker to process")
	tickerFile := flag.String("file", "", "Path to CSV file containing tickers")
	flag.Parse()

	// Initialize logger for the application
	logger, err := utils.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	logger.Info("Starting web scraper application")

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	config, err := utils.LoadConfig(configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}

	// Update initializeScraper to use config
	s, cancel, err := initializeScraper(logger, config)
	if err != nil {
		logger.Fatal("Failed to initialize scraper: %v", err)
	}

	// Run preflight checks
	if err := s.PreflightCheck(); err != nil {
		logger.Fatal("Preflight check failed: %v", err)
	}

	// Ensure cleanup happens in the correct order
	defer func() {
		fmt.Println("Starting cleanup...")
		s.Close() // Close the scraper first
		cancel()  // Then cancel the context
		fmt.Println("Cleanup completed")
	}()

	// Process based on input flags
	if *singleTicker != "" {
		err = processSingleTicker(s, logger, *singleTicker)
		if err != nil {
			logger.Fatal("Failed to process ticker %s: %v", *singleTicker, err)
		}
	} else if *tickerFile != "" {
		tickers, err := utils.ReadTickersFromCSV(*tickerFile)
		if err != nil {
			logger.Fatal("Error reading CSV file %s: %v", *tickerFile, err)
		}

		logger.Info("Found %d tickers to process", len(tickers))
		err = processTickerList(s, logger, tickers)
		if err != nil {
			logger.Fatal("Failed to process ticker list: %v", err)
		}
	} else {
		logger.Fatal("No input specified. Use -ticker for single ticker or -file for ticker list")
	}

	// Log overall execution time
	duration := time.Since(startTime)
	logger.Info("Total execution time: %v", duration.Round(time.Second))

	logger.Info("Scraping completed successfully!")
}
