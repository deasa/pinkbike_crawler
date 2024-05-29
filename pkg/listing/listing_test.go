package listing

import (
	"testing"
)

func TestExtractManufacturer(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{"Manufacturer at start", "Specialized Bike Model", "Specialized"},
		{"Manufacturer in middle", "Bike Specialized Model", "Specialized"},
		{"No manufacturer", "Bike Model", ""},
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
        {"Price with comma", "1,000 CAD", "1,000"},
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