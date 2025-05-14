package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"pinkbike-scraper/pkg/db"
	"pinkbike-scraper/pkg/exporter"
	"pinkbike-scraper/pkg/listing"
	"pinkbike-scraper/pkg/mlb"
	"pinkbike-scraper/pkg/scraper"
)

const (
	urlBase       = "https://www.pinkbike.com/buysell/list/"
	spreadsheetID = "16GYqn_Asp6_MhsJNAiMSphtUpJn6P1nNw-BRQG0s5Ik"
)

type Config struct {
	// Input configuration
	InputMode  string
	FilePath   string
	NumPages   int
	BikeType   string
	Headless   bool
	GetDetails bool

	// Export configuration
	ExportModes    []string
	SheetsCredPath string
	SpreadsheetID  string
	DBPath         string

	// MLB player configuration
	PlayerName string
	PlayerMode bool
}

type ExchangeRateResponse struct {
	Rates map[string]float64
}

func main() {
	cfg := parseFlags()

	// If PlayerMode is enabled, get player data and exit
	if cfg.PlayerMode {
		if cfg.PlayerName == "" {
			log.Fatal("Player name is required when using player mode")
		}
		getPlayerData(cfg.PlayerName)
		return
	}

	dbWorker, err := db.NewDBWorker(cfg.DBPath)
	if err != nil {
		log.Fatalf("could not create database worker: %v", err)
	}
	defer dbWorker.Close()

	scraper, err := scraper.NewScraper(
		cfg.FilePath,
		cfg.Headless,
		urlBase,
		getBikeType(cfg.BikeType),
		dbWorker,
	)
	if err != nil {
		log.Fatalf("could not create scraper: %v", err)
	}
	defer scraper.Close()

	// Setup exporters
	exporters, err := setupExporters(cfg, dbWorker)
	if err != nil {
		log.Fatalf("failed to setup exporters: %v", err)
	}
	defer func() {
		for _, e := range exporters {
			e.Close()
		}
	}()

	// Get exchange rate
	exchangeRate, err := getCADtoUSDExchangeRate()
	if err != nil {
		log.Fatalf("could not get exchange rate: %v", err)
	}
	fmt.Printf("CAD to USD exchange rate: %f\n", exchangeRate)

	// Get listings
	listings, err := getListings(cfg, dbWorker, scraper, exchangeRate)
	if err != nil {
		log.Fatalf("failed to get listings: %v", err)
	}

	if cfg.GetDetails {
		listings, err = scraper.GetDetailedListings(listings)
		if err != nil {
			log.Fatalf("failed to get detailed listings: %v", err)
		}
	}

	// Export listings
	for _, exp := range exporters {
		if err := exp.Export(listings); err != nil {
			log.Printf("export error: %v", err)
		}
	}
}

// getPlayerData fetches and displays MLB player data
func getPlayerData(playerName string) {
	fmt.Printf("Searching for player: %s\n", playerName)

	players, err := mlb.SearchPlayer(playerName)
	if err != nil {
		log.Fatalf("Failed to search for player: %v", err)
	}

	if len(players) == 0 {
		fmt.Println("No players found with that name")
		return
	}

	// If multiple players are found, display a list
	if len(players) > 1 {
		fmt.Printf("Found %d players matching '%s':\n", len(players), playerName)
		for i, player := range players {
			fmt.Printf("%d. %s (%s) - %s\n", i+1, player.Name, player.Position, player.Team)
		}

		// Ask user to select a player
		var selection int
		fmt.Print("Enter the number of the player to see details (or 0 to exit): ")
		fmt.Scan(&selection)

		if selection < 1 || selection > len(players) {
			fmt.Println("Exiting...")
			return
		}

		// Get detailed info for the selected player
		selectedPlayer, err := mlb.GetPlayerByID(players[selection-1].PlayerID)
		if err != nil {
			log.Fatalf("Failed to get player details: %v", err)
		}

		displayPlayerInfo(selectedPlayer)
	} else {
		// Only one player found, display details directly
		displayPlayerInfo(players[0])
	}
}

// displayPlayerInfo prints detailed player information
func displayPlayerInfo(player mlb.Player) {
	fmt.Println("\n===== Player Information =====")
	fmt.Printf("Name: %s\n", player.Name)
	fmt.Printf("Position: %s\n", player.Position)
	fmt.Printf("Team: %s\n", player.Team)
	fmt.Printf("Jersey Number: %s\n", player.JerseyNumber)
	fmt.Printf("Status: %s\n", player.Status)
	fmt.Printf("Age: %s\n", player.Age)
	fmt.Printf("Birth Date: %s\n", player.BirthDate)
	fmt.Printf("Birth Place: %s\n", player.BirthPlace)
	fmt.Printf("Height: %s\n", player.Height)
	fmt.Printf("Weight: %s lbs\n", player.Weight)
	fmt.Printf("Bats: %s\n", player.Bats)
	fmt.Printf("Throws: %s\n", player.Throws)
	fmt.Printf("Pro Debut: %s\n", player.ProDebutDate)
	fmt.Println("=============================")
}

