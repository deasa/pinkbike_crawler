package main

import (
	"log"

	"pinkbike-scraper/pkg/exporter"
	"pinkbike-scraper/pkg/scraper"

	"github.com/playwright-community/playwright-go"
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

	if _, err = page.Goto("https://www.pinkbike.com/buysell/list/?category=2"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	listings, err := scraper.ScrapePage(page)
	if err != nil {
		log.Fatalf("could not scrape page: %v", err)
	}

	err = exporter.WriteListingsToFile(listings, "listings.csv")
	if err != nil {
		log.Fatalf("could not write listings to file: %v", err)
	}

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}

	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}
