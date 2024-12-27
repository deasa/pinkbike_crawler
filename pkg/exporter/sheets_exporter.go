package exporter

import (
	"context"
	"fmt"
	"pinkbike-scraper/pkg/listing"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetsExporter struct {
	service       *sheets.Service
	spreadsheetID string
}

func NewSheetsExporter(credentialsFile, spreadsheetID string) (*SheetsExporter, error) {
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	return &SheetsExporter{
		service:       srv,
		spreadsheetID: spreadsheetID,
	}, nil
}

func (e *SheetsExporter) Close() error {
	return nil
}

func (e *SheetsExporter) Export(listings []listing.Listing) error {
	if err := e.appendToSheet(listings); err != nil {
		return fmt.Errorf("failed to export to sheets: %w", err)
	}
	return e.removeDuplicates()
}

func (e *SheetsExporter) appendToSheet(listings []listing.Listing) error {
	// Create a new Google Sheets service client
	ctx := context.Background()
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("pinkbike-exporter-8bc8e681ffa1.json"))
	if err != nil {
		return fmt.Errorf("Unable to retrieve Sheets client: %v", err)
	}

	var values [][]interface{}
	for _, l := range listings {
		values = append(values, []interface{}{l.Title, l.Year, l.Manufacturer, l.Model, l.Price, l.Condition, l.
			FrameSize, l.WheelSize, l.FrontTravel, l.RearTravel, l.FrameMaterial, l.NeedsReview, l.Currency, l.URL})
	}

	// Create the value range object
	valueRange := &sheets.ValueRange{
		Values: values,
	}

	// Append the data to the sheet
	appendRange := "Sheet1"
	_, err = srv.Spreadsheets.Values.Append(e.spreadsheetID, appendRange, valueRange).ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		return fmt.Errorf("Unable to append data to sheet: %v", err)
	}

	return nil
}

// SendDeDuplicateRequestToGoogleSheets removes duplicate rows from the Google Sheets document
// NOTE: Only the first match is kept! This means that when a listing's price changes, the old listing and old price will be kept.
func (e *SheetsExporter) removeDuplicates() error {
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

	_, err := e.service.Spreadsheets.BatchUpdate(e.spreadsheetID, deleteDuplicatesRequest).Do()
	if err != nil {
		return fmt.Errorf("Unable to remove duplicates from sheet: %v", err)
	}

	return nil
}

func createSheetAndShare(ctx context.Context, srv *sheets.Service, title, email, credentialFile string) error {
	sheet, err := srv.Spreadsheets.Create(&sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: title,
		},
	}).Do()
	if err != nil {
		return fmt.Errorf("Unable to create spreadsheet: %v", err)
	}

	fmt.Printf("Created new spreadsheet: %s\n", sheet.SpreadsheetUrl)

	driveService, err := drive.NewService(ctx, option.WithCredentialsFile(credentialFile))
	if err != nil {
		return fmt.Errorf("Unable to retrieve Drive client: %v", err)
	}

	_, err = driveService.Permissions.Create(sheet.SpreadsheetId, &drive.Permission{
		Type:         "user",
		Role:         "writer",
		EmailAddress: email,
	}).Do()
	if err != nil {
		return fmt.Errorf("Unable to share spreadsheet: %v", err)
	}

	return nil
}
