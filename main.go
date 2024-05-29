package main

import (
	"flag"
	"log"

	"pinkbike-scraper/pkg/exporter"
	"pinkbike-scraper/pkg/listing"
	"pinkbike-scraper/pkg/scraper"
)

const (
	urlBase = "https://www.pinkbike.com/buysell/list/"
)

func main() {
	fileMode := flag.Bool("fileMode", false, "Set to true to read listings from a file instead of web scraping")
    filePath := flag.String("filePath", "", "The path to the file to read listings from when in file mode")
    exportToGoogleSheets := flag.Bool("exportToGoogleSheets", false, "Set to true to export listings to Google Sheets")
	exportToFile := flag.Bool("writeToFile", false, "Set to true to write listings to a file")
	flag.Parse()

	var listings []listing.RawListing
	var err error
    if *fileMode {
        listings, err = exporter.ReadListingsFromFile(*filePath)
        if err != nil {
            log.Fatalf("could not read listings from file: %v", err)
        }
    } else {
		listings, err = scraper.PerformWebScraping(urlBase, 5)
		if err != nil {
			log.Fatalf("could not perform web scraping: %v", err)
		}
    }

	refinedListings := make([]listing.Listing, 0, len(listings))
	for _, listing := range listings {
		refinedListings = append(refinedListings, listing.PostProcess())
	}

	if *exportToFile {
		err = exporter.WriteListingsToFile(refinedListings, "listingsCache.csv")
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

// todo add a mode that will skip the scraping and read from a file
// todo scrape trail bike data

// todo research training a machine learning model on this data to predict the price of a bike