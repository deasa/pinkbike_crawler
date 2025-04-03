package exporter

import (
	"database/sql"
	"fmt"
	"pinkbike-scraper/pkg/db"
	"pinkbike-scraper/pkg/listing"

	_ "github.com/mattn/go-sqlite3"
)

type DBExporter struct {
	dbWorker *db.DBWorker
}

func NewDBExporter(dbWorker *db.DBWorker) *DBExporter {
	return &DBExporter{dbWorker: dbWorker}
}

func (e *DBExporter) Export(listings []listing.Listing) error {
	tx, err := e.dbWorker.DB.Begin()
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
	return e.dbWorker.DB.Close()
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
            price = excluded.price
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
	if _, err := stmt.Exec(
		l.Title, l.Year, l.Manufacturer, l.Model, l.Price,
		l.Currency, l.Condition, l.FrameSize, l.WheelSize,
		l.FrameMaterial, l.FrontTravel, l.RearTravel,
		l.NeedsReview, l.URL, l.Hash,
		l.Details.Description, l.Details.Restrictions, l.Details.SellerType, l.Details.OriginalPostDate,
	); err != nil {
		return fmt.Errorf("failed to insert listing: %w", err)
	}

	return e.recordPriceHistory(tx, l)
}

func (e *DBExporter) recordPriceHistory(tx *sql.Tx, l listing.Listing) error {
	_, err := tx.Exec(`
        INSERT INTO price_history (listing_hash, price, currency)
        SELECT ?, ?, ?
        WHERE NOT EXISTS (
            SELECT 1 FROM price_history 
            WHERE listing_hash = ? 
            AND price = ? 
            AND recorded_at > datetime('now', '-1 day')
        )
    `, l.Hash, l.Price, l.Currency, l.Hash, l.Price)

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
