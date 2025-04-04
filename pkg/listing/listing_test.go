package listing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractManufacturer(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		// Basic cases
		{"Manufacturer at start", "Specialized Bike Model", "Specialized"},
		{"Manufacturer in middle", "Bike Specialized Model", "Specialized"},
		{"No manufacturer", "Bike Model", "NoManufacturer"},

		// Similar name cases
		{"Marin not Ari", "2025 Marin Alpine Trail XR", "Marin"},
		{"Fezzari not Ari", "2019 Fezzari La Sal Peak (Carbon, Medium, 170/150mm)", "Fezzari"},
		{"Fezzari not Ari at beginning of listing title", "Fezzari La Sal Peak", "Fezzari"},

		// Case sensitivity and spacing
		{"Lowercase manufacturer", "specialized stumpjumper", "Specialized"},
		{"Mixed case manufacturer", "SpecIaliZed Stumpjumper", "Specialized"},
		{"Manufacturer with space", "Santa Cruz Bronson", "Santa Cruz"},

		// Tricky cases with substrings
		{"Trek not RE", "Trek Fuel EX", "Trek"},
		{"Evil not Vi", "Evil Following", "Evil"},
		{"Yeti not YT", "Yeti SB150", "Yeti"},
		{"Rocky Mountain not Mountain Bike", "Rocky Mountain Altitude", "Rocky Mountain"},
		{"GT not just T", "GT Sensor", "GT"},

		// Exact substring cases
		{"Surly exact match", "Surly Karate Monkey", "Surly"},
		{"YT exact match", "YT Capra", "YT"},

		// Boundary detection
		{"Transition in word", "The Transition from hardtail to full suspension", "Transition"},
		{"Specialized at beginning", "Specialized is a bike brand", "Specialized"},
		{"Giant at end", "This bike is a Giant", "Giant"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractManufacturer(tt.arg); got != tt.want {
				t.Errorf("extractManufacturer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractModel(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		// Basic cases
		{"Basic model", "Specialized Stumpjumper", "Stumpjumper"},
		{"Model with spaces", "Santa Cruz Bronson Carbon CC", "Bronson"},
		{"No model", "Specialized Unknown", "NoModelFound"},

		// Case sensitivity
		{"Lowercase model", "specialized stumpjumper", "Stumpjumper"},
		{"Mixed case model", "Trek Fuel eX", "Fuel EX"},

		// Models with variations
		{"Model with version", "Yeti SB150", "SB150"},
		{"Model with number", "Transition Sentinel GX", "Sentinel"},
		{"Model with special chars", "Ibis Ripmo V2", "Ripmo V2"},

		// Electric models
		{"Electric model", "Specialized Turbo Levo", "Turbo Levo Electric"},
		{"Embolden model electric", "Liv Embolden E+", "Embolden E+ Electric"},

		// Partial matches that shouldn't match
		{"Process part of Process X", "Kona Process X", "Process X"},
		{"Ripmo part of Ripmo V2", "Ibis Ripmo V2", "Ripmo V2"},

		// Multi-word models
		{"Meta AM", "Commencal Meta AM", "Meta AM"},
		{"Grand Canyon", "Canyon Grand Canyon", "Grand Canyon"},

		// Edge cases
		{"Model at beginning", "Stumpjumper from Specialized", "Stumpjumper"},
		{"Model at end", "My new Specialized Stumpjumper", "Stumpjumper"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractModel(tt.arg); got != tt.want {
				t.Errorf("extractModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractYear(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		// Basic cases
		{"Year at start", "2022 Bike Model", "2022"},
		{"Year in middle", "Bike 2022 Model", "2022"},
		{"No year", "Bike Model", ""},

		// Year ranges
		{"Very old year (1980)", "1980 Specialized", "1980"},
		{"Old year (1999)", "1999 GT", "1999"},
		{"Current year (2023)", "2023 Transition", "2023"},
		{"Future year (2025)", "2025 Trek", "2025"},
		{"Too old year (1979)", "1979 Nishiki", ""},
		{"Too future year (2050)", "2050 Concept Bike", ""},

		// Multiple years in string
		{"Multiple years", "2020 model upgraded to 2022 specs", "2020"},

		// Not a year
		{"Random 4 digits", "The bike weighs 1234 grams", ""},
		{"Non-year 4 digits", "Trek 8500", ""},

		// Edge cases
		{"Year at end", "Specialized Stumpjumper 2023", "2023"},
		{"Year with punctuation", "2023, Specialized Stumpjumper", "2023"},
		{"Year in parentheses", "Specialized Stumpjumper (2023)", "2023"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractYear(tt.arg); got != tt.want {
				t.Errorf("extractYear() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractCurrency(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		// Basic cases
		{"CAD", "1000 CAD", "CAD"},
		{"USD", "1000 USD", "USD"},
		{"No currency", "1000", ""},

		// Position variations
		{"Currency at end", "$1000 USD", "USD"},
		{"Currency at start", "CAD 1000", "CAD"},
		{"Currency with symbol", "$1000 CAD", "CAD"},

		// Case sensitivity
		{"Lowercase currency", "1000 cad", "CAD"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractCurrency(tt.arg); got != tt.want {
				t.Errorf("extractCurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractPrice(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		// Basic cases
		{"Price with comma", "1,000 CAD", "1000"},
		{"Price without comma", "1000 CAD", "1000"},
		{"No price", "CAD", ""},

		// Currency symbol
		{"Price with dollar sign", "$1000", "1000"},
		{"Price with dollar sign and comma", "$1,000", "1000"},

		// Different formats
		{"Price at end", "Asking 1000", "1000"},
		{"Price at start", "1000 is the price", "1000"},

		// Multiple numbers
		{"Price range", "1000-1500", "1000"},

		// Large and small prices
		{"Large price", "10,000", "10000"},
		{"Small price", "500", "500"},

		// Edge cases
		{"Price with decimal", "1000.00", "1000"},
		{"Price in text", "Price is one thousand dollars", ""},
		{"Price with multiple commas", "1,000,000", "1000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractPrice(tt.arg); got != tt.want {
				t.Errorf("extractPrice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertPrice(t *testing.T) {
	tests := []struct {
		name         string
		price        string
		currency     string
		exchangeRate float64
		want         string
	}{
		{"Price in CAD to CAD", "1000", "CAD", 1.0, "1000"},
		{"Price in CAD to USD with exchange rate 0.75", "1000", "CAD", 0.75, "750"},
		{"Price with comma in CAD to USD", "1,000", "CAD", 0.75, "750"},
		{"Invalid price format", "one thousand", "CAD", 0.75, ""},

		// Additional conversion cases
		{"USD price with no conversion", "1000", "USD", 0.75, "1000"},
		{"Large CAD price conversion", "10000", "CAD", 0.8, "8000"},
		{"Small CAD price conversion", "100", "CAD", 0.7, "70"},

		// Edge cases
		{"Empty price", "", "CAD", 0.75, ""},
		{"Zero price", "0", "CAD", 0.75, "0"},
		{"Very small exchange rate", "1000", "CAD", 0.01, "10"},
		{"Very large exchange rate", "1000", "CAD", 10.0, "10000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertPrice(tt.price, tt.currency, tt.exchangeRate)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostProcess(t *testing.T) {
	tests := []struct {
		name string
		arg  RawListing
		want Listing
	}{
		{
			"Valid listing 1",
			RawListing{
				Title:         "2023 Specialized Status 140 Shimano Build NEW! S4",
				Price:         "$1804 CAD",
				Condition:     "New - Unridden/With Tags",
				FrameSize:     "L",
				FrameMaterial: "Aluminum",
				FrontTravel:   "140 mm",
				RearTravel:    "140 mm",
				WheelSize:     "29",
			},
			Listing{
				Title:         "2023 Specialized Status 140 Shimano Build NEW! S4",
				Price:         "1804",
				Year:          "2023",
				Manufacturer:  "Specialized",
				Model:         "Status",
				Currency:      "CAD",
				Condition:     "New - Unridden/With Tags",
				FrameSize:     "L",
				WheelSize:     "29",
				FrontTravel:   "140 mm",
				RearTravel:    "140 mm",
				FrameMaterial: "Aluminum",
				Hash:          "a2191557da15c9e450149b815c6ae146db749b4e4bd52f30ae5b53aea2b20252",
			},
		},
		{
			"Valid listing 2",
			RawListing{
				Title:         "2018 Commencal Meta AM 4.2 World Cup Edition",
				Price:         "$2550 CAD",
				Condition:     "Good - Used, Mechanically Sound",
				FrameSize:     "M",
				WheelSize:     `27.5 / 650B`,
				FrontTravel:   "170 mm",
				RearTravel:    "160 mm",
				FrameMaterial: "Aluminum",
			},
			Listing{
				Title:         "2018 Commencal Meta AM 4.2 World Cup Edition",
				Price:         "2550",
				Year:          "2018",
				Manufacturer:  "Commencal",
				Model:         "Meta AM",
				Currency:      "CAD",
				Condition:     "Good - Used, Mechanically Sound",
				FrameSize:     "M",
				WheelSize:     "27.5 / 650B",
				FrontTravel:   "170 mm",
				RearTravel:    "160 mm",
				FrameMaterial: "Aluminum",
				Hash:          "a18c8f6017528538d13f5a32353f431815e9c1cd052e6444980a8880d27a551e",
			},
		},
		{
			"Lowercase title",
			RawListing{
				Title:         "2023 specialized stumpjumper expert carbon",
				Price:         "$4500 USD",
				Condition:     "Excellent - Lightly Ridden",
				FrameSize:     "M",
				WheelSize:     "29",
				FrontTravel:   "150 mm",
				RearTravel:    "140 mm",
				FrameMaterial: "Carbon Fiber",
			},
			Listing{
				Title:         "2023 specialized stumpjumper expert carbon",
				Price:         "4500",
				Year:          "2023",
				Manufacturer:  "Specialized",
				Model:         "Stumpjumper",
				Currency:      "USD",
				Condition:     "Excellent - Lightly Ridden",
				FrameSize:     "M",
				WheelSize:     "29",
				FrontTravel:   "150 mm",
				RearTravel:    "140 mm",
				FrameMaterial: "Carbon Fiber",
				Hash:          "3807e0456dd04434f323ed65e1f3da9e3e600de12a469ad0914989fb0bafdc21",
			},
		},
		{
			"Electric bike",
			RawListing{
				Title:         "2022 Specialized Turbo Levo Comp Carbon",
				Price:         "$7000 CAD",
				Condition:     "Good - Used, Mechanically Sound",
				FrameSize:     "L",
				WheelSize:     "29",
				FrontTravel:   "160 mm",
				RearTravel:    "150 mm",
				FrameMaterial: "Carbon Fiber",
			},
			Listing{
				Title:         "2022 Specialized Turbo Levo Comp Carbon",
				Price:         "7000",
				Year:          "2022",
				Manufacturer:  "Specialized",
				Model:         "Turbo Levo Electric",
				Currency:      "CAD",
				Condition:     "Good - Used, Mechanically Sound",
				FrameSize:     "L",
				WheelSize:     "29",
				FrontTravel:   "160 mm",
				RearTravel:    "150 mm",
				FrameMaterial: "Carbon Fiber",
				NeedsReview:   "model", // electric bikes are supposed to be flagged
				Hash:          "1d1043e765bda3ca7870296b83034076074f312c15251440f3a2e8673082948f",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.arg.PostProcess(1.0)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestComputeHash(t *testing.T) {
	tests := []struct {
		name    string
		listing Listing
		want    string
	}{
		{
			"Basic hash computation",
			Listing{
				Title:         "2023 Specialized Status 140 Shimano Build NEW! S4",
				Year:          "2023",
				Manufacturer:  "Specialized",
				Model:         "Status",
				Condition:     "New",
				FrameSize:     "L",
				FrameMaterial: "Aluminum",
				FrontTravel:   "140 mm",
				RearTravel:    "140 mm",
			},
			"c85c55ddab0c78a6611c22ba84056553800c9f3b431b1ce74647e57f4a6c6e47",
		},
		{
			"Case insensitivity",
			Listing{
				Title:         "2022 specialized stumpjumper",
				Year:          "2022",
				Manufacturer:  "Specialized",
				Model:         "Stumpjumper",
				Condition:     "excellent",
				FrameSize:     "l",
				FrameMaterial: "carbon",
				FrontTravel:   "150 mm",
				RearTravel:    "140 mm",
			},
			"e54622bd97d4665e4fd95ab0e29b916e3963c4d8abf4bbe3e02ae2867bfec834",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.listing.ComputeHash()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateListing(t *testing.T) {
	tests := []struct {
		name    string
		listing Listing
		want    string
	}{
		{
			"Valid listing",
			Listing{
				Title:         "2022 Specialized Stumpjumper",
				Year:          "2022",
				Manufacturer:  "Specialized",
				Model:         "Stumpjumper",
				Price:         "5000",
				Currency:      "USD",
				Condition:     "Excellent",
				FrameSize:     "L",
				WheelSize:     "29",
				FrameMaterial: "Carbon",
				FrontTravel:   "150 mm",
				RearTravel:    "140 mm",
			},
			"",
		},
		{
			"Missing price",
			Listing{
				Title:         "2022 Specialized Stumpjumper",
				Year:          "2022",
				Manufacturer:  "Specialized",
				Model:         "Stumpjumper",
				Price:         "",
				Currency:      "USD",
				Condition:     "Excellent",
				FrameSize:     "L",
				WheelSize:     "29",
				FrameMaterial: "Carbon",
				FrontTravel:   "150 mm",
				RearTravel:    "140 mm",
			},
			"price",
		},
		{
			"Missing year",
			Listing{
				Title:         "Specialized Stumpjumper",
				Year:          "",
				Manufacturer:  "Specialized",
				Model:         "Stumpjumper",
				Price:         "5000",
				Currency:      "USD",
				Condition:     "Excellent",
				FrameSize:     "L",
				WheelSize:     "29",
				FrameMaterial: "Carbon",
				FrontTravel:   "150 mm",
				RearTravel:    "140 mm",
			},
			"year",
		},
		{
			"Unknown manufacturer",
			Listing{
				Title:         "2022 Unknown Stumpjumper",
				Year:          "2022",
				Manufacturer:  "NoManufacturer",
				Model:         "Stumpjumper",
				Price:         "5000",
				Currency:      "USD",
				Condition:     "Excellent",
				FrameSize:     "L",
				WheelSize:     "29",
				FrameMaterial: "Carbon",
				FrontTravel:   "150 mm",
				RearTravel:    "140 mm",
			},
			"manufacturer",
		},
		{
			"Unknown model",
			Listing{
				Title:         "2022 Specialized Unknown",
				Year:          "2022",
				Manufacturer:  "Specialized",
				Model:         "NoModelFound",
				Price:         "5000",
				Currency:      "USD",
				Condition:     "Excellent",
				FrameSize:     "L",
				WheelSize:     "29",
				FrameMaterial: "Carbon",
				FrontTravel:   "150 mm",
				RearTravel:    "140 mm",
			},
			"model",
		},
		{
			"Electric model",
			Listing{
				Title:         "2022 Specialized Turbo Levo",
				Year:          "2022",
				Manufacturer:  "Specialized",
				Model:         "Turbo Levo Electric",
				Price:         "5000",
				Currency:      "USD",
				Condition:     "Excellent",
				FrameSize:     "L",
				WheelSize:     "29",
				FrameMaterial: "Carbon",
				FrontTravel:   "150 mm",
				RearTravel:    "140 mm",
			},
			"model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateListing(tt.listing)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseSellerType(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want SellerType
	}{
		{"Private seller", "private", Private},
		{"Business seller", "business", Business},
		{"Business seller with extra text", "business seller", Business},
		{"Private by default", "seller", Private},
		{"Trimmed input", "  private  ", Private},
		{"Case insensitive", "BUSINESS", Business},
		{"Empty string", "", Private},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSellerType(tt.arg)
			assert.Equal(t, tt.want, got)
		})
	}
}
