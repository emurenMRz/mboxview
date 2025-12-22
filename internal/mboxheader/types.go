package mboxheader

// ValidationResult represents the result of validating a message header
type ValidationResult struct {
	MsgIndex int    `json:"msgIndex"`
	Field    string `json:"field"`
	Status   string `json:"status"` // "valid", "missing", "invalid", "deleted"
	Detail   string `json:"detail,omitempty"`
}

// Message represents a parsed message with headers and body
type Message struct {
	EnvelopeLine  string
	Headers       string
	Body          string
	ParsedHeaders map[string]string
}
