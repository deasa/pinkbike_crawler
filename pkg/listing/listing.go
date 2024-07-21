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
	Title, Year, Manufacturer, Price, Currency, Condition, FrameSize, WheelSize, FrameMaterial, FrontTravel, RearTravel string
}

func (l RawListing) Print() string {
	return fmt.Sprintf("Title: %s\nPrice: %s\n\tCondition: %s\n\tFrame Size: %s\n\tWheel Size: %s\n\tFront Travel: %s\n\tRear Travel: %s\n\tFrame Material: %s\n\t\n",
		l.Title, l.Price, l.Condition, l.FrameSize, l.WheelSize, l.FrontTravel, l.RearTravel, l.FrameMaterial)
}

func (l RawListing) PostProcess(exchangeRate float64) Listing {
	newL := Listing{
		Title: l.Title,
		Year: extractYear(l.Title),
		Manufacturer: extractManufacturer(l.Title),
		Currency: extractCurrency(l.Price),
		Price: convertPrice(l.Price, extractCurrency(l.Price), exchangeRate),
		Condition: l.Condition,
		FrameSize: l.FrameSize,
		WheelSize: l.WheelSize, //todo: convert to float - remove 650B
		FrontTravel: l.FrontTravel, //todo: remove mm
		RearTravel: l.RearTravel, //todo: remove mm
		FrameMaterial: l.FrameMaterial,
	}

	return newL
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
        floatPrice = math.Round(floatPrice*exchangeRate)
        p = fmt.Sprintf("%.0f", floatPrice)
    }

    return p
}

func extractPrice(price string) string {
	reg := regexp.MustCompile(`[0-9,]+`)
	return reg.FindString(price)
}

func extractManufacturer(title string) string {
	for _, manufacturer := range knownManufacturers {
		if strings.Contains(strings.ToLower(title), strings.ToLower(manufacturer)) {
			return manufacturer
		}
	}
	return ""
}
