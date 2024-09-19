package exporter

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"pinkbike-scraper/pkg/listing"
)

const (
	spreadsheetID = "16GYqn_Asp6_MhsJNAiMSphtUpJn6P1nNw-BRQG0s5Ik"
)

var (
	csvHeaders = []string{"Title", "Year", "Manufacturer", "Model", "USD Price", "Original Currency", "Condition", "Frame Size", "Wheel Size", "Front Travel", "Rear Travel", "Material", "Reason for Review", "URL"}
)

func WriteListingsToFile(listings []listing.Listing, filenameForGoodListings, filenameForSuspectListings string) error {
	goodFile, err := os.Create(filenameForGoodListings)
	if err != nil {
		return err
	}
	defer goodFile.Close()

	suspectFile, err := os.Create(filenameForSuspectListings)
	if err != nil {
		return err
	}
	defer suspectFile.Close()

	goodWriter := csv.NewWriter(goodFile)
	defer goodWriter.Flush()

	suspectWriter := csv.NewWriter(suspectFile)
	defer suspectWriter.Flush()

	err = goodWriter.Write(csvHeaders)
	if err != nil {
		return err
	}

	err = suspectWriter.Write(csvHeaders)
	if err != nil {
		return err
	}

	for _, l := range listings {
		row := []string{l.Title, l.Year, l.Manufacturer, l.Model, l.Price, l.Currency, l.Condition, l.FrameSize, l.WheelSize, l.FrontTravel, l.RearTravel, l.FrameMaterial, l.NeedsReview, l.URL}
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

func ExportToGoogleSheets(listings []listing.Listing) error {
	// Create a new Google Sheets service client
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("pinkbike-exporter-8bc8e681ffa1.json"))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	var values [][]interface{}
	for _, l := range listings {
		values = append(values, []interface{}{l.Title, l.Year, l.Manufacturer, l.Model, l.Price, l.URL, l.Condition, l.FrameSize, l.WheelSize, l.FrontTravel, l.RearTravel, l.FrameMaterial, l.NeedsReview, l.Currency})
	}

	// Create the value range object
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	// Append the data to the sheet
	appendRange := "Sheet1"
	_, err = srv.Spreadsheets.Values.Append(spreadsheetID, appendRange, valueRange).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		return fmt.Errorf("Unable to append data to sheet: %v", err)
	}

	err = SendDeDuplicateRequestToGoogleSheets(srv)
	if err != nil {
		return err
	}

	return nil
}

// SendDeDuplicateRequestToGoogleSheets removes duplicate rows from the Google Sheets document
// NOTE: Only the first match is kept! This means that when a listing's price changes, the old listing and old price will be kept.
func SendDeDuplicateRequestToGoogleSheets(srv *sheets.Service) error {
	// Remove duplicates from the sheet, considering only specific columns
	deleteDuplicatesRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				DeleteDuplicates: &sheets.DeleteDuplicatesRequest{
					Range: &sheets.GridRange{
						SheetId:          0,
						StartRowIndex:    0,
						StartColumnIndex: 0,
						EndColumnIndex:   12, // Include columns 0 to 11 (Title to FrameMaterial)
					},
					ComparisonColumns: []*sheets.DimensionRange{
						{
							SheetId:    0,
							Dimension:  "COLUMNS",
							StartIndex: 0, // Title
							EndIndex:   3, // Model
						},
						{
							SheetId:    0,
							Dimension:  "COLUMNS",
							StartIndex: 6,  // Condition
							EndIndex:   11, // FrameMaterial
						},
					},
				},
			},
		},
	}

	_, err := srv.Spreadsheets.BatchUpdate(spreadsheetID, deleteDuplicatesRequest).Do()
	if err != nil {
		return fmt.Errorf("Unable to remove duplicates from sheet: %v", err)
	}

	return nil
}

func ExportToListingsDB(listings []listing.Listing) error {
	// Open or create the SQLite database file
	db, err := sql.Open("sqlite3", "listings.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create the listings table if it doesn't exist, with a unique index on all columns
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS listings (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT,
        year TEXT,
        manufacturer TEXT,
        model TEXT,
        price TEXT,
        currency TEXT,
        condition TEXT,
        frame_size TEXT,
        wheel_size TEXT,
        front_travel TEXT,
        rear_travel TEXT,
        frame_material TEXT,
        needs_review TEXT,
        url TEXT,
        UNIQUE(title, year, manufacturer, model, currency, condition, 
               frame_size, wheel_size, front_travel, rear_travel, frame_material)
    );
    `
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	// Insert listings into the database, replacing duplicates
	insertSQL := `
    REPLACE INTO listings (
        title, year, manufacturer, model, price, currency, condition, 
        frame_size, wheel_size, front_travel, rear_travel, frame_material,
        needs_review, url
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	for _, l := range listings {
		_, err = stmt.Exec(
			l.Title, l.Year, l.Manufacturer, l.Model, l.Price, l.Currency, l.Condition,
			l.FrameSize, l.WheelSize, l.FrontTravel, l.RearTravel, l.FrameMaterial,
			l.NeedsReview, l.URL,
		)
		if err != nil {
			return fmt.Errorf("failed to insert listing: %v", err)
		}
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
