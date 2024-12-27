package exporter

import (
	"pinkbike-scraper/pkg/listing"
)

// Exporter interface defines methods for exporting listings
type Exporter interface {
	Export(listings []listing.Listing) error
	Close() error
}
