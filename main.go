package main

import (
	"log"

	"pinkbike-scraper/pkg/exporter"
	"pinkbike-scraper/pkg/scraper"

	"github.com/playwright-community/playwright-go"
)

const (
	urlBase = "https://www.pinkbike.com/buysell/list/"
)

func main() {
	err := playwright.Install()
	if err != nil {
		log.Fatalf("could not install playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})

	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	if _, err = page.Goto(urlBase+"?category=2"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}
	
	listings, nextPageURL, err := scraper.ScrapePage(page)
	if err != nil {
		log.Fatalf("could not scrape page: %v", err)
	}

	var newListings []scraper.Listing
	pages := 1
	for nextPageURL != "" && pages <= 2 {
		if _, err = page.Goto(urlBase+nextPageURL); err != nil {
			log.Fatalf("could not goto: %v", err)
		}

		newListings, nextPageURL, err = scraper.ScrapePage(page)
		if err != nil {
			log.Fatalf("could not scrape page: %v", err)
		}

		listings = append(listings, newListings...)
		pages++
	}

	err = exporter.WriteListingsToFile(listings, "listings.csv")
	if err != nil {
		log.Fatalf("could not write listings to file: %v", err)
	}

	err = exporter.ExportToGoogleSheets(listings)
	if err != nil {
		log.Fatalf("could not export listings to Google Sheets: %v", err)
	}

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}

	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}
