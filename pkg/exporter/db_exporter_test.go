package exporter

import (
	"database/sql"
	"os"
	"path/filepath"
	"pinkbike-scraper/pkg/db"
	"pinkbike-scraper/pkg/listing"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDBExporter(t *testing.T) {
	// Create a mock DB worker
	dbWorker := &db.DBWorker{DB: &sql.DB{}}

	// Create a new DBExporter
	exporter := NewDBExporter(dbWorker)

	// Ensure the exporter was created with the correct DB worker
	assert.Equal(t, dbWorker, exporter.dbWorker)
}

func setupTestDB(t *testing.T) (*db.DBWorker, string) {
	// Create a temporary database file
	tempDir, err := os.MkdirTemp("", "pinkbike-test")
	require.NoError(t, err)

	dbPath := filepath.Join(tempDir, "test.db")

	// Create a new DB worker with the temporary database
	dbWorker, err := db.NewDBWorker(dbPath)
	require.NoError(t, err)

	return dbWorker, tempDir
}

func cleanupTestDB(tempDir string, dbWorker *db.DBWorker) {
	dbWorker.Close()
	os.RemoveAll(tempDir)
}

func createTestListings() []listing.Listing {
	now := time.Now()
	return []listing.Listing{
		{
			Title:         "Test Bike 1",
			Year:          "2021",
			Manufacturer:  "TestMfg",
			Model:         "TestModel",
			Price:         "1000",
			Currency:      "USD",
			Condition:     "Good",
			FrameSize:     "M",
			WheelSize:     "29",
			FrameMaterial: "Carbon",
			FrontTravel:   "150",
			RearTravel:    "140",
			NeedsReview:   "0",
			URL:           "https://example.com/bike1",
			Hash:          "hash1",
			FirstSeen:     now,
			LastSeen:      now,
			Active:        true,
			Details: listing.ListingDetails{
				Description:      "Test Description 1",
				Restrictions:     "No Shipping",
				SellerType:       listing.Private,
				OriginalPostDate: now.Add(-24 * time.Hour),
			},
		},
		{
			Title:         "Test Bike 2",
			Year:          "2022",
			Manufacturer:  "AnotherMfg",
			Model:         "AnotherModel",
			Price:         "2000",
			Currency:      "CAD",
			Condition:     "Like New",
			FrameSize:     "L",
			WheelSize:     "27.5",
			FrameMaterial: "Aluminum",
			FrontTravel:   "160",
			RearTravel:    "150",
			NeedsReview:   "0",
			URL:           "https://example.com/bike2",
			Hash:          "hash2",
			FirstSeen:     now,
			LastSeen:      now,
			Active:        true,
			Details: listing.ListingDetails{
				Description:      "Test Description 2",
				Restrictions:     "Local Pickup Only",
				SellerType:       listing.Business,
				OriginalPostDate: now.Add(-48 * time.Hour),
			},
		},
	}
}

func TestDBExporter_Export(t *testing.T) {
	// Setup test database
	dbWorker, tempDir := setupTestDB(t)
	defer cleanupTestDB(tempDir, dbWorker)

	// Create a new DBExporter
	exporter := NewDBExporter(dbWorker)

	// Create test listings
	listings := createTestListings()

	// Export the listings
	err := exporter.Export(listings)
	require.NoError(t, err)

	// Verify the listings were inserted correctly
	var count int
	err = dbWorker.DB.QueryRow("SELECT COUNT(*) FROM listings").Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Check the first listing
	var title, manufacturer, hash, url, price string
	var active int
	err = dbWorker.DB.QueryRow("SELECT title, manufacturer, hash, url, price, active FROM listings WHERE hash = ?", "hash1").
		Scan(&title, &manufacturer, &hash, &url, &price, &active)
	require.NoError(t, err)
	assert.Equal(t, "Test Bike 1", title)
	assert.Equal(t, "TestMfg", manufacturer)
	assert.Equal(t, "hash1", hash)
	assert.Equal(t, "https://example.com/bike1", url)
	assert.Equal(t, "1000", price)
	assert.Equal(t, 1, active)

	// Verify price history was recorded
	var priceHistoryCount int
	err = dbWorker.DB.QueryRow("SELECT COUNT(*) FROM price_history WHERE listing_hash = ?", "hash1").Scan(&priceHistoryCount)
	require.NoError(t, err)
	assert.Equal(t, 1, priceHistoryCount)
}

// Test error handling in Export when transaction fails
func TestDBExporter_ExportTransactionError(t *testing.T) {
	// Create a DB that will fail on Begin()
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer sqlDB.Close()

	// Close the DB to force an error
	sqlDB.Close()

	// Create a DB worker with a closed DB connection
	dbWorker := &db.DBWorker{DB: sqlDB}

	// Create a new DBExporter
	exporter := NewDBExporter(dbWorker)

	// Try to export listings - should fail
	err = exporter.Export([]listing.Listing{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
}

// Test error in exportListings - statement preparation fails
func TestDBExporter_ExportListingsError(t *testing.T) {
	// Setup test database
	dbWorker, tempDir := setupTestDB(t)
	defer cleanupTestDB(tempDir, dbWorker)

	// Create a new DBExporter
	exporter := NewDBExporter(dbWorker)

	// Corrupt the database to cause a statement preparation error
	// Drop the listings table to cause an error
	_, err := dbWorker.DB.Exec("DROP TABLE listings")
	require.NoError(t, err)

	// Try to export listings - should fail due to missing table
	err = exporter.Export(createTestListings())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare statement")
}

// Test exportListing error handling
func TestDBExporter_ExportListingError(t *testing.T) {
	// Setup a test database
	dbWorker, tempDir := setupTestDB(t)
	defer cleanupTestDB(tempDir, dbWorker)

	// Create a transaction
	tx, err := dbWorker.DB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	// Create a statement with insufficient parameters to force an error
	stmt, err := tx.Prepare("INSERT INTO listings (hash) VALUES (?)")
	require.NoError(t, err)
	defer stmt.Close()

	// Create a DBExporter
	exporter := NewDBExporter(dbWorker)

	// Create a test listing
	listing := createTestListings()[0]

	// Test the exportListing method directly with the incorrect statement
	err = exporter.exportListing(stmt, tx, listing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert listing")
}

// Test recordPriceHistory error handling
func TestDBExporter_RecordPriceHistoryError(t *testing.T) {
	// Setup a test database
	dbWorker, tempDir := setupTestDB(t)
	defer cleanupTestDB(tempDir, dbWorker)

	// Create a new DBExporter
	exporter := NewDBExporter(dbWorker)

	// Start a transaction
	tx, err := dbWorker.DB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	// Drop the price_history table to cause an error
	_, err = tx.Exec("DROP TABLE price_history")
	require.NoError(t, err)

	// Create a test listing
	listing := createTestListings()[0]

	// Test the recordPriceHistory method directly with the dropped table
	err = exporter.recordPriceHistory(tx, listing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to record price history")
}

// Test markInactiveListings error handling
func TestDBExporter_MarkInactiveListingsError(t *testing.T) {
	// Setup a test database
	dbWorker, tempDir := setupTestDB(t)
	defer cleanupTestDB(tempDir, dbWorker)

	// Create a new DBExporter
	exporter := NewDBExporter(dbWorker)

	// Start a transaction
	tx, err := dbWorker.DB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	// Drop the listings table to cause an error
	_, err = tx.Exec("DROP TABLE listings")
	require.NoError(t, err)

	// Test the markInactiveListings method directly with the dropped table
	err = exporter.markInactiveListings(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to mark inactive listings")
}

func TestDBExporter_ExportUpdatesExistingListing(t *testing.T) {
	// Setup test database
	dbWorker, tempDir := setupTestDB(t)
	defer cleanupTestDB(tempDir, dbWorker)

	// Create a new DBExporter
	exporter := NewDBExporter(dbWorker)

	// Create a test listing
	listings := createTestListings()[:1] // Just use the first listing

	// Export the listing
	err := exporter.Export(listings)
	require.NoError(t, err)

	// Modify the listing
	listings[0].Price = "1500"                            // Changed price
	listings[0].URL = "https://example.com/bike1-updated" // Changed URL

	// Export the modified listing
	err = exporter.Export(listings)
	require.NoError(t, err)

	// Verify the listing was updated
	var url, price string
	err = dbWorker.DB.QueryRow("SELECT url, price FROM listings WHERE hash = ?", "hash1").
		Scan(&url, &price)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/bike1-updated", url)
	assert.Equal(t, "1500", price)

	// Verify price history was recorded twice
	var priceHistoryCount int
	err = dbWorker.DB.QueryRow("SELECT COUNT(*) FROM price_history WHERE listing_hash = ?", "hash1").Scan(&priceHistoryCount)
	require.NoError(t, err)
	assert.Equal(t, 2, priceHistoryCount)
}

// Test that price history is skipped when price hasn't changed recently
func TestDBExporter_RecordPriceHistorySkipsRecentDuplicates(t *testing.T) {
	// Setup test database
	dbWorker, tempDir := setupTestDB(t)
	defer cleanupTestDB(tempDir, dbWorker)

	// Create a new DBExporter
	exporter := NewDBExporter(dbWorker)

	// Get a test listing
	listing := createTestListings()[0]

	// Manually record a price with the current price
	_, err := dbWorker.DB.Exec(
		"INSERT INTO price_history (listing_hash, price, currency, recorded_at) VALUES (?, ?, ?, datetime('now', '-10 minutes'))",
		listing.Hash, listing.Price, listing.Currency)
	require.NoError(t, err)

	// Start a transaction
	tx, err := dbWorker.DB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	// Try to record price history - should not record a duplicate
	err = exporter.recordPriceHistory(tx, listing)
	require.NoError(t, err)

	// Commit the transaction
	err = tx.Commit()
	require.NoError(t, err)

	// Verify only one price history entry exists
	var count int
	err = dbWorker.DB.QueryRow("SELECT COUNT(*) FROM price_history WHERE listing_hash = ?", listing.Hash).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Now change the price
	listing.Price = "1200"

	// Start a new transaction
	tx, err = dbWorker.DB.Begin()
	require.NoError(t, err)
	defer tx.Rollback()

	// Record price history with new price
	err = exporter.recordPriceHistory(tx, listing)
	require.NoError(t, err)

	// Commit the transaction
	err = tx.Commit()
	require.NoError(t, err)

	// Verify now there are two price history entries
	err = dbWorker.DB.QueryRow("SELECT COUNT(*) FROM price_history WHERE listing_hash = ?", listing.Hash).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestDBExporter_MarkInactiveListings(t *testing.T) {
	// Setup test database
	dbWorker, tempDir := setupTestDB(t)
	defer cleanupTestDB(tempDir, dbWorker)

	// Create a new DBExporter
	exporter := NewDBExporter(dbWorker)

	// Insert a listing directly to control the last_seen timestamp
	_, err := dbWorker.DB.Exec(`
		INSERT INTO listings (
			title, manufacturer, hash, url, price, active, last_seen
		) VALUES (?, ?, ?, ?, ?, ?, datetime('now', '-10 days'))`,
		"Old Bike", "OldMfg", "old-hash", "https://example.com/old-bike", "500", 1)
	require.NoError(t, err)

	// Export an empty list of listings, which should still trigger markInactiveListings
	err = exporter.Export([]listing.Listing{})
	require.NoError(t, err)

	// Verify the old listing was marked as inactive
	var active int
	err = dbWorker.DB.QueryRow("SELECT active FROM listings WHERE hash = ?", "old-hash").Scan(&active)
	require.NoError(t, err)
	assert.Equal(t, 0, active)
}

func TestDBExporter_Close(t *testing.T) {
	// Setup test database
	dbWorker, tempDir := setupTestDB(t)
	defer os.RemoveAll(tempDir) // Clean up the temp dir, but don't close DB here

	// Create a new DBExporter
	exporter := NewDBExporter(dbWorker)

	// Close the exporter
	err := exporter.Close()
	assert.NoError(t, err)

	// Trying to use the closed database should result in an error
	_, err = dbWorker.DB.Exec("SELECT 1")
	assert.Error(t, err) // The database should be closed
}
