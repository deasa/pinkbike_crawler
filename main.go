package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"pinkbike-scraper/pkg/exporter"
	"pinkbike-scraper/pkg/listing"
	"pinkbike-scraper/pkg/scraper"
)

const (
	urlBase = "https://www.pinkbike.com/buysell/list/"
)

type ExchangeRateResponse struct {
	Rates map[string]float64
}

func main() {
	fileMode := flag.Bool("fileMode", false, "Set to true to read listings from a file instead of web scraping")
	filePath := flag.String("filePath", "", "The path to the file to read listings from when in file mode")
	exportToGoogleSheets := flag.Bool("exportToGoogleSheets", false, "Set to true to export listings to Google Sheets")
	exportToFile := flag.Bool("exportToFile", false, "Set to true to write listings to a file")
	flag.Parse()

	exchangeRate, err := getCADtoUSDExchangeRate()
	if err != nil {
		log.Fatalf("could not get exchange rate: %v", err)
	}
	fmt.Printf("CAD to USD exchange rate: %f\n", exchangeRate)

	var listings []listing.RawListing
	if *fileMode {
		listings, err = scraper.ReadListingsFromFile(*filePath)
		if err != nil {
			log.Fatalf("could not read listings from file: %v", err)
		}
	} else {
		listings, err = scraper.PerformWebScraping(urlBase, 700)
		if err != nil {
			log.Fatalf("could not perform web scraping: %v", err)
		}
	}

	refinedListings := make([]listing.Listing, 0, len(listings))
	for _, listing := range listings {
		refinedListings = append(refinedListings, listing.PostProcess(exchangeRate))
	}

	if *exportToFile {
		err = exporter.WriteListingsToFile(refinedListings, "goodListingsCache.csv", "suspectListingsCache.csv")
		if err != nil {
			log.Fatalf("could not write listings to file: %v", err)
		}
	}

	if *exportToGoogleSheets {
		err = exporter.ExportToGoogleSheets(refinedListings)
		if err != nil {
			log.Fatalf("could not export listings to Google Sheets: %v", err)
		}
	}
}

func getCADtoUSDExchangeRate() (float64, error) {
	resp, err := http.Get("https://api.exchangerate-api.com/v4/latest/CAD")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data ExchangeRateResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, err
	}

	return data.Rates["USD"], nil
}

// todo scrape trail bike data

// todo research training a machine learning model on this data to predict the price of a bike
