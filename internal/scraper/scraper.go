package scraper

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"time"
	"webscraper/internal/utils"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// StockData represents the structure of our scraped data
type StockData struct {
	Date        string
	OpenPrice   string
	HighPrice   string
	LowPrice    string
	ClosePrice  string
	Volume      string
	TotalShares string
	NumTrades   string
}

type Scraper struct {
	logger      *utils.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	config      *utils.Config
	perfTracker *utils.PerformanceTracker
}

func NewScraper(logger *utils.Logger, ctx context.Context, cancel context.CancelFunc, config *utils.Config) *Scraper {
	return &Scraper{
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		config:      config,
		perfTracker: utils.NewPerformanceTracker(),
	}
}

func (s *Scraper) GetStockData(ticker string) ([]StockData, error) {
	url := fmt.Sprintf("http://www.isx-iq.net/isxportal/portal/companyprofilecontainer.html?currLanguage=en&companyCode=%s%%20&activeTab=0", ticker)
	fmt.Printf("Starting data extraction for ticker: %s\n", ticker)

	// Add dialog handler before navigation
	chromedp.ListenTarget(s.ctx, func(ev interface{}) {
		if ev, ok := ev.(*page.EventJavascriptDialogOpening); ok {
			s.logger.Debug("Dialog detected: %s", ev.Message)
			go func() {
				if err := chromedp.Run(s.ctx,
					page.HandleJavaScriptDialog(true),
				); err != nil {
					s.logger.Debug("Failed to handle dialog: %v", err)
				}
			}()
		}
	})

	// Navigate to the page
	err := chromedp.Run(s.ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate: %v", err)
	}

	// Set up date range and trigger search
	err = chromedp.Run(s.ctx,
		chromedp.Evaluate(`
			(() => {
				const dateInput = document.querySelector("#fromDate");
				dateInput.value = "01/01/2020";
				const event = new Event('change', { bubbles: true });
				dateInput.dispatchEvent(event);

				const searchButton = document.querySelector("#command > div.filterbox > div.button-all > input[type=button]");
				searchButton.click();
				return true;
			})()
		`, nil),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set date range: %v", err)
	}

	// Wait for table to load
	time.Sleep(2 * time.Second)

	var allStockData []StockData
	currentPage := 1
	maxPages := s.config.Scraper.MaxPages

	fmt.Printf("Starting data extraction, will process %d pages\n", maxPages)

	for currentPage <= maxPages {
		// Extract data from current page
		var pageData []StockData
		err = chromedp.Run(s.ctx,
			chromedp.Evaluate(`
				(() => {
					const table = document.getElementById('dispTable');
					const rows = table.querySelectorAll('tbody tr');
					return Array.from(rows).map(row => {
						const cells = row.querySelectorAll('td');
						return {
							Date: cells[9].textContent.trim(),
							OpenPrice: cells[7].textContent.trim(),
							HighPrice: cells[6].textContent.trim(),
							LowPrice: cells[5].textContent.trim(),
							ClosePrice: cells[8].textContent.trim(),
							Volume: cells[1].textContent.trim(),
							TotalShares: cells[2].textContent.trim(),
							NumTrades: cells[0].textContent.trim()
						};
					});
				})()
			`, &pageData),
		)
		if err != nil {
			fmt.Printf("Error extracting data from page %d: %v\n", currentPage, err)
			return nil, fmt.Errorf("failed to extract data from page %d: %v", currentPage, err)
		}

		fmt.Printf("Successfully extracted %d records from page %d\n", len(pageData), currentPage)
		allStockData = append(allStockData, pageData...)

		if currentPage >= maxPages {
			break
		}

		// Navigate to next page
		nextPage := currentPage + 1
		fmt.Printf("Navigating to page %d...\n", nextPage)
		err = chromedp.Run(s.ctx,
			chromedp.Evaluate(fmt.Sprintf(`
				(() => {
					doAjax('companyperformancehistoryfilter.html',
						'fromDate=01/01/2020&d-6716032-p=%d&toDate=23/12/2024&companyCode=%s',
						'ajxDspId');
					return true;
				})()
			`, nextPage, ticker), nil),
		)
		if err != nil {
			fmt.Printf("Failed to navigate to page %d: %v\n", nextPage, err)
			break
		}

		time.Sleep(time.Duration(s.config.Scraper.Delay) * time.Second)
		currentPage++
	}

	return allStockData, nil
}

