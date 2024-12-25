package utils

import (
	"encoding/csv"
	"os"
)

// ReadTickersFromCSV reads ticker symbols from a CSV file.
func ReadTickersFromCSV(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var tickers []string
	for _, record := range records[1:] { // Skip header
		tickers = append(tickers, record[0]) // Assuming ticker is in the first column
	}

	return tickers, nil
}
