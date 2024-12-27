package exporter

import (
	"database/sql"
	"fmt"
	"pinkbike-scraper/pkg/listing"

	_ "github.com/mattn/go-sqlite3"
)

type DBExporter struct {
	db *sql.DB
}

func NewDBExporter(dbPath string) (*DBExporter, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := initializeDB(db); err != nil {
		db.Close()
		return nil, err
	}

	return &DBExporter{db: db}, nil
}

func (e *DBExporter) Export(listings []listing.Listing) error {
	tx, err := e.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := e.exportListings(tx, listings); err != nil {
		return err
	}

	if err := e.markInactiveListings(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (e *DBExporter) Close() error {
	return e.db.Close()
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

func (e *DBExporter) ListingExistsWithDetails(hash string) (bool, error) {
	var exists bool
	err := e.db.QueryRow("SELECT EXISTS(SELECT 1 FROM listings WHERE hash = ? AND description IS NOT NULL)", hash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if listing exists: %w", err)
	}
	return exists, nil
}

func (e *DBExporter) exportListings(tx *sql.Tx, listings []listing.Listing) error {
	stmt, err := tx.Prepare(`
        INSERT INTO listings (
            title, year, manufacturer, model, price, currency, 
            condition, frame_size, wheel_size, frame_material,
            front_travel, rear_travel, needs_review, url, hash,
            description, restrictions, seller_type, original_post_date,
            first_seen, last_seen, active
        ) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
                ?, ?, ?, ?,
                CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 1)
        ON CONFLICT(hash) DO UPDATE SET 
            last_seen = CURRENT_TIMESTAMP,
            active = 1,
            url = excluded.url,
            price = excluded.price,
    `)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, l := range listings {
		if err := e.exportListing(stmt, tx, l); err != nil {
			return err
		}
	}

	return nil
}

func (e *DBExporter) exportListing(stmt *sql.Stmt, tx *sql.Tx, l listing.Listing) error {
	hash := l.ComputeHash()
	if _, err := stmt.Exec(
		l.Title, l.Year, l.Manufacturer, l.Model, l.Price,
		l.Currency, l.Condition, l.FrameSize, l.WheelSize,
		l.FrameMaterial, l.FrontTravel, l.RearTravel,
		l.NeedsReview, l.URL, hash,
		l.Details.Description, l.Details.Restrictions, l.Details.SellerType, l.Details.OriginalPostDate,
	); err != nil {
		return fmt.Errorf("failed to insert listing: %w", err)
	}

	return e.recordPriceHistory(tx, l, hash)
}

func (e *DBExporter) recordPriceHistory(tx *sql.Tx, l listing.Listing, hash string) error {
	_, err := tx.Exec(`
        INSERT INTO price_history (listing_hash, price, currency)
        SELECT ?, ?, ?
        WHERE NOT EXISTS (
            SELECT 1 FROM price_history 
            WHERE listing_hash = ? 
            AND price = ? 
            AND recorded_at > datetime('now', '-1 day')
        )
    `, hash, l.Price, l.Currency, hash, l.Price)

	if err != nil {
		return fmt.Errorf("failed to record price history: %w", err)
	}
	return nil
}

func (e *DBExporter) markInactiveListings(tx *sql.Tx) error {
	_, err := tx.Exec(`
        UPDATE listings 
        SET active = 0 
        WHERE datetime(last_seen) < datetime('now', '-7 days')
    `)
	if err != nil {
		return fmt.Errorf("failed to mark inactive listings: %w", err)
	}
	return nil
}
