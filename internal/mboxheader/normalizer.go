package mboxheader

import (
	"fmt"
	"net/mail"
	"strings"
	"time"
)

const maxHeaderLineLength = 78

var requiredFieldNames = [...]string{"From", "Date", "Message-ID", "Subject", "To"}

// NormalizeHeaders normalizes a message's headers to RFC 5322 compliance
func NormalizeHeaders(headers string, msgIndex int) (string, []ValidationResult) {
	var results []ValidationResult

	// Parse headers into key-value pairs
	parsedHeaders := NewParsedMailHeaders(headers)

	// Check for required headers and add missing ones
	if _, exists := parsedHeaders.keys["from"]; !exists {
		results = append(results, ValidationResult{
			MsgIndex: msgIndex,
			Field:    "From",
			Status:   StatusMissing,
		})
	}

	if _, exists := parsedHeaders.keys["date"]; !exists {
		results = append(results, ValidationResult{
			MsgIndex: msgIndex,
			Field:    "Date",
			Status:   StatusMissing,
		})
	}

	if _, exists := parsedHeaders.keys["message-id"]; !exists {
		results = append(results, ValidationResult{
			MsgIndex: msgIndex,
			Field:    "Message-ID",
			Status:   StatusMissing,
		})
		// Add a default Message-ID
		uuid := makeUUIDByDateField(parsedHeaders)
		parsedHeaders.fields = append(parsedHeaders.fields, ParsedHeaderField{
			name:   "Message-ID",
			values: []string{fmt.Sprintf("<%s@%s>", uuid, "mboxfix")},
		})
	}

	// Rebuild headers string
	foldedHeaders := rebuildHeader(parsedHeaders)

	return foldedHeaders, results
}

func makeUUIDByDateField(h ParsedMailHeaders) string {
	timestamp := time.Time{}

	date, exists := h.GetFieldValue("date")
	if exists {
		var err error
		timestamp, err = mail.ParseDate(date)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	if timestamp.IsZero() {
		// Fallback to UUIDv4 if date is missing or parsing failed
		return makeUUID()
	}

	// Generate UUIDv7 based on timestamp
	return makeUUIDv7(timestamp)
}

func rebuildHeader(h ParsedMailHeaders) string {
	var folded strings.Builder

	for _, field := range h.fields {
		count := len(field.values)
		if count == 0 {
			fmt.Println("no field-value: " + field.name)
			continue
		}

		folded.WriteString(field.name + ": " + field.values[0] + "\n")
		if count == 1 {
			continue
		}

		for _, value := range field.values[1:] {
			folded.WriteString("\t" + value + "\n")
		}
	}

	return folded.String()
}
