package exporter

import (
	"database/sql"
	"errors"
	"pinkbike-scraper/pkg/db"
	"pinkbike-scraper/pkg/listing"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupDBExporterTest sets up sqlmock and creates a DBExporter instance
func setupDBExporterTest(t *testing.T) (*DBExporter, sqlmock.Sqlmock, *sql.DB) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")

	dbWorker := &db.DBWorker{
		DB: mockDB,
	}

	exporter := NewDBExporter(dbWorker)
	return exporter, mock, mockDB
}

// createMockListings creates sample listing data for tests
func createMockListings() []listing.Listing {
	return []listing.Listing{
		{
			Title:         "Test Bike 1",
			Year:          "2022",
			Manufacturer:  "TestBrand",
			Model:         "TestModel",
			Price:         "1000",
			Currency:      "USD",
			Condition:     "Good",
			FrameSize:     "M",
			WheelSize:     "29",
			FrameMaterial: "Carbon",
			FrontTravel:   "150",
			RearTravel:    "140",
			NeedsReview:   "",
			URL:           "https://example.com/bike1",
			Hash:          "hash1",
			Details: listing.ListingDetails{
				Description:      "Test description 1",
				Restrictions:     "No shipping",
				SellerType:       listing.Private,
				OriginalPostDate: time.Now().Add(-24 * time.Hour),
			},
		},
		{
			Title:         "Test Bike 2",
			Year:          "2023",
			Manufacturer:  "AnotherBrand",
			Model:         "AnotherModel",
			Price:         "2000",
			Currency:      "USD",
			Condition:     "Excellent",
			FrameSize:     "L",
			WheelSize:     "27.5",
			FrameMaterial: "Aluminum",
			FrontTravel:   "160",
			RearTravel:    "150",
			NeedsReview:   "",
			URL:           "https://example.com/bike2",
			Hash:          "hash2",
			Details: listing.ListingDetails{
				Description:      "Test description 2",
				Restrictions:     "Local pickup only",
				SellerType:       listing.Business,
				OriginalPostDate: time.Now().Add(-48 * time.Hour),
			},
		},
	}
}

// expectBeginTransaction sets up expectations for beginning a transaction
func expectBeginTransaction(mock sqlmock.Sqlmock) *sqlmock.ExpectedBegin {
	return mock.ExpectBegin()
}

// expectCommitTransaction sets up expectations for committing a transaction
func expectCommitTransaction(mock sqlmock.Sqlmock) *sqlmock.ExpectedCommit {
	return mock.ExpectCommit()
}

// expectRollbackTransaction sets up expectations for rolling back a transaction
func expectRollbackTransaction(mock sqlmock.Sqlmock) *sqlmock.ExpectedRollback {
	return mock.ExpectRollback()
}

// TestNewDBExporter tests the constructor function
func TestNewDBExporter(t *testing.T) {
	// Create a mock DBWorker
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err, "Failed to create mock database")

	dbWorker := &db.DBWorker{
		DB: mockDB,
	}

	// Call NewDBExporter
	exporter := NewDBExporter(dbWorker)

	// Assert that the returned DBExporter has the correct DBWorker
	assert.Equal(t, dbWorker, exporter.dbWorker, "DBExporter should have the provided DBWorker")
}

// TestClose tests the Close method
func TestClose(t *testing.T) {
	// Set up mock
	exporter, mock, mockDB := setupDBExporterTest(t)
	defer mockDB.Close()

	// Expect Close to be called
	mock.ExpectClose()

	// Call Close
	err := exporter.Close()

	// Verify expectations
	assert.NoError(t, err, "Close should not return an error")
	assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
}

