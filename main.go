package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/playwright-community/playwright-go"
)

type Listing struct {
	title, price, condition, frameSize, wheelSize, frameMaterial, frontTravel, rearTravel string
}

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

	entries, err := page.Locator("tr.bsitem-table").All()
	if err != nil {
		log.Fatalf("could not get entries: %v", err)
	}

	var listings []Listing
	for _, entry := range entries {
		title, err := entry.Locator("div.bsitem-title > a").TextContent()
		if err != nil {
			print("could not get title: %v\n", err)
		}

		condition, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Condition")]]`).InnerText()
		if err != nil {
			print("could not get condition: %v\n", err)
		}

		frameSize, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Frame Size")]]`).TextContent()
		if err != nil {
			print("could not get frame size: %v\n", err)
		}

		wheelSize, err := entry.Locator(`xpath=./descendant:div[b[contains(text(), "Wheel Size")]]`).TextContent()
		if err != nil {
			print("could not get wheel size: %v\n", err)
		}

		frontTravel, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Front Travel")]]`).TextContent()
		if err != nil {
			print("could not get front travel: %v\n", err)
		}

		rearTravel, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Rear Travel")]]`).TextContent()
		if err != nil {
			print("could not get rear travel: %v\n", err)
		}

		material, err := entry.Locator(`xpath=./descendant::div[b[contains(text(), "Material")]]`).TextContent()
		if err != nil {
			log.Fatalf("could not get material: %v\n", err)
		}

		price, err := entry.Locator("td.bsitem-price > b").TextContent()
		if err != nil {
			log.Fatalf("could not get price: %v\n", err)
		}

		listings = append(listings, Listing{
			title: title,
			condition: condition,
			frameSize: frameSize,
			wheelSize: wheelSize,
			frontTravel: frontTravel,
			rearTravel: rearTravel,
			price: price,
			frameMaterial: material,
		})
	}

	var sanitizedListings []Listing
	for _, listing := range listings {
		sanitizedListings = append(sanitizedListings, sanitize(listing))
		fmt.Println(listing.Print())
	}

	err = writeListingsToFile(sanitizedListings, "listings.csv")
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

func (l Listing) Print() string {
	listing := sanitize(l)
	return fmt.Sprintf("Title: %s\nPrice: %s\n\tCondition: %s\n\tFrame Size: %s\n\tWheel Size: %s\n\tFront Travel: %s\n\tRear Travel: %s\n\tFrame Material: %s\n\t\n",
		listing.title, listing.price, listing.condition, listing.frameSize, listing.wheelSize, listing.frontTravel, listing.rearTravel, listing.frameMaterial)
}

// Sanitize will remove spaces and labels from the listing
func sanitize(l Listing) Listing {
	newL := Listing{}

	newL.title = strings.TrimSpace(l.title)
	newL.price = strings.TrimSpace(l.price)
	newL.condition = parseItemDetail(l.condition, "Condition :")
	newL.frameSize = parseItemDetail(l.title, "Frame Size :")
	newL.wheelSize = parseItemDetail(l.wheelSize, "Wheel Size :")
	newL.frameSize = parseItemDetail(l.frameSize, "Frame Size :")
	newL.frontTravel = parseItemDetail(l.frontTravel, "Front Travel :")
	newL.rearTravel = parseItemDetail(l.rearTravel, "Rear Travel :")
	newL.frameMaterial = parseItemDetail(l.frameMaterial, "Material :")

	return newL
}

func parseItemDetail(detail, label string) string {
	split := strings.Split(detail, label)
	if len(split) < 2 {
		return ""
	}

	return strings.TrimSpace(split[1])
}

func writeListingsToFile(listings []Listing, filename string) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Write header
    err = writer.Write([]string{"Title", "Price", "Condition", "Frame Size", "Wheel Size", "Front Travel", "Rear Travel", "Material"})
    if err != nil {
        return err
    }

    // Write data
    for _, listing := range listings {
        err = writer.Write([]string{listing.title, listing.price, listing.condition, listing.frameSize, listing.wheelSize, listing.frontTravel, listing.rearTravel, listing.frameMaterial})
        if err != nil {
            return err
        }
    }

    return nil
}