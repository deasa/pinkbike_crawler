package scraper

import (
	_ "embed"
	"pinkbike-scraper/pkg/listing"
	"strings"
	"testing"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/detailsPage.html
var detailsPageHTML string

//go:embed testdata/listingsPage.html
var listingsPageHTML string

// setupPlaywright creates a new browser instance and page for testing
func setupPlaywright(t *testing.T) (page playwright.Page) {
	t.Helper()

	err := playwright.Install()
	require.NoError(t, err)

	pw, err := playwright.Run()
	require.NoError(t, err)

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	require.NoError(t, err)

	page, err = browser.NewPage()
	require.NoError(t, err)

	return page
}

// TestDetailsScrapeWithHTML tests the detailsScrape function using sample HTML
func TestDetailsScrapeWithHTML(t *testing.T) {
	page := setupPlaywright(t)

	// Set the content of the page to our sample HTML
	err := page.SetContent(detailsPageHTML)
	require.NoError(t, err)

	// Create a scraper instance
	s := &Scraper{}

	// Test the detailsScrape function
	details, err := s.detailsScrape(page)
	require.NoError(t, err)

	// Assert the expected values
	assert.Equal(t, "business", string(details.SellerType))
	expectedDate, _ := time.Parse("2006-01-02", "2024-09-05")
	assert.Equal(t, expectedDate, details.OriginalPostDate)
	assert.Equal(t, "Firm, No Trades, Local pickup only", details.Restrictions)

	expectedDesc := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(expectedDetailedDescription, "\n", ""), "\t", ""), " ", "")

	actualDesc := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(details.Description, "\n", ""), "\t", ""), " ", "")
	assert.Equal(t, expectedDesc, actualDesc)
}

func TestPerformWebScraping(t *testing.T) {
	page := setupPlaywright(t)

	// Set the content of the page to our sample HTML
	err := page.SetContent(listingsPageHTML)
	require.NoError(t, err)

	s := &Scraper{
		page: page,
	}

	listings, err := s.PerformWebScraping(1)
	require.NoError(t, err)

	require.Equal(t, 20, len(listings))

	refinedListings := []listing.Listing{}
	for _, l := range listings {
		list := l.PostProcess(1.0)
		refinedListings = append(refinedListings, list)
	}

	assert.Equal(t, refinedListings[17], listing.Listing{
		Title:         "2022                                                                NEW Scott Contessa Spark 920, size S, 29.52lbs",
		Year:          "2022",
		Manufacturer:  "Scott",
		Model:         "Spark",
		Price:         "3300",
		Currency:      "USD",
		Condition:     "New - Unridden/With Tags",
		FrameSize:     "S",
		WheelSize:     "29",
		FrameMaterial: "Carbon Fiber",
		FrontTravel:   "130 mm",
		RearTravel:    "120 mm",
		URL:           "https://www.pinkbike.com/buysell/3960926/",
	})
}

var expectedDetailedDescription = `
	2024 Orbea Occam LT H20

2024 Demo Bike, you are purchasing from a dealer and will receive receipt to register for 1st owner warranty. All Demo bikes receive mechanic check up after every ride, all parts are replaced or serviced as needed. Very low miles.
*Pedals not included
Frame: Orbea Occam Hydro. Concentric Boost 12x148 rear axle. Pure Trail geometry. Internal cable routing. Asymmetric design. 29" wheels. Linkage compatible with OC multitool

Fork: FOX FACTORY 36/150MM
Rear Shock: Fox Factory Float
Wheels: Race Face AR 30c Tubeless Ready
Tires: Maxxis Dissector 2.40" 60 TPI 3CMaxxTerra Exo TLR
Crankset: Race Face Aeffect 32t
Shifters: Shimano SLX M7100 I-Spec EV
Rear Derailleur: Shimano XT M8100 SGS Shadow Plus
Cassette/Freewheel: Shimano CS-M7100 10-51t 12-Speed
Chain: Shimano M6100
Brakes: Shimano XT M8120 Hydraulic Disc
Rotors: Shimano Centerlock 203/180
Handlebars: OC Mountain Control MC30, Rise20, Width 800
Grips: OC Lock On
Stem: OC Mountain Control MC20, 0º
Headset: FSA 1-1/2" Integrated Aluminium Cup
Seatpost: OC Mountain Control MC21, 31.6mm, Dropper
Saddle: Ergon SM Enduro

24OOM1`
