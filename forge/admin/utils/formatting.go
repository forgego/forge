package utils

import (
	"fmt"
	"html/template"
	"time"
)

// FormatValue formats a value for display based on its type
func FormatValue(value interface{}) string {
	if value == nil {
		return "-"
	}
	
	switch v := value.(type) {
	case string:
		return v
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%.2f", v)
	case bool:
		if v {
			return "Yes"
		}
		return "No"
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	case []byte:
		return fmt.Sprintf("<binary data: %d bytes>", len(v))
	default:
		return fmt.Sprintf("%v", v)
	}
}

// FormatDate formats a date value
func FormatDate(value interface{}) string {
	if value == nil {
		return "-"
	}
	
	switch v := value.(type) {
	case time.Time:
		return v.Format("2006-01-02")
	case string:
		// Try to parse and format
		if t, err := time.Parse("2006-01-02", v); err == nil {
			return t.Format("2006-01-02")
		}
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// FormatDateTime formats a datetime value
func FormatDateTime(value interface{}) string {
	if value == nil {
		return "-"
	}
	
	switch v := value.(type) {
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	case string:
		// Try to parse and format
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t.Format("2006-01-02 15:04:05")
		}
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// FormatCurrency formats a number as currency
func FormatCurrency(value interface{}, currency string) string {
	if value == nil {
		return "-"
	}
	
	var amount float64
	switch v := value.(type) {
	case float64:
		amount = v
	case float32:
		amount = float64(v)
	case int:
		amount = float64(v)
	case int64:
		amount = float64(v)
	default:
		return fmt.Sprintf("%v", value)
	}
	
	if currency == "" {
		currency = "$"
	}
	
	return fmt.Sprintf("%s%.2f", currency, amount)
}

// FormatFileSize formats a file size in bytes
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Truncate truncates a string to a maximum length
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// SafeHTML returns HTML that is safe to render
func SafeHTML(s string) template.HTML {
	return template.HTML(s)
}

// Pluralize returns plural form of a word based on count
func Pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	if plural == "" {
		return singular + "s"
	}
	return plural
}
