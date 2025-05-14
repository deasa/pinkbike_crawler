# PinkBike Crawler

A Go application for scraping, processing, and exporting bike listings from PinkBike's Buy/Sell section.

## Overview

This application scrapes mountain bike listings from PinkBike's buy/sell section, processes the data to extract structured information, and provides multiple export options including CSV, Google Sheets, and SQLite database storage.

## Features

- Web scraping of PinkBike buy/sell listings with configurable parameters
- Support for different bike types (Enduro, Trail, XC, DH)
- Extraction of detailed listing information including manufacturer, model, price, specifications
- Automatic currency conversion from CAD to USD
- Multiple export options (CSV, Google Sheets, SQLite)
- Ability to read listings from previously exported files
- Detailed listing information scraping for deeper analysis

## Prerequisites

- Go 1.19 or higher
- Google Sheets API credentials (for Google Sheets export)
- Internet connection for web scraping and exchange rate API

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/deasa/pinkbike_crawler.git
   cd pinkbike_crawler
   ```

2. Install dependencies:
   ```
   go mod download
   ```

3. The application uses Playwright for web scraping. The first time you run the application, it will automatically install Playwright dependencies.

## Usage

### Basic Usage

Run the application with default settings:

```
go run main.go
```

This will scrape 5 pages of Enduro bike listings and save them to the SQLite database.

### Command Line Options

The application supports various command-line flags for customization:

```
Usage:
  go run main.go [flags]

Flags:
  -input string           Input mode: 'web', 'file', or 'db' (default "web")
  -filePath string        Path to input file when using file mode
  -numPages int           Number of pages to scrape in web mode (default 5)
  -bikeType string        Type of bike to scrape (enduro, trail, xc, dh) (default "enduro")
  -headless               Run browser in headless mode
  -getDetails             Get detailed listing information
  -export string          Comma-separated list of export modes: 'csv', 'sheets', 'db' (default "db")
  -sheetsCredPath string  Path to Google Sheets credentials (default "pinkbike-exporter-8bc8e681ffa1.json")
  -spreadsheetID string   Google Sheets spreadsheet ID (default "16GYqn_Asp6_MhsJNAiMSphtUpJn6P1nNw-BRQG0s5Ik")
  -dbPath string          Path to SQLite database (default "listings.db")
```

### Example Commands

Scrape 10 pages of Trail bike listings and export to CSV and database:
```
go run main.go -numPages 10 -bikeType trail -export csv,db
```

Load listings from the database and export to Google Sheets:
```
go run main.go -input db -export sheets
```

Scrape listings with detailed information:
```
go run main.go -getDetails
```

## Project Structure

- `main.go`: Application entry point and configuration
- `pkg/scraper`: Web scraping functionality using Playwright
- `pkg/listing`: Listing data structures and processing functions
- `pkg/exporter`: Export functionality for different formats
- `pkg/db`: Database operations for storing and retrieving listings

## Google Sheets Integration

To use the Google Sheets export feature:

1. Create a service account in Google Cloud Platform
2. Download the credentials JSON file
3. Share your Google Sheet with the service account email
4. Provide the credentials file path and spreadsheet ID as command-line arguments

## Development

### Running Tests

```
go test ./...
```

### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- PinkBike for their buy/sell platform
- Playwright for providing the web scraping capabilities
- Exchange Rate API for currency conversion functionality 