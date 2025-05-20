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

// TestNewDBExporter tests the NewDBExporter function
func TestNewDBExporter(t *testing.T) {
	// Create a mock DB
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer mockDB.Close()

	// Create a DBWorker with the mock DB
	dbWorker := &db.DBWorker{DB: mockDB}

	// Call the function under test
	exporter := NewDBExporter(dbWorker)

	// Assert that the exporter is not nil and has the correct DBWorker
	assert.NotNil(t, exporter)
	assert.Equal(t, dbWorker, exporter.dbWorker)
}

// TestExport tests the Export method
func TestExport(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		listings       []listing.Listing
		setupMock      func(mock sqlmock.Sqlmock)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "Success",
			listings: createTestListings(),
			setupMock: func(mock sqlmock.Sqlmock) {
				// Expect transaction begin
				mock.ExpectBegin()

				// Expect prepare statement for exportListings
				mock.ExpectPrepare("INSERT INTO listings").
					ExpectExec().
					WithArgs(
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Expect query for recordPriceHistory
				mock.ExpectExec("INSERT INTO price_history").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Expect markInactiveListings
				mock.ExpectExec("UPDATE listings").
					WillReturnResult(sqlmock.NewResult(0, 0))

				// Expect transaction commit
				mock.ExpectCommit()
			},
			expectedError: false,
		},
		{
			name:     "Begin Transaction Error",
			listings: createTestListings(),
			setupMock: func(mock sqlmock.Sqlmock) {
				// Expect transaction begin with error
				mock.ExpectBegin().WillReturnError(errors.New("begin transaction error"))
			},
			expectedError:  true,
			expectedErrMsg: "failed to begin transaction: begin transaction error",
		},
		{
			name:     "Export Listings Error",
			listings: createTestListings(),
			setupMock: func(mock sqlmock.Sqlmock) {
				// Expect transaction begin
				mock.ExpectBegin()

				// Expect prepare statement with error
				mock.ExpectPrepare("INSERT INTO listings").WillReturnError(errors.New("prepare statement error"))

				// Expect transaction rollback
				mock.ExpectRollback()
			},
			expectedError:  true,
			expectedErrMsg: "failed to prepare statement: prepare statement error",
		},
		{
			name:     "Mark Inactive Listings Error",
			listings: createTestListings(),
			setupMock: func(mock sqlmock.Sqlmock) {
				// Expect transaction begin
				mock.ExpectBegin()

				// Expect prepare statement for exportListings
				mock.ExpectPrepare("INSERT INTO listings").
					ExpectExec().
					WithArgs(
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Expect query for recordPriceHistory
				mock.ExpectExec("INSERT INTO price_history").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Expect markInactiveListings with error
				mock.ExpectExec("UPDATE listings").
					WillReturnError(errors.New("mark inactive error"))

				// Expect transaction rollback
				mock.ExpectRollback()
			},
			expectedError:  true,
			expectedErrMsg: "failed to mark inactive listings: mark inactive error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock DB
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock DB: %v", err)
			}
			defer mockDB.Close()

			// Setup mock expectations
			tc.setupMock(mock)

			// Create a DBWorker with the mock DB
			dbWorker := &db.DBWorker{DB: mockDB}

			// Create exporter
			exporter := NewDBExporter(dbWorker)

			// Call the function under test
			err = exporter.Export(tc.listings)

			// Check error
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestClose tests the Close method
func TestClose(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		setupMock      func(mock sqlmock.Sqlmock)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "Success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectClose()
			},
			expectedError: false,
		},
		{
			name: "Close Error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectClose().WillReturnError(errors.New("close error"))
			},
			expectedError:  true,
			expectedErrMsg: "close error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock DB
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock DB: %v", err)
			}

			// Setup mock expectations
			tc.setupMock(mock)

			// Create a DBWorker with the mock DB
			dbWorker := &db.DBWorker{DB: mockDB}

			// Create exporter
			exporter := NewDBExporter(dbWorker)

			// Call the function under test
			err = exporter.Close()

			// Check error
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Helper function to create test listings
func createTestListings() []listing.Listing {
	return []listing.Listing{
		{
			Title:         "Test Bike 2022",
			Year:          "2022",
			Manufacturer:  "Test",
			Model:         "Bike",
			Price:         "1000",
			Currency:      "USD",
			Condition:     "Good",
			FrameSize:     "M",
			WheelSize:     "29",
			FrameMaterial: "Carbon",
			FrontTravel:   "150",
			RearTravel:    "140",
			NeedsReview:   "",
			URL:           "https://example.com/bike",
			Hash:          "testhash123",
			FirstSeen:     time.Now(),
			LastSeen:      time.Now(),
			Active:        true,
			Details: listing.ListingDetails{
				Description:      "Test description",
				Restrictions:     "No shipping",
				SellerType:       listing.Private,
				OriginalPostDate: time.Now(),
			},
		},
	}
}

