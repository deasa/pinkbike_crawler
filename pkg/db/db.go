package db

import (
	"database/sql"
	"fmt"
	"pinkbike-scraper/pkg/listing"
)

type DBWorker struct {
	DB *sql.DB
}

func NewDBWorker(dbPath string) (*DBWorker, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := initializeDB(db); err != nil {
		db.Close()
		return nil, err
	}
	return &DBWorker{DB: db}, nil
}

func (d *DBWorker) Close() error {
	return d.DB.Close()
}

func initializeDB(db *sql.DB) error {
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

    CREATE INDEX IF NOT EXISTS idx_listings_hash ON listings(hash);
    CREATE INDEX IF NOT EXISTS idx_price_history_listing_hash ON price_history(listing_hash);
    `
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %v", err)
	}

	return nil
}

func (d *DBWorker) ListingExistsWithDetails(hash string) (bool, error) {
	var exists bool
	err := d.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM listings WHERE hash = ? AND description = '')", hash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if listing exists: %w", err)
	}
	return exists, nil
}

func (d *DBWorker) GetListings() ([]listing.Listing, error) {
	rows, err := d.DB.Query(`
		SELECT 
			title, year, manufacturer, model, price, currency,
			condition, frame_size, wheel_size, frame_material,
			front_travel, rear_travel, needs_review, url, hash,
			description, restrictions, seller_type, original_post_date,
			first_seen, last_seen, active
		FROM listings
		WHERE active = 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query listings: %w", err)
	}
	defer rows.Close()

	var listings []listing.Listing
	for rows.Next() {
		var l listing.Listing

		err := rows.Scan(
			&l.Title, &l.Year, &l.Manufacturer, &l.Model, &l.Price, &l.Currency,
			&l.Condition, &l.FrameSize, &l.WheelSize, &l.FrameMaterial,
			&l.FrontTravel, &l.RearTravel, &l.NeedsReview, &l.URL, &l.Hash,
			&l.Details.Description, &l.Details.Restrictions, &l.Details.SellerType, &l.Details.OriginalPostDate,
			&l.FirstSeen, &l.LastSeen, &l.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan listing: %w", err)
		}

		listings = append(listings, l)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return listings, nil
}
