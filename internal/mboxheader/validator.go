package mboxheader

import (
	"net/mail"
	"regexp"
	"strings"
)

const (
	StatusValid   = "valid"
	StatusMissing = "missing"
	StatusInvalid = "invalid"
	StatusDeleted = "deleted"
)

var (
	// Required headers for RFC 5322 compliance
	requiredHeaders = []string{"from", "date", "message-id"}

	// Regular expression for valid message ID format
	messageIDRegex = regexp.MustCompile(`^<[^<>@]+@[^<>@]+>$`)
)

// ValidateHeaders validates a message's headers against RFC 5322
func ValidateHeaders(headers string, msgIndex int) []ValidationResult {
	var results []ValidationResult

	// Parse headers into key-value pairs
	parsedHeaders := NewParsedMailHeaders(headers)

	// Check for required headers
	for _, headerName := range requiredHeaders {
		if _, exists := parsedHeaders.keys[headerName]; !exists {
			results = append(results, ValidationResult{
				MsgIndex: msgIndex,
				Field:    headerName,
				Status:   StatusMissing,
			})
		}
	}

	// Validate specific header fields
	if from, exists := parsedHeaders.GetFieldValue("from"); exists {
		if !isValidFrom(from) {
			results = append(results, ValidationResult{
				MsgIndex: msgIndex,
				Field:    "From",
				Status:   StatusInvalid,
				Detail:   "Invalid From address format",
			})
		}
	}

	if date, exists := parsedHeaders.GetFieldValue("date"); exists {
		if !isValidDate(date) {
			results = append(results, ValidationResult{
				MsgIndex: msgIndex,
				Field:    "Date",
				Status:   StatusInvalid,
				Detail:   "Invalid Date format",
			})
		}
	}

	if msgID, exists := parsedHeaders.GetFieldValue("message-id"); exists {
		if !isValidMessageID(msgID) {
			results = append(results, ValidationResult{
				MsgIndex: msgIndex,
				Field:    "Message-ID",
				Status:   StatusInvalid,
				Detail:   "Invalid Message-ID format",
			})
		}
	}

	// Check for Status: D
	if status, exists := parsedHeaders.GetFieldValue("status"); exists {
		if status == "D" {
			results = append(results, ValidationResult{
				MsgIndex: msgIndex,
				Field:    "Status",
				Status:   StatusDeleted,
			})
		}
	}

	return results
}

// isValidFrom checks if a From header is valid
func isValidFrom(from string) bool {
	// Basic check: should be parseable by net/mail
	_, err := mail.ParseAddressList(from)
	return err == nil
}

// isValidDate checks if a Date header is valid
func isValidDate(date string) bool {
	// Basic check: should be parseable by net/mail
	_, err := mail.ParseDate(date)
	return err == nil
}

// isValidMessageID checks if a Message-ID header is valid
func isValidMessageID(msgID string) bool {
	// Remove surrounding <>
	msgID = strings.Trim(msgID, "<>")
	return messageIDRegex.MatchString("<" + msgID + ">")
}
