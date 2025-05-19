package exporter

import (
	"errors"
	"pinkbike-scraper/pkg/db"
	"pinkbike-scraper/pkg/listing"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// TestNewDBExporter tests the creation of a new DBExporter
func TestNewDBExporter(t *testing.T) {
	// Create a mock DB worker
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	dbWorker := &db.DBWorker{DB: mockDB}
	
	// Create a new DBExporter
	exporter := NewDBExporter(dbWorker)
	
	// Assert that the exporter is not nil and has the correct DB worker
	assert.NotNil(t, exporter)
	assert.Equal(t, dbWorker, exporter.dbWorker)
}

// TestDBExporterClose tests the Close method
func TestDBExporterClose(t *testing.T) {
	// Create a mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	
	dbWorker := &db.DBWorker{DB: mockDB}
	exporter := NewDBExporter(dbWorker)
	
	// Expect Close to be called
	mock.ExpectClose()
	
	// Call Close
	err = exporter.Close()
	
	// Assert that there was no error
	assert.NoError(t, err)
	
	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestDBExporterExport tests the Export method
func TestDBExporterExport(t *testing.T) {
	// Create a mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	
	dbWorker := &db.DBWorker{DB: mockDB}
	exporter := NewDBExporter(dbWorker)
	
	// Create test listings
	listings := []listing.Listing{
		{
			Title:        "Test Bike 1",
			Year:         "2021",
			Manufacturer: "TestMfg",
			Model:        "TestModel",
			Price:        "1000",
			Currency:     "USD",
			Hash:         "hash1",
			Details: listing.ListingDetails{
				SellerType:       listing.Private,
				OriginalPostDate: time.Now(),
				Description:      "Test Description",
				Restrictions:     "None",
			},
		},
	}
	
	// Expect a transaction to be started
	mock.ExpectBegin()
	
	// Expect a prepared statement for exportListings
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
	
	// Expect a query for recordPriceHistory
	mock.ExpectExec("INSERT INTO price_history").
		WithArgs(
			listings[0].Hash, listings[0].Price, listings[0].Currency, 
			listings[0].Hash, listings[0].Price,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	// Expect a query for markInactiveListings
	mock.ExpectExec("UPDATE listings").
		WillReturnResult(sqlmock.NewResult(0, 0))
	
	// Expect the transaction to be committed
	mock.ExpectCommit()
	
	// Call Export
	err = exporter.Export(listings)
	
	// Assert that there was no error
	assert.NoError(t, err)
	
	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestDBExporterExportError tests error handling in the Export method
func TestDBExporterExportError(t *testing.T) {
	// Create a mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	
	dbWorker := &db.DBWorker{DB: mockDB}
	exporter := NewDBExporter(dbWorker)
	
	// Create test listings
	listings := []listing.Listing{
		{
			Title:        "Test Bike 1",
			Year:         "2021",
			Manufacturer: "TestMfg",
			Model:        "TestModel",
			Price:        "1000",
			Currency:     "USD",
			Hash:         "hash1",
		},
	}
	
	// Test case 1: Error beginning transaction
	mock.ExpectBegin().WillReturnError(errors.New("begin error"))
	
	// Call Export
	err = exporter.Export(listings)
	
	// Assert that there was an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
	
	// Test case 2: Error in exportListings
	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO listings").WillReturnError(errors.New("prepare error"))
	mock.ExpectRollback()
	
	// Call Export
	err = exporter.Export(listings)
	
	// Assert that there was an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare statement")
	
	// Test case 3: Error in markInactiveListings
	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO listings")
	mock.ExpectExec("INSERT INTO listings").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO price_history").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE listings").WillReturnError(errors.New("update error"))
	mock.ExpectRollback()
	
	// Call Export
	err = exporter.Export(listings)
	
	// Assert that there was an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to mark inactive listings")
	
	// Test case 4: Error in commit
	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO listings")
	mock.ExpectExec("INSERT INTO listings").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO price_history").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("UPDATE listings").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit().WillReturnError(errors.New("commit error"))
	
	// Call Export
	err = exporter.Export(listings)
	
	// Assert that there was an error
	assert.Error(t, err)
	
	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestMarkInactiveListings tests the markInactiveListings method
func TestMarkInactiveListings(t *testing.T) {
	// Create a mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	
	dbWorker := &db.DBWorker{DB: mockDB}
	exporter := NewDBExporter(dbWorker)
	
	// Create a mock transaction
	mock.ExpectBegin()
	tx, _ := mockDB.Begin()
	
	// Expect a query for markInactiveListings
	mock.ExpectExec("UPDATE listings").
		WillReturnResult(sqlmock.NewResult(0, 5)) // 5 rows affected
	
	// Call markInactiveListings
	err = exporter.markInactiveListings(tx)
	
	// Assert that there was no error
	assert.NoError(t, err)
	
	// Test error case
	mock.ExpectExec("UPDATE listings").
		WillReturnError(errors.New("update error"))
	
	// Call markInactiveListings
	err = exporter.markInactiveListings(tx)
	
	// Assert that there was an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to mark inactive listings")
	
	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestRecordPriceHistory tests the recordPriceHistory method
func TestRecordPriceHistory(t *testing.T) {
	// Create a mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	
	dbWorker := &db.DBWorker{DB: mockDB}
	exporter := NewDBExporter(dbWorker)
	
	// Create a mock transaction
	mock.ExpectBegin()
	tx, _ := mockDB.Begin()
	
	// Create test listing
	testListing := listing.Listing{
		Hash:     "testhash",
		Price:    "1000",
		Currency: "USD",
	}
	
	// Expect a query for recordPriceHistory
	mock.ExpectExec("INSERT INTO price_history").
		WithArgs(
			testListing.Hash, testListing.Price, testListing.Currency, 
			testListing.Hash, testListing.Price,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	// Call recordPriceHistory
	err = exporter.recordPriceHistory(tx, testListing)
	
	// Assert that there was no error
	assert.NoError(t, err)
	
	// Test error case
	mock.ExpectExec("INSERT INTO price_history").
		WithArgs(
			testListing.Hash, testListing.Price, testListing.Currency, 
			testListing.Hash, testListing.Price,
		).
		WillReturnError(errors.New("insert error"))
	
	// Call recordPriceHistory
	err = exporter.recordPriceHistory(tx, testListing)
	
	// Assert that there was an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to record price history")
	
	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestExportListing tests the exportListing method
func TestExportListing(t *testing.T) {
	// Create a mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	
	dbWorker := &db.DBWorker{DB: mockDB}
	exporter := NewDBExporter(dbWorker)
	
	// Create a mock transaction
	mock.ExpectBegin()
	tx, _ := mockDB.Begin()
	
	// Create a mock prepared statement
	mock.ExpectPrepare("INSERT INTO listings")
	stmt, _ := tx.Prepare("INSERT INTO listings")
	
	// Create test listing
	testListing := listing.Listing{
		Title:        "Test Bike",
		Year:         "2021",
		Manufacturer: "TestMfg",
		Model:        "TestModel",
		Price:        "1000",
		Currency:     "USD",
		Hash:         "testhash",
		Details: listing.ListingDetails{
			SellerType:       listing.Private,
			OriginalPostDate: time.Now(),
			Description:      "Test Description",
			Restrictions:     "None",
		},
	}
	
	// Expect Exec to be called on the statement
	mock.ExpectExec("INSERT INTO listings").
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	// Expect a query for recordPriceHistory
	mock.ExpectExec("INSERT INTO price_history").
		WithArgs(
			testListing.Hash, testListing.Price, testListing.Currency, 
			testListing.Hash, testListing.Price,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	// Call exportListing
	err = exporter.exportListing(stmt, tx, testListing)
	
	// Assert that there was no error
	assert.NoError(t, err)
	
	// Test error case 1 - Exec error
	mock.ExpectExec("INSERT INTO listings").
		WillReturnError(errors.New("exec error"))
	
	// Call exportListing
	err = exporter.exportListing(stmt, tx, testListing)
	
	// Assert that there was an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert listing")
	
	// Test error case 2 - recordPriceHistory error
	mock.ExpectExec("INSERT INTO listings").
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	mock.ExpectExec("INSERT INTO price_history").
		WithArgs(
			testListing.Hash, testListing.Price, testListing.Currency, 
			testListing.Hash, testListing.Price,
		).
		WillReturnError(errors.New("price history error"))
	
	// Call exportListing
	err = exporter.exportListing(stmt, tx, testListing)
	
	// Assert that there was an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to record price history")
	
	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// TestExportListings tests the exportListings method
func TestExportListings(t *testing.T) {
	// Create a mock DB
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()
	
	dbWorker := &db.DBWorker{DB: mockDB}
	exporter := NewDBExporter(dbWorker)
	
	// Create a mock transaction
	mock.ExpectBegin()
	tx, _ := mockDB.Begin()
	
	// Create test listings
	listings := []listing.Listing{
		{
			Title:        "Test Bike 1",
			Year:         "2021",
			Manufacturer: "TestMfg",
			Model:        "TestModel",
			Price:        "1000",
			Currency:     "USD",
			Hash:         "hash1",
			Details: listing.ListingDetails{
				SellerType:       listing.Private,
				OriginalPostDate: time.Now(),
				Description:      "Test Description",
				Restrictions:     "None",
			},
		},
		{
			Title:        "Test Bike 2",
			Year:         "2022",
			Manufacturer: "TestMfg2",
			Model:        "TestModel2",
			Price:        "2000",
			Currency:     "USD",
			Hash:         "hash2",
			Details: listing.ListingDetails{
				SellerType:       listing.Business,
				OriginalPostDate: time.Now(),
				Description:      "Test Description 2",
				Restrictions:     "None",
			},
		},
	}
	
	// Test case 1: Successful export of multiple listings
	// Expect a prepared statement
	mock.ExpectPrepare("INSERT INTO listings")
	
	// Expect Exec to be called for each listing
	mock.ExpectExec("INSERT INTO listings").
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	// Expect a query for recordPriceHistory for the first listing
	mock.ExpectExec("INSERT INTO price_history").
		WithArgs(
			listings[0].Hash, listings[0].Price, listings[0].Currency, 
			listings[0].Hash, listings[0].Price,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	// Expect Exec to be called for the second listing
	mock.ExpectExec("INSERT INTO listings").
		WillReturnResult(sqlmock.NewResult(2, 1))
	
	// Expect a query for recordPriceHistory for the second listing
	mock.ExpectExec("INSERT INTO price_history").
		WithArgs(
			listings[1].Hash, listings[1].Price, listings[1].Currency, 
			listings[1].Hash, listings[1].Price,
		).
		WillReturnResult(sqlmock.NewResult(2, 1))
	
	// Call exportListings
	err = exporter.exportListings(tx, listings)
	
	// Assert that there was no error
	assert.NoError(t, err)
	
	// Test case 2: Empty listings slice
	emptyListings := []listing.Listing{}
	
	// Expect a prepared statement
	mock.ExpectPrepare("INSERT INTO listings")
	
	// Call exportListings with empty slice
	err = exporter.exportListings(tx, emptyListings)
	
	// Assert that there was no error
	assert.NoError(t, err)
	
	// Test case 3: Prepare error
	mock.ExpectPrepare("INSERT INTO listings").WillReturnError(errors.New("prepare error"))
	
	// Call exportListings
	err = exporter.exportListings(tx, listings)
	
	// Assert that there was an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare statement")
	
	// Test case 4: Error in exportListing for a specific listing
	mock.ExpectPrepare("INSERT INTO listings")
	mock.ExpectExec("INSERT INTO listings").WillReturnError(errors.New("exec error"))
	
	// Call exportListings
	err = exporter.exportListings(tx, listings)
	
	// Assert that there was an error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert listing")
	
	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