// TestExport tests the Export method
func TestExport(t *testing.T) {
	t.Run("successful export", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		listings := createMockListings()

		// Expect transaction to begin
		expectBeginTransaction(mock)

		// Expect prepare statement for exportListings
		mock.ExpectPrepare("INSERT INTO listings").
			ExpectExec().
			WithArgs(
				listings[0].Title, listings[0].Year, listings[0].Manufacturer, listings[0].Model,
				listings[0].Price, listings[0].Currency, listings[0].Condition, listings[0].FrameSize,
				listings[0].WheelSize, listings[0].FrameMaterial, listings[0].FrontTravel,
				listings[0].RearTravel, listings[0].NeedsReview, listings[0].URL, listings[0].Hash,
				listings[0].Details.Description, listings[0].Details.Restrictions,
				listings[0].Details.SellerType, listings[0].Details.OriginalPostDate,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect price history to be recorded for first listing
		mock.ExpectExec("INSERT INTO price_history").
			WithArgs(
				listings[0].Hash, listings[0].Price, listings[0].Currency,
				listings[0].Hash, listings[0].Price,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect prepare statement for second listing
		mock.ExpectExec("INSERT INTO listings").
			WithArgs(
				listings[1].Title, listings[1].Year, listings[1].Manufacturer, listings[1].Model,
				listings[1].Price, listings[1].Currency, listings[1].Condition, listings[1].FrameSize,
				listings[1].WheelSize, listings[1].FrameMaterial, listings[1].FrontTravel,
				listings[1].RearTravel, listings[1].NeedsReview, listings[1].URL, listings[1].Hash,
				listings[1].Details.Description, listings[1].Details.Restrictions,
				listings[1].Details.SellerType, listings[1].Details.OriginalPostDate,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect price history to be recorded for second listing
		mock.ExpectExec("INSERT INTO price_history").
			WithArgs(
				listings[1].Hash, listings[1].Price, listings[1].Currency,
				listings[1].Hash, listings[1].Price,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect markInactiveListings to be called
		mock.ExpectExec("UPDATE listings").
			WillReturnResult(sqlmock.NewResult(0, 5)) // 5 rows affected

		// Expect transaction to be committed
		expectCommitTransaction(mock)

		// Call Export
		err := exporter.Export(listings)

		// Verify expectations
		assert.NoError(t, err, "Export should not return an error")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})

	t.Run("failed to begin transaction", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		listings := createMockListings()

		// Expect transaction to begin but fail
		mock.ExpectBegin().WillReturnError(errors.New("begin transaction error"))

		// Call Export
		err := exporter.Export(listings)

		// Verify expectations
		assert.Error(t, err, "Export should return an error")
		assert.Contains(t, err.Error(), "failed to begin transaction", "Error message should mention transaction failure")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})

	t.Run("failed to export listings", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		listings := createMockListings()

		// Expect transaction to begin
		expectBeginTransaction(mock)

		// Expect prepare statement to fail
		mock.ExpectPrepare("INSERT INTO listings").WillReturnError(errors.New("prepare statement error"))

		// Expect transaction to be rolled back
		expectRollbackTransaction(mock)

		// Call Export
		err := exporter.Export(listings)

		// Verify expectations
		assert.Error(t, err, "Export should return an error")
		assert.Contains(t, err.Error(), "failed to prepare statement", "Error message should mention prepare statement failure")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})

	t.Run("failed to mark inactive listings", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		listings := createMockListings()

		// Expect transaction to begin
		expectBeginTransaction(mock)

		// Expect prepare statement for exportListings
		mock.ExpectPrepare("INSERT INTO listings").
			ExpectExec().
			WithArgs(
				listings[0].Title, listings[0].Year, listings[0].Manufacturer, listings[0].Model,
				listings[0].Price, listings[0].Currency, listings[0].Condition, listings[0].FrameSize,
				listings[0].WheelSize, listings[0].FrameMaterial, listings[0].FrontTravel,
				listings[0].RearTravel, listings[0].NeedsReview, listings[0].URL, listings[0].Hash,
				listings[0].Details.Description, listings[0].Details.Restrictions,
				listings[0].Details.SellerType, listings[0].Details.OriginalPostDate,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect price history to be recorded for first listing
		mock.ExpectExec("INSERT INTO price_history").
			WithArgs(
				listings[0].Hash, listings[0].Price, listings[0].Currency,
				listings[0].Hash, listings[0].Price,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect prepare statement for second listing
		mock.ExpectExec("INSERT INTO listings").
			WithArgs(
				listings[1].Title, listings[1].Year, listings[1].Manufacturer, listings[1].Model,
				listings[1].Price, listings[1].Currency, listings[1].Condition, listings[1].FrameSize,
				listings[1].WheelSize, listings[1].FrameMaterial, listings[1].FrontTravel,
				listings[1].RearTravel, listings[1].NeedsReview, listings[1].URL, listings[1].Hash,
				listings[1].Details.Description, listings[1].Details.Restrictions,
				listings[1].Details.SellerType, listings[1].Details.OriginalPostDate,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect price history to be recorded for second listing
		mock.ExpectExec("INSERT INTO price_history").
			WithArgs(
				listings[1].Hash, listings[1].Price, listings[1].Currency,
				listings[1].Hash, listings[1].Price,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect markInactiveListings to fail
		mock.ExpectExec("UPDATE listings").
			WillReturnError(errors.New("update error"))

		// Expect transaction to be rolled back
		expectRollbackTransaction(mock)

		// Call Export
		err := exporter.Export(listings)

		// Verify expectations
		assert.Error(t, err, "Export should return an error")
		assert.Contains(t, err.Error(), "failed to mark inactive listings", "Error message should mention marking inactive listings failure")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})
}

// TestMarkInactiveListings tests the markInactiveListings method
func TestMarkInactiveListings(t *testing.T) {
	t.Run("successful marking", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		// Expect transaction to begin
		mock.ExpectBegin()

		// Get the transaction
		tx, err := mockDB.Begin()
		require.NoError(t, err, "Failed to begin transaction")

		// Expect markInactiveListings to be called
		mock.ExpectExec("UPDATE listings").
			WillReturnResult(sqlmock.NewResult(0, 5)) // 5 rows affected

		// Call markInactiveListings
		err = exporter.markInactiveListings(tx)

		// Verify expectations
		assert.NoError(t, err, "markInactiveListings should not return an error")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})

	t.Run("failed marking", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		// Expect transaction to begin
		mock.ExpectBegin()

		// Get the transaction
		tx, err := mockDB.Begin()
		require.NoError(t, err, "Failed to begin transaction")

		// Expect markInactiveListings to fail
		mock.ExpectExec("UPDATE listings").
			WillReturnError(errors.New("update error"))

		// Call markInactiveListings
		err = exporter.markInactiveListings(tx)

		// Verify expectations
		assert.Error(t, err, "markInactiveListings should return an error")
		assert.Contains(t, err.Error(), "failed to mark inactive listings", "Error message should mention marking inactive listings failure")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})
}

// TestExportListings tests the exportListings method
func TestExportListings(t *testing.T) {
	t.Run("successful export of multiple listings", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		listings := createMockListings()

		// Expect transaction to begin
		mock.ExpectBegin()

		// Get the transaction
		tx, err := mockDB.Begin()
		require.NoError(t, err, "Failed to begin transaction")

		// Expect prepare statement for exportListings
		mock.ExpectPrepare("INSERT INTO listings").
			ExpectExec().
			WithArgs(
				listings[0].Title, listings[0].Year, listings[0].Manufacturer, listings[0].Model,
				listings[0].Price, listings[0].Currency, listings[0].Condition, listings[0].FrameSize,
				listings[0].WheelSize, listings[0].FrameMaterial, listings[0].FrontTravel,
				listings[0].RearTravel, listings[0].NeedsReview, listings[0].URL, listings[0].Hash,
				listings[0].Details.Description, listings[0].Details.Restrictions,
				listings[0].Details.SellerType, listings[0].Details.OriginalPostDate,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect price history to be recorded for first listing
		mock.ExpectExec("INSERT INTO price_history").
			WithArgs(
				listings[0].Hash, listings[0].Price, listings[0].Currency,
				listings[0].Hash, listings[0].Price,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect prepare statement for second listing
		mock.ExpectExec("INSERT INTO listings").
			WithArgs(
				listings[1].Title, listings[1].Year, listings[1].Manufacturer, listings[1].Model,
				listings[1].Price, listings[1].Currency, listings[1].Condition, listings[1].FrameSize,
				listings[1].WheelSize, listings[1].FrameMaterial, listings[1].FrontTravel,
				listings[1].RearTravel, listings[1].NeedsReview, listings[1].URL, listings[1].Hash,
				listings[1].Details.Description, listings[1].Details.Restrictions,
				listings[1].Details.SellerType, listings[1].Details.OriginalPostDate,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect price history to be recorded for second listing
		mock.ExpectExec("INSERT INTO price_history").
			WithArgs(
				listings[1].Hash, listings[1].Price, listings[1].Currency,
				listings[1].Hash, listings[1].Price,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Call exportListings
		err = exporter.exportListings(tx, listings)

		// Verify expectations
		assert.NoError(t, err, "exportListings should not return an error")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})

	t.Run("failed to prepare statement", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		listings := createMockListings()

		// Expect transaction to begin
		mock.ExpectBegin()

		// Get the transaction
		tx, err := mockDB.Begin()
		require.NoError(t, err, "Failed to begin transaction")

		// Expect prepare statement to fail
		mock.ExpectPrepare("INSERT INTO listings").WillReturnError(errors.New("prepare statement error"))

		// Call exportListings
		err = exporter.exportListings(tx, listings)

		// Verify expectations
		assert.Error(t, err, "exportListings should return an error")
		assert.Contains(t, err.Error(), "failed to prepare statement", "Error message should mention prepare statement failure")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})

	t.Run("failed to export a listing", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		listings := createMockListings()

		// Expect transaction to begin
		mock.ExpectBegin()

		// Get the transaction
		tx, err := mockDB.Begin()
		require.NoError(t, err, "Failed to begin transaction")

		// Expect prepare statement for exportListings
		mock.ExpectPrepare("INSERT INTO listings").
			ExpectExec().
			WithArgs(
				listings[0].Title, listings[0].Year, listings[0].Manufacturer, listings[0].Model,
				listings[0].Price, listings[0].Currency, listings[0].Condition, listings[0].FrameSize,
				listings[0].WheelSize, listings[0].FrameMaterial, listings[0].FrontTravel,
				listings[0].RearTravel, listings[0].NeedsReview, listings[0].URL, listings[0].Hash,
				listings[0].Details.Description, listings[0].Details.Restrictions,
				listings[0].Details.SellerType, listings[0].Details.OriginalPostDate,
			).
			WillReturnError(errors.New("exec error"))

		// Call exportListings
		err = exporter.exportListings(tx, listings)

		// Verify expectations
		assert.Error(t, err, "exportListings should return an error")
		assert.Contains(t, err.Error(), "failed to insert listing", "Error message should mention insert listing failure")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})
}

// TestExportListing tests the exportListing method
func TestExportListing(t *testing.T) {
	t.Run("successful export of a single listing", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		listing := createMockListings()[0]

		// Expect transaction to begin
		mock.ExpectBegin()

		// Get the transaction
		tx, err := mockDB.Begin()
		require.NoError(t, err, "Failed to begin transaction")

		// Create a prepared statement
		mock.ExpectPrepare("INSERT INTO listings")
		stmt, err := tx.Prepare("INSERT INTO listings")
		require.NoError(t, err, "Failed to prepare statement")

		// Expect exec to be called on the statement
		mock.ExpectExec("INSERT INTO listings").
			WithArgs(
				listing.Title, listing.Year, listing.Manufacturer, listing.Model,
				listing.Price, listing.Currency, listing.Condition, listing.FrameSize,
				listing.WheelSize, listing.FrameMaterial, listing.FrontTravel,
				listing.RearTravel, listing.NeedsReview, listing.URL, listing.Hash,
				listing.Details.Description, listing.Details.Restrictions,
				listing.Details.SellerType, listing.Details.OriginalPostDate,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect price history to be recorded
		mock.ExpectExec("INSERT INTO price_history").
			WithArgs(
				listing.Hash, listing.Price, listing.Currency,
				listing.Hash, listing.Price,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Call exportListing
		err = exporter.exportListing(stmt, tx, listing)

		// Verify expectations
		assert.NoError(t, err, "exportListing should not return an error")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})

	t.Run("failed to execute statement", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		listing := createMockListings()[0]

		// Expect transaction to begin
		mock.ExpectBegin()

		// Get the transaction
		tx, err := mockDB.Begin()
		require.NoError(t, err, "Failed to begin transaction")

		// Create a prepared statement
		mock.ExpectPrepare("INSERT INTO listings")
		stmt, err := tx.Prepare("INSERT INTO listings")
		require.NoError(t, err, "Failed to prepare statement")

		// Expect exec to fail
		mock.ExpectExec("INSERT INTO listings").
			WithArgs(
				listing.Title, listing.Year, listing.Manufacturer, listing.Model,
				listing.Price, listing.Currency, listing.Condition, listing.FrameSize,
				listing.WheelSize, listing.FrameMaterial, listing.FrontTravel,
				listing.RearTravel, listing.NeedsReview, listing.URL, listing.Hash,
				listing.Details.Description, listing.Details.Restrictions,
				listing.Details.SellerType, listing.Details.OriginalPostDate,
			).
			WillReturnError(errors.New("exec error"))

		// Call exportListing
		err = exporter.exportListing(stmt, tx, listing)

		// Verify expectations
		assert.Error(t, err, "exportListing should return an error")
		assert.Contains(t, err.Error(), "failed to insert listing", "Error message should mention insert listing failure")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})

	t.Run("failed to record price history", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		listing := createMockListings()[0]

		// Expect transaction to begin
		mock.ExpectBegin()

		// Get the transaction
		tx, err := mockDB.Begin()
		require.NoError(t, err, "Failed to begin transaction")

		// Create a prepared statement
		mock.ExpectPrepare("INSERT INTO listings")
		stmt, err := tx.Prepare("INSERT INTO listings")
		require.NoError(t, err, "Failed to prepare statement")

		// Expect exec to be called on the statement
		mock.ExpectExec("INSERT INTO listings").
			WithArgs(
				listing.Title, listing.Year, listing.Manufacturer, listing.Model,
				listing.Price, listing.Currency, listing.Condition, listing.FrameSize,
				listing.WheelSize, listing.FrameMaterial, listing.FrontTravel,
				listing.RearTravel, listing.NeedsReview, listing.URL, listing.Hash,
				listing.Details.Description, listing.Details.Restrictions,
				listing.Details.SellerType, listing.Details.OriginalPostDate,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Expect price history recording to fail
		mock.ExpectExec("INSERT INTO price_history").
			WithArgs(
				listing.Hash, listing.Price, listing.Currency,
				listing.Hash, listing.Price,
			).
			WillReturnError(errors.New("insert error"))

		// Call exportListing
		err = exporter.exportListing(stmt, tx, listing)

		// Verify expectations
		assert.Error(t, err, "exportListing should return an error")
		assert.Contains(t, err.Error(), "failed to record price history", "Error message should mention recording price history failure")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})
}

// TestRecordPriceHistory tests the recordPriceHistory method
func TestRecordPriceHistory(t *testing.T) {
	t.Run("successful recording", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		// Expect transaction to begin
		mock.ExpectBegin()

		// Create a transaction
		tx, err := mockDB.Begin()
		require.NoError(t, err, "Failed to begin transaction")

		listing := createMockListings()[0]

		// Expect recordPriceHistory to be called
		mock.ExpectExec("INSERT INTO price_history").
			WithArgs(
				listing.Hash, listing.Price, listing.Currency,
				listing.Hash, listing.Price,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Call recordPriceHistory
		err = exporter.recordPriceHistory(tx, listing)

		// Verify expectations
		assert.NoError(t, err, "recordPriceHistory should not return an error")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})

	t.Run("failed recording", func(t *testing.T) {
		// Set up mock
		exporter, mock, mockDB := setupDBExporterTest(t)
		defer mockDB.Close()

		// Expect transaction to begin
		mock.ExpectBegin()

		// Create a transaction
		tx, err := mockDB.Begin()
		require.NoError(t, err, "Failed to begin transaction")

		listing := createMockListings()[0]

		// Expect recordPriceHistory to fail
		mock.ExpectExec("INSERT INTO price_history").
			WithArgs(
				listing.Hash, listing.Price, listing.Currency,
				listing.Hash, listing.Price,
			).
			WillReturnError(errors.New("insert error"))

		// Call recordPriceHistory
		err = exporter.recordPriceHistory(tx, listing)

		// Verify expectations
		assert.Error(t, err, "recordPriceHistory should return an error")
		assert.Contains(t, err.Error(), "failed to record price history", "Error message should mention recording price history failure")
		assert.NoError(t, mock.ExpectationsWereMet(), "All expectations should be met")
	})
}