func (s *Scraper) SaveToCSV(ticker string, data []StockData) error {
	if len(data) == 0 {
		return fmt.Errorf("no data to save")
	}

	// Create the output directory if it doesn't exist
	err := os.MkdirAll("output", 0755)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create CSV file
	filename := fmt.Sprintf("output/%s_data.csv", ticker)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %v", err)
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header with new column
	headers := []string{"Date", "Open", "High", "Low", "Close", "Volume", "T.Shares", "Trades"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %v", err)
	}

	// Write data including T.Shares
	for _, record := range data {
		row := []string{
			record.Date,
			record.OpenPrice,
			record.HighPrice,
			record.LowPrice,
			record.ClosePrice,
			record.Volume,
			record.TotalShares,
			record.NumTrades,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write record: %v", err)
		}
	}

	s.logger.Info("Successfully saved data to %s", filename)
	return nil
}

func (s *Scraper) Close() {
	fmt.Println("\nClosing browser...")
	if s.cancel != nil {
		// Create a new context with a short timeout for cleanup
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Try to close gracefully
		if err := chromedp.Run(ctx, chromedp.Stop()); err != nil {
			fmt.Printf("Error during graceful shutdown: %v\n", err)
		}

		s.cancel()
		time.Sleep(2 * time.Second)
		fmt.Println("Browser closed successfully")
	}
}

func (s *Scraper) GetPerformanceTracker() *utils.PerformanceTracker {
	return s.perfTracker
}

// PreflightCheck verifies all dependencies and configurations
func (s *Scraper) PreflightCheck() error {
	checks := []struct {
		name  string
		check func() error
	}{
		{"Config Validation", s.validateConfig},
		{"Directory Structure", s.checkDirectories},
		{"Browser Launch", s.testBrowserLaunch},
		{"Network Settings", s.testNetworkSettings},
	}

	for _, c := range checks {
		s.logger.Debug("Running preflight check: %s", c.name)
		if err := c.check(); err != nil {
			return fmt.Errorf("%s check failed: %v", c.name, err)
		}
		s.logger.Debug("%s check passed", c.name)
	}

	return nil
}

func (s *Scraper) validateConfig() error {
	if s.config == nil {
		return fmt.Errorf("configuration is nil")
	}
	if s.config.Scraper.Timeout <= 0 {
		return fmt.Errorf("invalid timeout value")
	}
	if s.config.Scraper.MaxPages <= 0 {
		return fmt.Errorf("invalid max pages value")
	}
	return nil
}

func (s *Scraper) checkDirectories() error {
	dirs := []string{
		"output",
		"logs",
		"temp_builds",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory: %v", dir, err)
		}
	}
	return nil
}

func (s *Scraper) testBrowserLaunch() error {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	return chromedp.Run(ctx, chromedp.Navigate("about:blank"))
}

func (s *Scraper) testNetworkSettings() error {
	ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		network.Enable(),
		network.SetCacheDisabled(true),
		emulation.SetUserAgentOverride("Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0.4472.124 Safari/537.36"),
	)
}

// Add browser refresh mechanism
func (s *Scraper) refreshBrowser() error {
	s.logger.Debug("Refreshing browser session")

	// Cancel old context
	if s.cancel != nil {
		s.cancel()
	}

	// Create new context and browser
	ctx, cancel := chromedp.NewContext(context.Background())
	s.ctx = ctx
	s.cancel = cancel

	// Test new browser
	err := chromedp.Run(ctx, chromedp.Navigate("about:blank"))
	if err != nil {
		return fmt.Errorf("failed to refresh browser: %v", err)
	}

	return nil
}

// Add this function before processTickerList
func processSingleTicker(s *Scraper, logger *utils.Logger, ticker string) error {
	logger.Info("Processing ticker: %s", ticker)

	// Get stock data
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

// Update processTickerList in main.go to handle browser refresh
func processTickerList(s *Scraper, logger *utils.Logger, tickers []string) error {
	totalTickers := len(tickers)
	logger.Info("Starting to process %d tickers", totalTickers)

	for i, ticker := range tickers {
		logger.Info("Processing ticker %d/%d: %s", i+1, totalTickers, ticker)

		// Refresh browser every 5 tickers
		if i > 0 && i%5 == 0 {
			logger.Debug("Performing browser refresh")
			if err := s.refreshBrowser(); err != nil {
				logger.Error("Failed to refresh browser: %v", err)
				time.Sleep(30 * time.Second) // Wait longer before next attempt
				continue
			}
		}

		err := processSingleTicker(s, logger, ticker)
		if err != nil {
			logger.Error("Failed to process ticker %s: %v", ticker, err)
			// If navigation fails, try refreshing the browser
			if err.Error() == "failed to navigate: context canceled" {
				logger.Debug("Navigation failed, refreshing browser")
				if err := s.refreshBrowser(); err != nil {
					logger.Error("Failed to refresh browser: %v", err)
				}
			}
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
