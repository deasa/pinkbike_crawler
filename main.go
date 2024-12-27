package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"pinkbike-scraper/pkg/exporter"
	"pinkbike-scraper/pkg/listing"
	"pinkbike-scraper/pkg/scraper"
)

const (
	urlBase       = "https://www.pinkbike.com/buysell/list/"
	spreadsheetID = "16GYqn_Asp6_MhsJNAiMSphtUpJn6P1nNw-BRQG0s5Ik"
)

type ExchangeRateResponse struct {
	Rates map[string]float64
}

func main() {
	fileMode := flag.Bool("fileMode", false, "Set to true to read listings from a file instead of web scraping")
	filePath := flag.String("filePath", "", "The path to the file to read listings from when in file mode")
	exportToGoogleSheets := flag.Bool("exportToGoogleSheets", false, "Set to true to export listings to Google Sheets")
	exportToFile := flag.Bool("exportToFile", false, "Set to true to write listings to a file")
	exportToDB := flag.Bool("exportToDB", false, "Set to true to write listings to a database")
	bikeType := flag.String("bikeType", "enduro", "The type of bike to scrape listings for")
	numPages := flag.Int("numPages", 5, "The number of pages to scrape")
	headless := flag.Bool("headless", false, "Run browser in headless mode")
	flag.Parse()

	bikeTypeVal := getBikeType(*bikeType)

	exchangeRate, err := getCADtoUSDExchangeRate()
	if err != nil {
		log.Fatalf("could not get exchange rate: %v", err)
	}
	fmt.Printf("CAD to USD exchange rate: %f\n", exchangeRate)

	scraper, err := scraper.NewScraper(*filePath, *headless)
	if err != nil {
		log.Fatalf("could not create scraper: %v", err)
	}
	defer scraper.Close()

	var refinedListings []listing.Listing
	if *fileMode {
		refinedListings, err = scraper.ReadListingsFromFile()
		if err != nil {
			log.Fatalf("could not read listings from file: %v", err)
		}
	} else {
		rawListings, err := scraper.PerformWebScraping(urlBase, *numPages, bikeTypeVal)
		if err != nil {
			log.Fatalf("could not perform web scraping: %v", err)
		}
		for _, l := range rawListings {
			refinedListings = append(refinedListings, l.PostProcess(exchangeRate))
		}
	}

	fileName := getFileName(bikeTypeVal)

	var exporters []exporter.Exporter

	if *exportToFile {
		csvExp := exporter.NewCSVExporter(
			"runs/"+fileName,
			"runs/suspect_"+fileName,
		)
		exporters = append(exporters, csvExp)
	}

	if *exportToGoogleSheets {
		sheetsExp, err := exporter.NewSheetsExporter(
			"pinkbike-exporter-8bc8e681ffa1.json",
			spreadsheetID,
		)
		if err != nil {
			log.Fatalf("could not create sheets exporter: %v", err)
		}
		exporters = append(exporters, sheetsExp)
	}

	if *exportToDB {
		dbExp, err := exporter.NewDBExporter("listings.db")
		if err != nil {
			log.Fatalf("could not create database exporter: %v", err)
		}
		exporters = append(exporters, dbExp)
	}

	// Export using all configured exporters
	for _, exp := range exporters {
		if err := exp.Export(refinedListings); err != nil {
			log.Printf("export error: %v", err)
		}
	}
}

func getFileName(bikeType scraper.BikeType) string {
	bt := string(bikeType)
	fileName := fmt.Sprintf("%sListings%s.csv", bt, time.Now().Format("2006-01-02"))
	return fileName
}

func getBikeType(bikeType string) scraper.BikeType {
	var bikeTypeVal scraper.BikeType
	switch bikeType {
	case "enduro":
		bikeTypeVal = scraper.Enduro
	case "trail":
		bikeTypeVal = scraper.Trail
	case "xc":
		bikeTypeVal = scraper.XC
	case "dh":
		bikeTypeVal = scraper.DH
	default:
		log.Fatalf("invalid bike type: %s", bikeType)
	}
	return bikeTypeVal
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

// todo implement "a.k.a" for models and manufacturers so that they all get normalized to a single name
// priority is on the manufacturer though because we probably wont use the model name in the prediction
