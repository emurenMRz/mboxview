package server

import "time"

type Email struct {
	ID      int    `json:"id"`
	From    string `json:"from"`
	Date    string `json:"date"`
	Subject string `json:"subject"`
	Status  string `json:"status"`
	// Timestamp is parsed Date used for sorting. Not exported to JSON.
	Timestamp time.Time `json:"-"`
}

type EmailContent struct {
	BodyText     string   `json:"bodyText"`     // Plain text version
	BodyHTML     string   `json:"bodyHTML"`     // HTML version
	BodyType     string   `json:"bodyType"`     // Primary body type (text/plain or text/html)
	HasAlternate bool     `json:"hasAlternate"` // Whether both text and HTML are available
	Attachments  []string `json:"attachments"`
}
