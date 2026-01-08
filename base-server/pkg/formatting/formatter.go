package formatting

import (
	"fmt"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Formatter provides formatting utilities based on organization settings
type Formatter struct {
	Locale       string
	DateFormat   string
	TimeFormat   string
	AmountFormat string
}

// NewFormatter creates a new formatter with the given settings
func NewFormatter(locale, dateFormat, timeFormat, amountFormat string) *Formatter {
	// Set defaults if not provided
	if locale == "" {
		locale = "de-DE"
	}
	if dateFormat == "" {
		dateFormat = "02.01.2006"
	}
	if timeFormat == "" {
		timeFormat = "15:04"
	}
	if amountFormat == "" {
		amountFormat = "de"
	}

	return &Formatter{
		Locale:       locale,
		DateFormat:   dateFormat,
		TimeFormat:   timeFormat,
		AmountFormat: amountFormat,
	}
}

// FormatDate formats a date according to the configured format
func (f *Formatter) FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(f.DateFormat)
}

// FormatTime formats a time according to the configured format
func (f *Formatter) FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(f.TimeFormat)
}

// FormatDateTime formats a date and time according to the configured formats
func (f *Formatter) FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(f.DateFormat + " " + f.TimeFormat)
}

// FormatAmount formats a monetary amount according to the configured locale
func (f *Formatter) FormatAmount(amount float64) string {
	// Parse the locale
	tag, err := language.Parse(f.Locale)
	if err != nil {
		// Fallback to German locale if parsing fails
		tag = language.German
	}

	// Create a printer for the locale
	p := message.NewPrinter(tag)

	// Format based on amount_format setting
	switch f.AmountFormat {
	case "en", "en-US", "en-GB":
		// English format: 1,234.56
		return p.Sprintf("%.2f", amount)
	case "de", "de-DE", "de-AT", "de-CH":
		// German format: 1.234,56
		return formatGermanAmount(amount)
	default:
		// Default to locale-based formatting
		return p.Sprintf("%.2f", amount)
	}
}

// formatGermanAmount formats an amount in German style (1.234,56)
func formatGermanAmount(amount float64) string {
	// Convert to string with 2 decimal places
	str := fmt.Sprintf("%.2f", amount)

	// Find the decimal point
	var intPart, decPart string
	for i, c := range str {
		if c == '.' {
			intPart = str[:i]
			decPart = str[i+1:]
			break
		}
	}

	// Add thousand separators to integer part
	result := ""
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result += "."
		}
		result += string(c)
	}

	// Combine with decimal part using comma
	return result + "," + decPart
}

// GetSupportedDateFormats returns a map of date format examples
func GetSupportedDateFormats() map[string]string {
	now := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	return map[string]string{
		"02.01.2006":    now.Format("02.01.2006"),    // 31.12.2024 (German)
		"01/02/2006":    now.Format("01/02/2006"),    // 12/31/2024 (US)
		"02/01/2006":    now.Format("02/01/2006"),    // 31/12/2024 (UK)
		"2006-01-02":    now.Format("2006-01-02"),    // 2024-12-31 (ISO)
		"Jan 02, 2006":  now.Format("Jan 02, 2006"),  // Dec 31, 2024
		"02 Jan 2006":   now.Format("02 Jan 2006"),   // 31 Dec 2024
		"Monday, 02.01": now.Format("Monday, 02.01"), // Tuesday, 31.12
	}
}

// GetSupportedTimeFormats returns a map of time format examples
func GetSupportedTimeFormats() map[string]string {
	now := time.Date(2024, 12, 31, 14, 30, 0, 0, time.UTC)
	return map[string]string{
		"15:04":       now.Format("15:04"),       // 14:30 (24h)
		"15:04:05":    now.Format("15:04:05"),    // 14:30:00 (24h with seconds)
		"3:04 PM":     now.Format("3:04 PM"),     // 2:30 PM (12h)
		"03:04 PM":    now.Format("03:04 PM"),    // 02:30 PM (12h with leading zero)
		"3:04:05 PM":  now.Format("3:04:05 PM"),  // 2:30:00 PM (12h with seconds)
		"03:04:05 PM": now.Format("03:04:05 PM"), // 02:30:00 PM (12h with leading zero and seconds)
	}
}

// GetSupportedAmountFormats returns a map of amount format examples
func GetSupportedAmountFormats() map[string]string {
	return map[string]string{
		"de":    "1.234,56", // German format
		"de-DE": "1.234,56", // German (Germany)
		"de-AT": "1.234,56", // German (Austria)
		"de-CH": "1'234,56", // German (Switzerland) - simplified to same as DE for now
		"en":    "1,234.56", // English format
		"en-US": "1,234.56", // English (US)
		"en-GB": "1,234.56", // English (UK)
	}
}

// GetSupportedLocales returns a list of supported locales
func GetSupportedLocales() []string {
	return []string{
		"de-DE", // German (Germany)
		"de-AT", // German (Austria)
		"de-CH", // German (Switzerland)
		"en-US", // English (US)
		"en-GB", // English (UK)
		"fr-FR", // French
		"es-ES", // Spanish
		"it-IT", // Italian
	}
}
