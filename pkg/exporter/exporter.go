package exporter

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"pinkbike-scraper/pkg/listing"
)


func WriteListingsToFile(listings []listing.Listing, filename string) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Write header
    err = writer.Write([]string{"Title", "Year", "Price", "Currency", "Condition", "Frame Size", "Wheel Size", "Front Travel", "Rear Travel", "Material"})
    if err != nil {
        return err
    }

    // Write data
	for _, listing := range listings {
		err = writer.Write([]string{listing.Title, listing.Year,listing.Manufacturer, listing.Price, listing.Currency, listing.Condition, listing.FrameSize, listing.WheelSize, listing.FrontTravel, listing.RearTravel, listing.FrameMaterial})
		if err != nil {
			return err
		}
	}

    return nil
}

func ExportToGoogleSheets(listings []listing.Listing) error {
	// Create a new Google Sheets service client
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("pinkbike-exporter-8bc8e681ffa1.json"))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// createSheetAndShare(ctx, srv, "Pinkbike Crawler Data", "bgeorgeashton@gmail.com", "pinkbike-exporter-8bc8e681ffa1.json")

	// Define the spreadsheet ID and range
	spreadsheetID := "16GYqn_Asp6_MhsJNAiMSphtUpJn6P1nNw-BRQG0s5Ik"
	sheetName := "Sheet1"
	writeRange := sheetName + "!A1:ZZ"

	// Prepare the data to be written to the sheet
	var values [][]interface{}
	values = append(values, []interface{}{"Title", "Year", "Price", "Currency", "Condition", "Frame Size", "Wheel Size", "Front Travel", "Rear Travel", "Material"})
	for _, listing := range listings {
		values = append(values, []interface{}{listing.Title, listing.Year, listing.Manufacturer, listing.Price, listing.Currency, listing.Condition, listing.FrameSize, listing.WheelSize, listing.FrontTravel, listing.RearTravel, listing.FrameMaterial})
	}

	// Create the value range object
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	// Write the data to the sheet
	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, writeRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("Unable to write data to sheet: %v", err)
	}

	return nil
}

func createSheetAndShare(ctx context.Context, srv *sheets.Service, title, email, credentialFile string) {
	sheet, err := srv.Spreadsheets.Create(&sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
	}).Do()
	if err != nil {
		log.Fatalf("Unable to create spreadsheet: %v", err)
	}

	fmt.Printf("Created new spreadsheet: %s\n", sheet.SpreadsheetUrl)

	driveService, err := drive.NewService(ctx, option.WithCredentialsFile(credentialFile))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	_, err = driveService.Permissions.Create(sheet.SpreadsheetId, &drive.Permission{
		Type:         "user",
		Role:         "writer",
		EmailAddress: email,
	}).Do()
	if err != nil {
		log.Fatalf("Unable to share spreadsheet: %v", err)
	}
}