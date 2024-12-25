package main

import (
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	// Create file server for static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Serve the index.html file
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join("static", "index.html"))
			return
		}
		http.NotFound(w, r)
	})

	// Serve data files (CSV files from output directory)
	dataFs := http.FileServer(http.Dir("../output"))
	http.Handle("/data/", http.StripPrefix("/data/", dataFs))

	// Serve TICKERS.csv from root directory
	tickersFs := http.FileServer(http.Dir(".."))
	http.Handle("/tickers/", http.StripPrefix("/tickers/", tickersFs))

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
