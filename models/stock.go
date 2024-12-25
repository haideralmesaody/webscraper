// Package models defines the data structures used in the application.
package models

// StockData represents the structure of stock data.
type StockData struct {
	Date      string
	Close     string
	Open      string
	High      string
	Low       string
	Volume    string // TShares Volume
	NumTrades string // No.Trades
}
