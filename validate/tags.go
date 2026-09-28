package validation

import "strings"

// Common validation tags for framework use
const (
	TagRequired      = "required"
	TagEmail         = "email"
	TagMin           = "min"
	TagMax           = "max"
	TagLen           = "len"
	TagNumeric       = "numeric"
	TagAlpha         = "alpha"
	TagAlphanum      = "alphanum"
	TagURL           = "url"
	TagUUID          = "uuid"
	TagGTE           = "gte" // Greater than or equal
	TagLTE           = "lte" // Less than or equal
	TagGT            = "gt"  // Greater than
	TagLT            = "lt"  // Less than
	TagOneOf         = "oneof"
	TagSlug          = "slug"
	TagPhone         = "phone"
	TagIP            = "ip"
	TagIPv4          = "ipv4"
	TagIPv6          = "ipv6"
	TagMAC           = "mac"
	TagJSON          = "json"
	TagBase64        = "base64"
	TagHostname      = "hostname"
	TagFQDN          = "fqdn"
	TagURI           = "uri"
	TagChoices       = "choices"
	TagDecimalDigits = "decimal_max_digits"
	TagDecimalPlaces = "decimal_places"
)

// FieldTags maps field types to default validation tags
var FieldTags = map[string]string{
	"email":    "required,email",
	"url":      "required,url",
	"uuid":     "required,uuid",
	"string":   "required",
	"int":      "required,numeric",
	"int32":    "required,numeric",
	"int64":    "required,numeric",
	"float32":  "required,numeric",
	"float64":  "required,numeric",
	"decimal":  "required,numeric",
	"bool":     "boolean",
	"time":     "datetime",
	"date":     "date",
	"datetime": "datetime",
	"json":     "json",
	"bytes":    "base64",
}

// GetValidationTags returns validation tags for a field type
func GetValidationTags(fieldType string, required bool) string {
	tags := FieldTags[fieldType]
	if !required && tags != "" {
		// Remove "required" from tags if field is optional
		tags = removeTag(tags, TagRequired)
	}
	return tags
}

// removeTag removes a tag from a tag string
func removeTag(tags, tag string) string {
	if tags == "" {
		return ""
	}

	// Split tags by comma
	tagList := strings.Split(tags, ",")
	var result []string

	for _, t := range tagList {
		t = strings.TrimSpace(t)
		// Check if this tag matches (could be "tag" or "tag=value")
		if t != tag && !strings.HasPrefix(t, tag+"=") {
			result = append(result, t)
		}
	}

	return strings.Join(result, ",")
}

// AddTag adds a tag to a tag string
func AddTag(tags, tag string) string {
	if tags == "" {
		return tag
	}
	return tags + "," + tag
}

// BuildTagString builds a validation tag string from multiple tags
func BuildTagString(tags ...string) string {
	var validTags []string
	for _, tag := range tags {
		if tag != "" {
			validTags = append(validTags, tag)
		}
	}
	return strings.Join(validTags, ",")
}