func parseFlags() *Config {
	cfg := &Config{}

	// Input flags
	flag.StringVar(&cfg.InputMode, "input", "web", "Input mode: 'web', 'file', or 'db'")
	flag.StringVar(&cfg.FilePath, "filePath", "", "Path to input file when using file mode")
	flag.IntVar(&cfg.NumPages, "numPages", 5, "Number of pages to scrape in web mode")
	flag.StringVar(&cfg.BikeType, "bikeType", "enduro", "Type of bike to scrape")
	flag.BoolVar(&cfg.Headless, "headless", false, "Run browser in headless mode")
	flag.BoolVar(&cfg.GetDetails, "getDetails", false, "Get detailed listing information")

	// Export flags
	var exportModes string
	flag.StringVar(&exportModes, "export", "db", "Comma-separated list of export modes: 'csv', 'sheets', 'db'")
	flag.StringVar(&cfg.SheetsCredPath, "sheetsCredPath", "pinkbike-exporter-8bc8e681ffa1.json", "Path to Google Sheets credentials")
	flag.StringVar(&cfg.SpreadsheetID, "spreadsheetID", spreadsheetID, "Google Sheets spreadsheet ID")
	flag.StringVar(&cfg.DBPath, "dbPath", "listings.db", "Path to SQLite database")

	// MLB player flags
	flag.StringVar(&cfg.PlayerName, "playerName", "", "Name of MLB player to search for")
	flag.BoolVar(&cfg.PlayerMode, "playerMode", false, "Enable MLB player lookup mode")

	flag.Parse()

	// Parse export modes
	cfg.ExportModes = strings.Split(exportModes, ",")
	return cfg
}

func setupExporters(cfg *Config, dbWorker *db.DBWorker) ([]exporter.Exporter, error) {
	var exporters []exporter.Exporter

	for _, mode := range cfg.ExportModes {
		switch strings.TrimSpace(mode) {
		case "csv":
			fileName := getFileName(getBikeType(cfg.BikeType))
			csvExp := exporter.NewCSVExporter(
				"runs/"+fileName,
				"runs/suspect_"+fileName,
			)
			exporters = append(exporters, csvExp)

		case "sheets":
			sheetsExp, err := exporter.NewSheetsExporter(
				cfg.SheetsCredPath,
				cfg.SpreadsheetID,
			)
			if err != nil {
				return nil, fmt.Errorf("could not create sheets exporter: %v", err)
			}
			exporters = append(exporters, sheetsExp)

		case "db":
			dbExp := exporter.NewDBExporter(dbWorker)
			exporters = append(exporters, dbExp)
		}
	}

	return exporters, nil
}

func getListings(cfg *Config, dbWorker *db.DBWorker, scraper *scraper.Scraper, exchangeRate float64) ([]listing.Listing, error) {
	switch cfg.InputMode {
	case "file":
		return readListingsFromFile(cfg.FilePath)
	case "db":
		return dbWorker.GetListings()
	case "web":
		rawListings, err := scraper.PerformWebScraping(cfg.NumPages)
		if err != nil {
			return nil, err
		}
		var refinedListings []listing.Listing
		for _, l := range rawListings {
			refinedListings = append(refinedListings, l.PostProcess(exchangeRate))
		}
		return refinedListings, nil
	default:
		return nil, fmt.Errorf("invalid input mode: %s", cfg.InputMode)
	}
}

func getFileName(bikeType scraper.BikeType) string {
	bt := string(bikeType)
	fileName := fmt.Sprintf("%sListings%s.csv", bt, time.Now().Format("2006-01-02"))
	return fileName
}

func getBikeType(bikeType string) scraper.BikeType {
	var bikeTypeVal scraper.BikeType
	switch bikeType {
	case "enduro":
		bikeTypeVal = scraper.Enduro
	case "trail":
		bikeTypeVal = scraper.Trail
	case "xc":
		bikeTypeVal = scraper.XC
	case "dh":
		bikeTypeVal = scraper.DH
	default:
		log.Fatalf("invalid bike type: %s", bikeType)
	}
	return bikeTypeVal
}

func getCADtoUSDExchangeRate() (float64, error) {
	resp, err := http.Get("https://api.exchangerate-api.com/v4/latest/CAD")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var data ExchangeRateResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, err
	}

	return data.Rates["USD"], nil
}

// ReadListingsFromFile reads listings from the configured file path
func readListingsFromFile(filePath string) ([]listing.Listing, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read file: %v", err)
	}

	listings := make([]listing.Listing, 0, len(records))
	for _, record := range records {
		l := listing.Listing{
			Title:         record[0],
			Year:          record[1],
			Price:         record[2],
			Currency:      record[3],
			Condition:     record[4],
			FrameSize:     record[5],
			WheelSize:     record[6],
			FrontTravel:   record[7],
			RearTravel:    record[8],
			FrameMaterial: record[9],
		}

		listings = append(listings, l)
	}

	return listings, nil
}

// todo implement "a.k.a" for models and manufacturers so that they all get normalized to a single name
// priority is on the manufacturer though because we probably wont use the model name in the prediction
