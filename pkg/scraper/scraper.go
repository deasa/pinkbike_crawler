package scraper

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/playwright-community/playwright-go"

	"pinkbike-scraper/pkg/listing"
)

var (
	Enduro BikeType = "enduro"
	Trail  BikeType = "trail"
	XC     BikeType = "xc"
	DH     BikeType = "dh"
)

// biketype enum
type BikeType string

func ReadListingsFromFile(filePath string) ([]listing.Listing, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read file: %v", err)
	}

	listings := make([]listing.Listing, 0, len(records))
	for _, record := range records {
		l := listing.Listing{
			Title:         record[0],
			Year:          record[1],
			Price:         record[2],
			Currency:      record[3],
			Condition:     record[4],
			FrameSize:     record[5],
			WheelSize:     record[6],
			FrontTravel:   record[7],
			RearTravel:    record[8],
			FrameMaterial: record[9],
		}

		listings = append(listings, l)
	}

	return listings, nil
}

func PerformWebScraping(url string, numPages int, bikeType BikeType) ([]listing.RawListing, error) {
	err := playwright.Install()
	if err != nil {
		log.Fatalf("could not install playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}

	defer func() {
		if err = pw.Stop(); err != nil {
			log.Fatalf("could not stop Playwright: %v", err)
		}
	}()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})

	defer func() {
		if err = browser.Close(); err != nil {
			log.Fatalf("could not close browser: %v", err)
		}
	}()

	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}

	if _, err = page.Goto(getListingsUrl(url, bikeType)); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	fmt.Println("Scraping page: 1")

	listings, nextPageURL, err := scrapePage(page)
	if err != nil {
		log.Fatalf("could not scrape page: %v", err)
	}

	var newListings []listing.RawListing
	pages := 1
	for nextPageURL != "" && pages < numPages {
		pages++
		fmt.Println("Scraping page: ", pages)

		if _, err = page.Goto(url + nextPageURL); err != nil {
			log.Fatalf("could not goto: %v", err)
		}

		newListings, nextPageURL, err = scrapePage(page)
		if err != nil {
			log.Fatalf("could not scrape page: %v", err)
		}

		listings = append(listings, newListings...)
	}

	return listings, nil
}

func getListingsUrl(urlBase string, bikeType BikeType) string {
	switch bikeType {
	case Enduro:
		return urlBase + "/?category=2"
	case Trail:
		return urlBase + "/?category=102"
	case XC:
		return urlBase + "/?category=75"
	case DH:
		return urlBase + "/?category=1"
	default:
		log.Fatalf("invalid bike type: %s", bikeType)
		return ""
	}
}

// todo implement an auto-dedupe function that will compare each parsed listing from the page and will not add it to the list if it already exists

func scrapePage(page playwright.Page) ([]listing.RawListing, string, error) {
	entries, err := page.Locator("tr.bsitem-table").All()
	if err != nil {
		return nil, "", fmt.Errorf("could not get entries: %v", err)
	}

	var sanitizedListings []listing.RawListing
	for _, entry := range entries {
		sanitizedListings = append(sanitizedListings, getListing(entry))
	}

	// Find the "Next Page" link
	nextPageLink := page.Locator(`xpath=//a[text()='Next']`)

	// Get the URL of the "Next Page" link
	nextPageURL, err := nextPageLink.GetAttribute("href")
	if err != nil {
		// If an error occurred, the link was not found
		nextPageURL = ""
	}

	return sanitizedListings, nextPageURL, nil
}

func getListing(entry playwright.Locator) listing.RawListing {
	title, err := entry.Locator("div.bsitem-title > a").TextContent()
	if err != nil {
		fmt.Println("\tcould not get title")
	}

	url, err := entry.Locator("div.bsitem-title > a").GetAttribute("href")
	if err != nil {
		fmt.Println("\tcould not get url")
	}

	condition, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Condition")]]`).InnerText(playwright.LocatorInnerTextOptions{Timeout: playwright.Float(1000)})
	if err != nil {
		fmt.Println("\tcould not get condition")
	}

	frameSize, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Frame Size")]]`).TextContent(playwright.LocatorTextContentOptions{Timeout: playwright.Float(1000)})
	if err != nil {
		fmt.Println("\tcould not get frame size")
	}

	wheelSize, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Wheel Size")]]`).TextContent(playwright.LocatorTextContentOptions{Timeout: playwright.Float(1000)})
	if err != nil {
		fmt.Println("\tcould not get wheel size")
	}

	frontTravel, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Front Travel")]]`).TextContent(playwright.LocatorTextContentOptions{Timeout: playwright.Float(1000)})
	if err != nil {
		fmt.Println("\tcould not get front travel")
	}

	rearTravel, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Rear Travel")]]`).TextContent(playwright.LocatorTextContentOptions{Timeout: playwright.Float(1000)})
	if err != nil {
		fmt.Println("\tcould not get rear travel")
	}

	material, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Material")]]`).TextContent(playwright.LocatorTextContentOptions{Timeout: playwright.Float(1000)})
	if err != nil {
		fmt.Println("\tcould not get material")
	}

	price, err := entry.Locator("td.bsitem-price > b").TextContent(playwright.LocatorTextContentOptions{Timeout: playwright.Float(1000)})
	if err != nil {
		fmt.Println("\tcould not get price")
	}

	l := listing.RawListing{
		Title:         title,
		Price:         price,
		Condition:     condition,
		FrameSize:     frameSize,
		WheelSize:     wheelSize,
		FrontTravel:   frontTravel,
		RearTravel:    rearTravel,
		FrameMaterial: material,
		URL:           url,
	}

	return sanitize(l)
}

// Sanitize will remove spaces and labels from the listing
func sanitize(l listing.RawListing) listing.RawListing {
	newL := listing.RawListing{}

	newL.Title = strings.TrimSpace(l.Title)
	newL.Price = strings.TrimSpace(l.Price)
	newL.Condition = parseItemDetail(l.Condition, "Condition :")
	newL.FrameSize = parseItemDetail(l.FrameSize, "Frame Size :")
	newL.WheelSize = parseItemDetail(l.WheelSize, "Wheel Size :")
	newL.FrontTravel = parseItemDetail(l.FrontTravel, "Front Travel :")
	newL.RearTravel = parseItemDetail(l.RearTravel, "Rear Travel :")
	newL.FrameMaterial = parseItemDetail(l.FrameMaterial, "Material :")
	newL.URL = strings.TrimSpace(l.URL)

	return newL
}

func parseItemDetail(detail, label string) string {
	split := strings.Split(detail, label)
	if len(split) < 2 {
		return ""
	}

	s := strings.TrimSpace(split[1])
	s = strings.ReplaceAll(s, `"`, "")

	return s
}
