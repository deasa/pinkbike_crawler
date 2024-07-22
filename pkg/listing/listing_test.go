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
		{"Manufacturer at start", "Specialized Bike Model", "Specialized"},
		{"Manufacturer in middle", "Bike Specialized Model", "Specialized"},
		{"No manufacturer", "Bike Model", "NoManufacturer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractManufacturer(tt.arg); got != tt.want {
				t.Errorf("extractManufacturer() = %v, want %v", got, tt.want)
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
		{"Year at start", "2022 Bike Model", "2022"},
		{"Year in middle", "Bike 2022 Model", "2022"},
		{"No year", "Bike Model", ""},
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
		{"CAD", "1000 CAD", "CAD"},
		{"USD", "1000 USD", "USD"},
		{"No currency", "1000", ""},
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
		{"Price with comma", "1,000 CAD", "1000"},
		{"Price without comma", "1000 CAD", "1000"},
		{"No price", "CAD", ""},
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
				Title:         "2024 Transition Spire AXS T-Type Fox Factory Reserve Wheels",
				Price:         "$5300 USD",
				Condition:     "Excellent - Lightly Ridden",
				FrameSize:     "L",
				WheelSize:     `29`,
				FrontTravel:   "170 mm",
				RearTravel:    "170 mm",
				FrameMaterial: "Carbon Fiber",
			},
			Listing{
				Title:         "2024 Transition Spire AXS T-Type Fox Factory Reserve Wheels",
				Price:         "5300",
				Year:          "2024",
				Manufacturer:  "Transition",
				Model:         "Spire",
				Currency:      "USD",
				Condition:     "Excellent - Lightly Ridden",
				FrameSize:     "L",
				WheelSize:     "29",
				FrontTravel:   "170 mm",
				RearTravel:    "170 mm",
				FrameMaterial: "Carbon Fiber",
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