// TestExportListings tests the exportListings method
func TestExportListings(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		listings       []listing.Listing
		setupMock      func(mock sqlmock.Sqlmock)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:     "Success",
			listings: createTestListings(),
			setupMock: func(mock sqlmock.Sqlmock) {
				// Expect begin transaction
				mock.ExpectBegin()

				// Expect prepare statement
				mock.ExpectPrepare("INSERT INTO listings").
					ExpectExec().
					WithArgs(
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Expect query for recordPriceHistory
				mock.ExpectExec("INSERT INTO price_history").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: false,
		},
		{
			name:     "Prepare Statement Error",
			listings: createTestListings(),
			setupMock: func(mock sqlmock.Sqlmock) {
				// Expect begin transaction
				mock.ExpectBegin()

				// Expect prepare statement with error
				mock.ExpectPrepare("INSERT INTO listings").WillReturnError(errors.New("prepare statement error"))
			},
			expectedError:  true,
			expectedErrMsg: "failed to prepare statement: prepare statement error",
		},
		{
			name:     "Export Listing Error",
			listings: createTestListings(),
			setupMock: func(mock sqlmock.Sqlmock) {
				// Expect begin transaction
				mock.ExpectBegin()

				// Expect prepare statement
				mock.ExpectPrepare("INSERT INTO listings").
					ExpectExec().
					WithArgs(
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
						sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
					).
					WillReturnError(errors.New("exec error"))
			},
			expectedError:  true,
			expectedErrMsg: "failed to insert listing: exec error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock DB
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock DB: %v", err)
			}
			defer mockDB.Close()

			// Setup mock expectations
			tc.setupMock(mock)

			// Create a DBWorker with the mock DB
			dbWorker := &db.DBWorker{DB: mockDB}

			// Create exporter
			exporter := NewDBExporter(dbWorker)

			// Begin transaction
			tx, err := mockDB.Begin()
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}

			// Call the function under test
			err = exporter.exportListings(tx, tc.listings)

			// Check error
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestRecordPriceHistory tests the recordPriceHistory method
func TestRecordPriceHistory(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		listing        listing.Listing
		setupMock      func(mock sqlmock.Sqlmock)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name:    "Success",
			listing: createTestListings()[0],
			setupMock: func(mock sqlmock.Sqlmock) {
				// Expect begin transaction
				mock.ExpectBegin()

				// Expect query
				mock.ExpectExec("INSERT INTO price_history").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedError: false,
		},
		{
			name:    "Exec Error",
			listing: createTestListings()[0],
			setupMock: func(mock sqlmock.Sqlmock) {
				// Expect begin transaction
				mock.ExpectBegin()

				// Expect query with error
				mock.ExpectExec("INSERT INTO price_history").
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("exec error"))
			},
			expectedError:  true,
			expectedErrMsg: "failed to record price history: exec error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock DB
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock DB: %v", err)
			}
			defer mockDB.Close()

			// Setup mock expectations
			tc.setupMock(mock)

			// Create a DBWorker with the mock DB
			dbWorker := &db.DBWorker{DB: mockDB}

			// Create exporter
			exporter := NewDBExporter(dbWorker)

			// Begin transaction
			tx, err := mockDB.Begin()
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}

			// Call the function under test
			err = exporter.recordPriceHistory(tx, tc.listing)

			// Check error
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// TestMarkInactiveListings tests the markInactiveListings method
func TestMarkInactiveListings(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		setupMock      func(mock sqlmock.Sqlmock)
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "Success",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Expect begin transaction
				mock.ExpectBegin()

				// Expect query
				mock.ExpectExec("UPDATE listings").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedError: false,
		},
		{
			name: "Exec Error",
			setupMock: func(mock sqlmock.Sqlmock) {
				// Expect begin transaction
				mock.ExpectBegin()

				// Expect query with error
				mock.ExpectExec("UPDATE listings").
					WillReturnError(errors.New("exec error"))
			},
			expectedError:  true,
			expectedErrMsg: "failed to mark inactive listings: exec error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock DB
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock DB: %v", err)
			}
			defer mockDB.Close()

			// Setup mock expectations
			tc.setupMock(mock)

			// Create a DBWorker with the mock DB
			dbWorker := &db.DBWorker{DB: mockDB}

			// Create exporter
			exporter := NewDBExporter(dbWorker)

			// Begin transaction
			tx, err := mockDB.Begin()
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}

			// Call the function under test
			err = exporter.markInactiveListings(tx)

			// Check error
			if tc.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrMsg)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
