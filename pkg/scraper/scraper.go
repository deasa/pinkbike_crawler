package scraper

import (
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"
)

type Listing struct {
	Title, Price, Condition, FrameSize, WheelSize, FrameMaterial, FrontTravel, RearTravel string
}

func ScrapePage(page playwright.Page) ([]Listing, error) {
	
	entries, err := page.Locator("tr.bsitem-table").All()
	if err != nil {
		return nil, fmt.Errorf("could not get entries: %v", err)
	}

	var listings []Listing
	for _, entry := range entries {
		listings = append(listings, getListing(entry))
	}

	var sanitizedListings []Listing
	for _, listing := range listings {
		sanitizedListings = append(sanitizedListings, sanitize(listing))
		fmt.Println(listing.Print())
	}

	return sanitizedListings, nil
}

func getListing(entry playwright.Locator) Listing {
	title, err := entry.Locator("div.bsitem-title > a").TextContent()
	if err != nil {
		fmt.Printf("could not get title: %v\n", err)
	}

	condition, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Condition")]]`).InnerText()
	if err != nil {
		fmt.Printf("could not get condition: %v\n", err)
	}

	frameSize, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Frame Size")]]`).TextContent()
	if err != nil {
		fmt.Printf("could not get frame size: %v\n", err)
	}

	wheelSize, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Wheel Size")]]`).TextContent()
	if err != nil {
		fmt.Printf("could not get wheel size: %v\n", err)
	}

	frontTravel, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Front Travel")]]`).TextContent()
	if err != nil {
		fmt.Printf("could not get front travel: %v\n", err)
	}

	rearTravel, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Rear Travel")]]`).TextContent()
	if err != nil {
		fmt.Printf("could not get rear travel: %v\n", err)
	}

	material, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Material")]]`).TextContent()
	if err != nil {
		fmt.Printf("could not get material: %v\n", err)
	}

	price, err := entry.Locator("td.bsitem-price > b").TextContent()
	if err != nil {
		fmt.Printf("could not get price: %v\n", err)
	}

	return Listing{
		Title:         title,
		Condition:     condition,
		FrameSize:     frameSize,
		WheelSize:     wheelSize,
		FrontTravel:   frontTravel,
		RearTravel:    rearTravel,
		Price:         price,
		FrameMaterial: material,
	}
}


func (l Listing) Print() string {
	listing := sanitize(l)
	return fmt.Sprintf("Title: %s\nPrice: %s\n\tCondition: %s\n\tFrame Size: %s\n\tWheel Size: %s\n\tFront Travel: %s\n\tRear Travel: %s\n\tFrame Material: %s\n\t\n",
		listing.Title, listing.Price, listing.Condition, listing.FrameSize, listing.WheelSize, listing.FrontTravel, listing.RearTravel, listing.FrameMaterial)
}

// Sanitize will remove spaces and labels from the listing
func sanitize(l Listing) Listing {
	newL := Listing{}

	newL.Title = strings.TrimSpace(l.Title)
	newL.Price = strings.TrimSpace(l.Price)
	newL.Condition = parseItemDetail(l.Condition, "Condition :")
	newL.FrameSize = parseItemDetail(l.FrameSize, "Frame Size :")
	newL.WheelSize = parseItemDetail(l.WheelSize, "Wheel Size :")
	newL.FrontTravel = parseItemDetail(l.FrontTravel, "Front Travel :")
	newL.RearTravel = parseItemDetail(l.RearTravel, "Rear Travel :")
	newL.FrameMaterial = parseItemDetail(l.FrameMaterial, "Material :")

	return newL
}

func parseItemDetail(detail, label string) string {
	split := strings.Split(detail, label)
	if len(split) < 2 {
		return ""
	}

	return strings.TrimSpace(split[1])
}
