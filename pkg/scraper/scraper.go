package scraper

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"

	"pinkbike-scraper/pkg/db"
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

// Scraper holds configuration for scraping operations
type Scraper struct {
	filePath string
	headless bool
	pw       *playwright.Playwright
	browser  playwright.Browser
	baseUrl  string
	dbWorker *db.DBWorker
	page     playwright.Page
}

// NewScraper creates and returns a new Scraper instance
func NewScraper(filePath string, headless bool, baseUrl string, bikeType BikeType, dbWorker *db.DBWorker) (*Scraper, error) {
	err := playwright.Install()
	if err != nil {
		return nil, fmt.Errorf("could not install playwright: %v", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		pw.Stop()
		return nil, fmt.Errorf("could not launch browser: %v", err)
	}

	page, err := browser.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}

	url := getListingsUrl(baseUrl, bikeType)

	resp, err := page.Goto(url)
	if err != nil {
		return nil, fmt.Errorf("could not goto: %v", err)
	}

	if resp.Status() != 200 {
		return nil, fmt.Errorf("could not get 200 status: %v", resp.Status())
	}

	return &Scraper{
		filePath: filePath,
		headless: headless,
		pw:       pw,
		browser:  browser,
		baseUrl:  baseUrl,
		page:     page,
		dbWorker: dbWorker,
	}, nil
}

// Close cleanly shuts down the scraper
func (s *Scraper) Close() error {
	if err := s.browser.Close(); err != nil {
		return fmt.Errorf("could not close browser: %v", err)
	}
	if err := s.pw.Stop(); err != nil {
		return fmt.Errorf("could not stop Playwright: %v", err)
	}
	return nil
}

// PerformWebScraping performs the web scraping operation
func (s *Scraper) PerformWebScraping(numPages int) ([]listing.RawListing, error) {
	fmt.Println("Scraping page: 1")

	listings, nextPageURL, err := scrapePage(s.page)
	if err != nil {
		return nil, fmt.Errorf("could not scrape page: %v", err)
	}

	var newListings []listing.RawListing
	pages := 1
	for nextPageURL != "" && pages < numPages {
		pages++
		fmt.Println("Scraping page: ", pages)

		if _, err = s.page.Goto(s.baseUrl + nextPageURL); err != nil {
			return nil, fmt.Errorf("could not goto: %v", err)
		}

		newListings, nextPageURL, err = scrapePage(s.page)
		if err != nil {
			return nil, fmt.Errorf("could not scrape page: %v", err)
		}

		listings = append(listings, newListings...)
	}

	return listings, nil
}

func (s *Scraper) GetDetailedListings(listings []listing.Listing) ([]listing.Listing, error) {
	page, err := s.browser.NewPage()
	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}

	listingsWithDetails := []listing.Listing{}

	for _, l := range listings {
		// if listing exists in db, and has details, skip
		exists, err := s.dbWorker.ListingExistsWithDetails(l.Hash)
		if err != nil {
			fmt.Printf("could not check if listing exists: %v", err)
			continue
		}

		if exists {
			continue
		}

		// if listing exists in db, and does not have details, perform details scrape
		resp, err := page.Goto(l.URL)
		if err != nil {
			fmt.Printf("could not goto: %v", err)
			continue
		}

		if resp.Status() != 200 {
			fmt.Printf("could not get 200 status: %v", resp.Status())
			continue
		}

		details, err := s.detailsScrape(page)
		if err != nil {
			fmt.Printf("could not scrape details: %v", err)
			continue
		}

		l.Details = *details
		listingsWithDetails = append(listingsWithDetails, l)
	}

	return listingsWithDetails, nil
}

func (s *Scraper) detailsScrape(page playwright.Page) (*listing.ListingDetails, error) {
	details := listing.ListingDetails{}

	sellerType, err := page.Locator(`xpath=//div[contains(@class, "buysell-details-column")]//b[contains(text(), "Seller Type")]/parent::*`).TextContent(playwright.LocatorTextContentOptions{Timeout: playwright.Float(1000)})
	if err != nil {
		return nil, fmt.Errorf("\tcould not get seller type: %v", err)
	}

	originalPostDate, err := page.Locator(`xpath=//div[contains(@class, "buysell-details-column")]//b[contains(text(), "Original Post Date")]//parent::div`).TextContent(playwright.LocatorTextContentOptions{Timeout: playwright.Float(1000)})
	if err != nil {
		return nil, fmt.Errorf("\tcould not get original post date: %v", err)
	}

	dateRegex := regexp.MustCompile(`Original Post Date:\s*((?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)-\d{2}-\d{4})`)
	matches := dateRegex.FindStringSubmatch(originalPostDate)
	if len(matches) < 2 {
		return nil, fmt.Errorf("\tcould not find date in string: %s", originalPostDate)
	}

	postDate, err := time.Parse("Jan-02-2006", matches[1])
	if err != nil {
		return nil, fmt.Errorf("\tcould not parse original post date: %v", err)
	}

	description, err := page.Locator(`xpath=//div[contains(@class, 'buysell-container description')]`).TextContent(playwright.LocatorTextContentOptions{Timeout: playwright.Float(1000)})
	if err != nil {
		return nil, fmt.Errorf("\tcould not get description: %v", err)
	}

	restrictions, err := page.Locator(`.buysell-container-right.buysell-restrictions .buysell-container`).TextContent(playwright.LocatorTextContentOptions{
		Timeout: playwright.Float(1000),
	})
	if err != nil {
		return nil, fmt.Errorf("\tcould not get restrictions: %v", err)
	}

	restrictions = strings.Split(restrictions, "Phone Number:")[0]

	details.SellerType = listing.ParseSellerType(parseItemDetail(sellerType, "Seller Type:"))
	details.OriginalPostDate = postDate
	details.Description = description
	details.Restrictions = parseItemDetail(restrictions, "Restrictions:")

	return &details, nil
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
	titleElement := entry.Locator("div.bsitem-title > a")
	title, err := titleElement.TextContent()
	if err != nil {
		fmt.Println("\tcould not get title")
	}
	title = strings.ReplaceAll(title, "\n", "")

	link, err := titleElement.GetAttribute("href")
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
		DetailsLink:   link,
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
