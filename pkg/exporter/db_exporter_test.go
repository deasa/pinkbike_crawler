package exporter

import (
	"database/sql"
	"fmt"
	"pinkbike-scraper/pkg/db"
	"pinkbike-scraper/pkg/listing"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (*db.DBWorker, func()) {
	tempDBPath := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())

	sqlDB, err := sql.Open("sqlite3", tempDBPath)
	if err != nil {
		t.Fatalf("Failed to open in-memory SQLite DB: %v", err)
	}

	// Create tables
	_, err = sqlDB.Exec(`
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
        description TEXT,
        restrictions TEXT,
        seller_type TEXT,
        original_post_date DATETIME,
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
    `)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	dbWorker := &db.DBWorker{
		DB: sqlDB,
	}

	cleanup := func() {
		sqlDB.Close()
	}

	return dbWorker, cleanup
}

func createTestListing() listing.Listing {
	return listing.Listing{
		Title:         "Test Bike 2023",
		Year:          "2023",
		Manufacturer:  "TestBike",
		Model:         "SuperTest",
		Price:         "2000",
		Currency:      "USD",
		Condition:     "Good",
		FrameSize:     "Large",
		WheelSize:     "29",
		FrameMaterial: "Carbon",
		FrontTravel:   "150",
		RearTravel:    "140",
		NeedsReview:   "",
		URL:           "https://example.com/bike",
		Hash:          "testhash123",
		Details: listing.ListingDetails{
			Description:      "A test bike description",
			Restrictions:     "No shipping",
			SellerType:       listing.Private,
			OriginalPostDate: time.Now(),
		},
	}
}

func TestNewDBExporter(t *testing.T) {
	dbWorker, cleanup := setupTestDB(t)
	defer cleanup()

	exporter := NewDBExporter(dbWorker)
	assert.NotNil(t, exporter)
	assert.Equal(t, dbWorker, exporter.dbWorker)
}

func TestExport(t *testing.T) {
	dbWorker, cleanup := setupTestDB(t)
	defer cleanup()

	exporter := NewDBExporter(dbWorker)
	testListing := createTestListing()

	err := exporter.Export([]listing.Listing{testListing})
	assert.NoError(t, err)

	// Verify the listing was inserted correctly
	var count int
	err = dbWorker.DB.QueryRow("SELECT COUNT(*) FROM listings WHERE hash = ?", testListing.Hash).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify price history was recorded
	err = dbWorker.DB.QueryRow("SELECT COUNT(*) FROM price_history WHERE listing_hash = ?", testListing.Hash).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestClose(t *testing.T) {
	dbWorker, cleanup := setupTestDB(t)
	defer cleanup()

	exporter := NewDBExporter(dbWorker)
	err := exporter.Close()
	assert.NoError(t, err)
}

func TestExportErrorHandling(t *testing.T) {
	// Create a DB that will be closed to simulate error
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory SQLite DB: %v", err)
	}

	// Close the DB to force errors
	sqlDB.Close()

	dbWorker := &db.DBWorker{
		DB: sqlDB,
	}

	exporter := NewDBExporter(dbWorker)
	testListing := createTestListing()

	err = exporter.Export([]listing.Listing{testListing})
	assert.Error(t, err)
}

func TestMarkInactiveListings(t *testing.T) {
	dbWorker, cleanup := setupTestDB(t)
	defer cleanup()

	// Insert a test listing with an old last_seen date
	_, err := dbWorker.DB.Exec(`
		INSERT INTO listings (
			title, year, manufacturer, model, price, currency, 
			condition, frame_size, wheel_size, frame_material,
			front_travel, rear_travel, hash, url, 
			last_seen, active
		) VALUES (
			'Old Bike', '2020', 'TestBike', 'OldModel', '1000', 'USD',
			'Used', 'Medium', '27.5', 'Aluminum',
			'140', '130', 'oldhash123', 'https://example.com/old',
			datetime('now', '-10 days'), 1
		)
	`)
	assert.NoError(t, err)

	exporter := NewDBExporter(dbWorker)

	// Start a transaction manually since we're testing only markInactiveListings
	tx, err := dbWorker.DB.Begin()
	assert.NoError(t, err)

	err = exporter.markInactiveListings(tx)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Verify the listing was marked as inactive
	var active int
	err = dbWorker.DB.QueryRow("SELECT active FROM listings WHERE hash = 'oldhash123'").Scan(&active)
	assert.NoError(t, err)
	assert.Equal(t, 0, active)
}

func TestRecordPriceHistory(t *testing.T) {
	dbWorker, cleanup := setupTestDB(t)
	defer cleanup()

	exporter := NewDBExporter(dbWorker)
	testListing := createTestListing()

	// Start a transaction
	tx, err := dbWorker.DB.Begin()
	assert.NoError(t, err)

	// Record price history
	err = exporter.recordPriceHistory(tx, testListing)
	assert.NoError(t, err)

	// Record the same price again (should not create another entry due to the NOT EXISTS clause)
	err = exporter.recordPriceHistory(tx, testListing)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Verify only one price history entry was created
	var count int
	err = dbWorker.DB.QueryRow("SELECT COUNT(*) FROM price_history WHERE listing_hash = ?", testListing.Hash).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	// Change the price and record it
	testListing.Price = "2100"
	tx, err = dbWorker.DB.Begin()
	assert.NoError(t, err)

	err = exporter.recordPriceHistory(tx, testListing)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Verify a second price history entry was created
	err = dbWorker.DB.QueryRow("SELECT COUNT(*) FROM price_history WHERE listing_hash = ?", testListing.Hash).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestExportListings(t *testing.T) {
	dbWorker, cleanup := setupTestDB(t)
	defer cleanup()

	exporter := NewDBExporter(dbWorker)
	testListing1 := createTestListing()
	testListing2 := createTestListing()
	testListing2.Hash = "another-hash-123"
	testListing2.Title = "Another Test Bike"

	// Start a transaction
	tx, err := dbWorker.DB.Begin()
	assert.NoError(t, err)

	// Export multiple listings
	err = exporter.exportListings(tx, []listing.Listing{testListing1, testListing2})
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Verify both listings were inserted
	var count int
	err = dbWorker.DB.QueryRow("SELECT COUNT(*) FROM listings").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}
