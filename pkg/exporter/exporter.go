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
	db, err := sql.Open("sqlite3", "listings.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// SQLite-compatible schema
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
		hash TEXT UNIQUE,
		first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
		active INTEGER DEFAULT 1
	);

	CREATE TABLE IF NOT EXISTS price_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		listing_hash TEXT,
		price TEXT,
		currency TEXT,
		recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(listing_hash) REFERENCES listings(hash)
	);

	CREATE INDEX IF NOT EXISTS idx_listings_hash ON listings(hash);
	CREATE INDEX IF NOT EXISTS idx_price_history_listing_hash ON price_history(listing_hash);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	// SQLite-compatible upsert
	stmt, err := db.Prepare(`
	INSERT INTO listings (
		title, year, manufacturer, model, price, currency, 
		condition, frame_size, wheel_size, frame_material,
		front_travel, rear_travel, needs_review, url, hash,
		first_seen, last_seen, active
	) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 1)
	ON CONFLICT(hash) DO UPDATE SET 
		last_seen = CURRENT_TIMESTAMP,
		active = 1,
		url = excluded.url,
		price = excluded.price,
		condition = excluded.condition
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// Insert price history
	priceStmt, err := db.Prepare(`
	INSERT INTO price_history (listing_hash, price, currency)
	SELECT ?, ?, ?
	WHERE NOT EXISTS (
		SELECT 1 FROM price_history 
		WHERE listing_hash = ? 
		AND price = ? 
		AND currency = ? 
		AND recorded_at > datetime('now', '-1 day')
	)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare price history statement: %v", err)
	}
	defer priceStmt.Close()

	for _, l := range listings {
		hash := l.ComputeHash()
		_, err = stmt.Exec(
			l.Title, l.Year, l.Manufacturer, l.Model, l.Price,
			l.Currency, l.Condition, l.FrameSize, l.WheelSize,
			l.FrameMaterial, l.FrontTravel, l.RearTravel,
			l.NeedsReview, l.URL, hash,
		)
		if err != nil {
			return fmt.Errorf("failed to insert listing: %v", err)
		}

		// Record price change
		_, err = priceStmt.Exec(hash, l.Price, l.Currency, hash, l.Price, l.Currency)
		if err != nil {
			return fmt.Errorf("failed to record price change: %v", err)
		}
	}

	// Mark old listings as inactive
	_, err = db.Exec(`
	UPDATE listings 
	SET active = 0 
	WHERE datetime(last_seen) < datetime('now', '-7 days')
	`)
	if err != nil {
		return fmt.Errorf("failed to mark old listings as inactive: %v", err)
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
