package server

import "time"

type Email struct {
    ID      int       `json:"id"`
    From    string    `json:"from"`
    Date    string    `json:"date"`
    Subject string    `json:"subject"`
    Status  string    `json:"status"`
    // Timestamp is parsed Date used for sorting. Not exported to JSON.
    Timestamp time.Time `json:"-"`
}

type EmailContent struct {
    Body        string   `json:"body"`
    BodyType    string   `json:"bodyType"`
    Attachments []string `json:"attachments"`
}
