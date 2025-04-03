package listing

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type RawListing struct {
	Title, Price, Condition, FrameSize, WheelSize, FrameMaterial, FrontTravel, RearTravel, URL, DetailsLink string
}

type Listing struct {
	Title, Year, Manufacturer, Model, Price, Currency, Condition                         string
	FrameSize, WheelSize, FrameMaterial, FrontTravel, RearTravel, NeedsReview, URL, Hash string
	FirstSeen, LastSeen                                                                  time.Time
	Active                                                                               bool
	Details                                                                              ListingDetails
}

type ListingDetails struct {
	SellerType       SellerType
	OriginalPostDate time.Time
	Description      string
	Restrictions     string
}

type SellerType string

const (
	Private  SellerType = "private"
	Business SellerType = "business"
)

func ParseSellerType(s string) SellerType {
	s = strings.ToLower(strings.TrimSpace(s))
	if strings.Contains(s, "business") {
		return Business
	}
	return Private
}

func (l RawListing) Print() string {
	return fmt.Sprintf("Title: %s\nPrice: %s\n\tCondition: %s\n\tFrame Size: %s\n\tWheel Size: %s\n\tFront Travel: %s\n\tRear Travel: %s\n\tFrame Material: %s\n\tURL: %s\n\t\n",
		l.Title, l.Price, l.Condition, l.FrameSize, l.WheelSize, l.FrontTravel, l.RearTravel, l.FrameMaterial, l.URL)
}

func (l RawListing) PostProcess(exchangeRate float64) Listing {
	newL := Listing{
		Title:         strings.ReplaceAll(l.Title, "\n", ""),
		Year:          extractYear(l.Title),
		Manufacturer:  extractManufacturer(l.Title),
		Model:         extractModel(l.Title),
		Currency:      extractCurrency(l.Price),
		Price:         convertPrice(l.Price, extractCurrency(l.Price), exchangeRate),
		Condition:     l.Condition,
		FrameSize:     l.FrameSize,
		WheelSize:     l.WheelSize,   //todo: convert to float - remove 650B
		FrontTravel:   l.FrontTravel, //todo: remove mm
		RearTravel:    l.RearTravel,  //todo: remove mm
		FrameMaterial: l.FrameMaterial,
		URL:           l.URL,
	}
	newL.Hash = newL.ComputeHash()

	if reason := validateListing(newL); reason != "" {
		newL.NeedsReview = reason
	}

	return newL
}

func validateListing(l Listing) string {
	if l.Price == "" || l.Price == "0" {
		return "price"
	}
	if l.Year == "" {
		return "year"
	}
	if l.Manufacturer == "NoManufacturer" || l.Manufacturer == "" {
		return "manufacturer"
	}
	if l.Model == "NoModelFound" || strings.Contains(l.Model, "Electric") || l.Model == "" {
		return "model"
	}
	if l.Currency == "" {
		return "currency"
	}
	if l.Condition == "" {
		return "condition"
	}
	if l.FrameSize == "" {
		return "frame size"
	}
	if l.WheelSize == "" {
		return "wheel size"
	}
	if l.FrontTravel == "" {
		return "front travel"
	}
	if l.RearTravel == "" {
		return "rear travel"
	}
	if l.FrameMaterial == "" {
		return "frame material"
	}

	return ""
}

func extractYear(title string) string {
	reg := regexp.MustCompile(`\d{4}`)
	s := reg.FindString(title)
	return s
}

func extractCurrency(price string) string {
	reg := regexp.MustCompile(`(CAD|USD)`)
	return reg.FindString(price)
}

func convertPrice(price, currency string, exchangeRate float64) string {
	p := extractPrice(price)

	floatPrice, err := strconv.ParseFloat(p, 32)
	if err != nil {
		return ""
	}

	if currency == "CAD" {
		floatPrice = math.Round(floatPrice * exchangeRate)
		p = fmt.Sprintf("%.0f", floatPrice)
	}

	return p
}

func extractPrice(price string) string {
	reg := regexp.MustCompile(`[0-9,]+`)
	res := reg.FindString(price)
	return strings.ReplaceAll(res, ",", "")
}

func extractManufacturer(title string) string {
	for manufacturer := range bikeModels {
		if strings.Contains(strings.ToLower(title), strings.ToLower(manufacturer)) {
			return manufacturer
		}
	}
	return "NoManufacturer"
}

func extractModel(title string) string {
	manufacturer := extractManufacturer(title)
	bikes := bikeModels[manufacturer]

	for _, model := range bikes {
		if strings.Contains(strings.ToLower(title), strings.ToLower(model.Name)) {
			if model.Purpose == Electric {
				return model.Name + " Electric"
			}
			return model.Name
		}
	}
	return "NoModelFound"
}

func (l Listing) ComputeHash() string {
	// Combine fields that would uniquely identify a bike listing
	uniqueString := strings.Join([]string{
		strings.ToLower(l.Title),
		l.Year,
		l.Model,
		strings.ToLower(l.Condition),
		strings.ToLower(l.FrameSize),
		strings.ToLower(l.FrameMaterial),
		l.FrontTravel,
		l.RearTravel,
	}, "|")

	hasher := sha256.New()
	hasher.Write([]byte(uniqueString))
	return hex.EncodeToString(hasher.Sum(nil))
}
