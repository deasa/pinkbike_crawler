package listing

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type RawListing struct {
	Title, Price, Condition, FrameSize, WheelSize, FrameMaterial, FrontTravel, RearTravel string
}

type Listing struct {
	Title, Year, Manufacturer, Model, Price, Currency, Condition, FrameSize, WheelSize, FrameMaterial, FrontTravel, RearTravel string
	NeedsReview                                                                                                                bool
}

func (l RawListing) Print() string {
	return fmt.Sprintf("Title: %s\nPrice: %s\n\tCondition: %s\n\tFrame Size: %s\n\tWheel Size: %s\n\tFront Travel: %s\n\tRear Travel: %s\n\tFrame Material: %s\n\t\n",
		l.Title, l.Price, l.Condition, l.FrameSize, l.WheelSize, l.FrontTravel, l.RearTravel, l.FrameMaterial)
}

func (l RawListing) PostProcess(exchangeRate float64) Listing {
	newL := Listing{
		Title:         l.Title,
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
	}

	if !validateListing(newL) {
		newL.NeedsReview = true
	}

	return newL
}

func validateListing(l Listing) bool {
	if l.Price == "" || l.Price == "0" {
		return false
	}
	if l.Year == "" {
		return false
	}
	if l.Manufacturer == "NoManufacturer" || l.Manufacturer == "" {
		return false
	}
	if l.Model == "NoModelFound" || l.Model == "" {
		return false
	}
	if l.Currency == "" {
		return false
	}
	if l.Condition == "" {
		return false
	}
	if l.FrameSize == "" {
		return false
	}
	if l.WheelSize == "" {
		return false
	}
	if l.FrontTravel == "" {
		return false
	}
	if l.RearTravel == "" {
		return false
	}
	if l.FrameMaterial == "" {
		return false
	}

	return true

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
	return reg.FindString(price)
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
	for _, models := range bikeModels {
		for _, model := range models {
			if strings.Contains(strings.ToLower(title), strings.ToLower(model.Name)) {
				return model.Name
			}
		}
	}
	return "NoModelFound"
}
