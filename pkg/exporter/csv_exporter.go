package exporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"pinkbike-scraper/pkg/listing"
)

type CSVExporter struct {
	goodListingsPath    string
	suspectListingsPath string
}

func NewCSVExporter(goodPath, suspectPath string) *CSVExporter {
	return &CSVExporter{
		goodListingsPath:    goodPath,
		suspectListingsPath: suspectPath,
	}
}

func (e *CSVExporter) Close() error {
	return nil
}

func (e *CSVExporter) Export(listings []listing.Listing) error {
	if err := e.writeToFile(listings); err != nil {
		return fmt.Errorf("failed to write to CSV: %w", err)
	}
	return nil
}

func (e *CSVExporter) writeToFile(listings []listing.Listing) error {
	goodFile, err := os.Create(e.goodListingsPath)
	if err != nil {
		return err
	}
	defer goodFile.Close()

	suspectFile, err := os.Create(e.goodListingsPath)
	if err != nil {
		return err
	}
	defer suspectFile.Close()

	goodWriter := csv.NewWriter(goodFile)
	defer goodWriter.Flush()

	suspectWriter := csv.NewWriter(suspectFile)
	defer suspectWriter.Flush()

	csvHeaders := []string{"Title", "Year", "Manufacturer", "Model", "Price", "Currency", "Condition", "Frame Size", "Wheel Size", "Frame Material", "Front Travel", "Rear Travel", "Needs Review"}

	err = goodWriter.Write(csvHeaders)
	if err != nil {
		return err
	}

	err = suspectWriter.Write(csvHeaders)
	if err != nil {
		return err
	}

	for _, l := range listings {
		row := []string{l.Title, l.Year, l.Manufacturer, l.Model, l.Price, l.Currency, l.Condition, l.FrameSize, l.WheelSize, l.FrameMaterial, l.FrontTravel, l.RearTravel, l.NeedsReview}
		if l.NeedsReview != "" {
			err = suspectWriter.Write(row)
			if err != nil {
				return err
			}
			continue
		}

		err = goodWriter.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}
