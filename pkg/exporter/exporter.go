package exporter

import (
	"encoding/csv"
	"os"
	"pinkbike-scraper/pkg/scraper"
)


func WriteListingsToFile(listings []scraper.Listing, filename string) error {
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
		err = writer.Write([]string{listing.Title, listing.Price, listing.Condition, listing.FrameSize, listing.WheelSize, listing.FrontTravel, listing.RearTravel, listing.FrameMaterial})
		if err != nil {
			return err
		}
	}

    return nil
}