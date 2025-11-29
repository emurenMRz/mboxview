package server

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	"sort"
	"strconv"
	"time"

	"github.com/emersion/go-imap/utf7"
	"github.com/emersion/go-mbox"
	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
)

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
	Body        string   `json:"body"`
	BodyType    string   `json:"bodyType"`
	Attachments []string `json:"attachments"`
}

var basePath string
var editMode bool

// RegisterHandlers registers HTTP handlers for the server. Call this before ListenAndServe.
func RegisterHandlers(path string) {
	basePath = path

	http.HandleFunc("/api/mailboxes/", handleMailboxRoutes)

	// Serve static files under /static/ (API lives under /api/)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Ensure common MIME types are set (some platforms lack .css/.js by default)
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".js", "application/javascript")

	// Serve index at root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "index.html"))
	})
}

// ListenAndServe is a thin wrapper to allow main to call server.ListenAndServe
func ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, nil)
}

// SetEditMode enables or disables edit mode for the server.
func SetEditMode(v bool) {
	editMode = v
}

func handleMailboxRoutes(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Method + " " + r.URL.Path)

	// Guard POST methods for edit mode
	if r.Method == "POST" && !editMode {
		http.NotFound(w, r)
		return
	}

	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/mailboxes/"), "/")
	segmentCount := len(parts)

	if segmentCount == 1 {
		mailboxesHandler(w, r)
		return
	}

	if parts[1] == "emails" {
		mboxName := parts[0]
		switch segmentCount {
		case 2:
			listEmailsHandler(w, r, mboxName)
		case 3:
			emailContentHandler(w, r, mboxName, parts[2])
		case 4:
			if r.Method == "POST" && parts[3] == "read" {
				markEmailReadHandler(w, r, mboxName, parts[2])
			}
		}
		return
	}

	http.NotFound(w, r)
}

func markEmailReadHandler(w http.ResponseWriter, r *http.Request, mailboxName string, emailIdStr string) {
	// mailboxName coming from API is UTF-8; encode to IMAP-UTF7 to find file on disk
	encodedMailboxName, err := utf7.Encoding.NewEncoder().String(mailboxName)
	if err != nil {
		http.Error(w, "Invalid mailbox name", http.StatusBadRequest)
		return
	}
	mboxPath := filepath.Join(basePath, encodedMailboxName)

	// String of mail id to integer
	emailId, err := strconv.Atoi(emailIdStr)
	if err != nil {
		http.Error(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	// Read the mbox file and parse messages
	messages, ok := readMessages(mboxPath, w, r)
	if !ok {
		return
	}

	if emailId < 0 || emailId >= len(messages) {
		http.Error(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	// Update the target message
	targetMsgStr := messages[emailId]
	// Split envelope line from the rest
	envelopeLine, rest := splitAtFirstNewline(targetMsgStr)
	// Find header part (before body)
	headers, body := splitHeadersFromBody(rest)
	// Update status header
	newHeaders := updateStatusHeader(headers, "RO")
	// Reconstruct the message
	messages[emailId] = envelopeLine + "\n" + newHeaders + "\n" + body

	// Now rewrite the mbox file with the updated messages
	tempFile, err := os.CreateTemp(basePath, "mboxview-update-*.mbox")
	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		http.Error(w, "Error updating mbox", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Write each message to the temp file directly
	for _, msgStr := range messages {
		_, err := tempFile.WriteString(msgStr)
		if err != nil {
			log.Printf("Error writing message to temp file: %v", err)
			http.Error(w, "Error updating mbox", http.StatusInternalServerError)
			return
		}
	}

	// Close the temp file to flush
	if err := tempFile.Close(); err != nil {
		log.Printf("Error closing temp file: %v", err)
		http.Error(w, "Error updating mbox", http.StatusInternalServerError)
		return
	}

	// Atomically replace the original file
	if err := os.Rename(filepath.Clean(tempFile.Name()), mboxPath); err != nil {
		log.Printf("Error replacing original file: %v", err)
		http.Error(w, "Error updating mbox", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// readMessages reads all messages from an mbox file and returns them as a slice of strings.
// It opens the file in read-only mode.
func readMessages(mboxPath string, w http.ResponseWriter, r *http.Request) ([]string, bool) {
	f, err := os.Open(mboxPath)
	if err != nil {
		http.NotFound(w, r)
		return nil, false
	}
	defer f.Close()

	// Read all messages to find the one we want to update
	var messages []string
	scanner := bufio.NewScanner(f)
	var currentMessage strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "From ") {
			// End of previous message
			if currentMessage.Len() > 0 {
				messages = append(messages, currentMessage.String())
				currentMessage.Reset()
			}
			// Start of new message
			currentMessage.WriteString(line)
			currentMessage.WriteString("\n")
		} else {
			currentMessage.WriteString(line)
			currentMessage.WriteString("\n")
		}
	}
	// Add the last message
	if currentMessage.Len() > 0 {
		messages = append(messages, currentMessage.String())
	}
	return messages, true
}

func mailboxesHandler(w http.ResponseWriter, _ *http.Request) {
	files, err := os.ReadDir(basePath)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	var mailboxes []string
	for _, file := range files {
		if !file.IsDir() {
			// Files on disk are IMAP-UTF7 encoded; decode to UTF-8 for API response
			decodedName, err := utf7.Encoding.NewDecoder().String(file.Name())
			if err != nil {
				log.Printf("Failed to decode mailbox filename %s: %v", file.Name(), err)
				continue
			}
			mailboxes = append(mailboxes, decodedName)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mailboxes)
}

func listEmailsHandler(w http.ResponseWriter, r *http.Request, mailboxName string) {
	// mailboxName coming from API is UTF-8; encode to IMAP-UTF7 to find file on disk
	encodedMailboxName, err := utf7.Encoding.NewEncoder().String(mailboxName)
	if err != nil {
		http.Error(w, "Invalid mailbox name", http.StatusBadRequest)
		return
	}
	mboxPath := filepath.Join(basePath, encodedMailboxName)
	f, err := os.Open(mboxPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	var emails []Email
	reader := mbox.NewReader(f)
	i := 0
	for {
		// reader.NextMessage returns an io.Reader positioned at the start of a message
		mrReader, err := reader.NextMessage()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading message in %s: %v", mailboxName, err)
			continue
		}

		// Parse the message headers using mail.ReadMessage
		mr, err := mail.ReadMessage(mrReader)
		if err != nil {
			log.Printf("Failed to parse message headers in %s: %v", mailboxName, err)
			continue
		}
		header := mr.Header
		subject := header.Get("Subject")

		decoder := new(mime.WordDecoder)
		// Use a WordDecoder with a CharsetReader so encoded-words with non-UTF8
		// charsets (e.g. ISO-2022-JP) are converted to UTF-8.
		decoder = &mime.WordDecoder{CharsetReader: charsetReader}
		decodedSubject, err := decoder.DecodeHeader(subject)
		if err != nil {
			decodedSubject = subject
		}

		decodedFrom := decodeAddressList(header.Get("From"), decoder)

		// parse Date header into time for sorting
		dateStr := header.Get("Date")
		ts := parseDate(dateStr)

		status := header.Get("Status")
		if status == "" {
			// ヘッダが無い場合は新着扱い
			status = "N"
		}

		emails = append(emails, Email{
			ID:        i,
			From:      decodedFrom,
			Date:      dateStr,
			Subject:   decodedSubject,
			Status:    status,
			Timestamp: ts,
		})
		i++
	}

	// sort by Timestamp descending (newest first). Zero timestamps go last.
	sort.SliceStable(emails, func(a, b int) bool {
		ta := emails[a].Timestamp
		tb := emails[b].Timestamp
		if ta.Equal(tb) {
			return emails[a].ID < emails[b].ID
		}
		if ta.IsZero() {
			return false
		}
		if tb.IsZero() {
			return true
		}
		return ta.After(tb)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(emails)
}

func emailContentHandler(w http.ResponseWriter, r *http.Request, mailboxName string, emailIdStr string) {
	emailId, err := strconv.Atoi(emailIdStr)
	if err != nil {
		http.Error(w, "Invalid email ID", http.StatusBadRequest)
		return
	}

	// mailboxName coming from API is UTF-8; encode to IMAP-UTF7 to find file on disk
	encodedMailboxName, err := utf7.Encoding.NewEncoder().String(mailboxName)
	if err != nil {
		http.Error(w, "Invalid mailbox name", http.StatusBadRequest)
		return
	}
	mboxPath := filepath.Join(basePath, encodedMailboxName)
	f, err := os.Open(mboxPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	reader := mbox.NewReader(f)
	i := 0
	var selectedMsg *mail.Message
	for {
		mrReader, err := reader.NextMessage()
		if err == io.EOF {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Printf("Error reading message in %s: %v", mailboxName, err)
			http.Error(w, "Error reading mbox", http.StatusInternalServerError)
			return
		}

		mr, err := mail.ReadMessage(mrReader)
		if err != nil {
			log.Printf("Failed to parse message in %s: %v", mailboxName, err)
			// skip this message but continue
			i++
			continue
		}

		if i == emailId {
			selectedMsg = mr
			break
		}
		i++
	}

	if selectedMsg == nil {
		http.NotFound(w, r)
		return
	}

	content := parseMessageBody(selectedMsg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(content)
}

func parseMessageBody(msg *mail.Message) EmailContent {
	var content EmailContent
	content.Attachments = []string{}

	// recursive entity processor
	var processEntity func(header interface{ Get(string) string }, body io.Reader)
	processEntity = func(header interface{ Get(string) string }, body io.Reader) {
		ctype, params, err := mime.ParseMediaType(header.Get("Content-Type"))
		if err != nil {
			ctype = "text/plain"
		}

		// handle multipart recursively
		if strings.HasPrefix(ctype, "multipart/") {
			mr := multipart.NewReader(body, params["boundary"])
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Printf("Error reading multipart body: %v", err)
					break
				}
				processEntity(p.Header, p)
			}
			return
		}

		// handle attachments
		if disp, dispParams, err := mime.ParseMediaType(header.Get("Content-Disposition")); err == nil && disp == "attachment" {
			content.Attachments = append(content.Attachments, dispParams["filename"])
			return
		}

		// handle content-transfer-encoding
		cte := strings.ToLower(strings.TrimSpace(header.Get("Content-Transfer-Encoding")))
		reader := body
		switch cte {
		case "base64":
			reader = base64.NewDecoder(base64.StdEncoding, body)
		case "quoted-printable":
			reader = quotedprintable.NewReader(body)
		default:
			// 7bit, 8bit, binary -> no wrapper
		}

		if content.Body == "" && (ctype == "text/plain" || ctype == "text/html") {
			bodyBytes, err := io.ReadAll(reader)
			if err != nil {
				return
			}

			charset := params["charset"]
			if charset == "" {
				charset = "utf-8"
			}

			encoding, err := ianaindex.IANA.Encoding(charset)
			if err != nil || encoding == nil {
				encoding, _ = ianaindex.IANA.Encoding("utf-8")
			}

			decodedBody, err := encoding.NewDecoder().Bytes(bodyBytes)
			if err != nil {
				content.Body = string(bodyBytes)
			} else {
				content.Body = string(decodedBody)
			}
			content.BodyType = ctype
		}
	}

	processEntity(msg.Header, msg.Body)
	return content
}

// parseDate tries to parse common email Date header formats and returns a time.Time.
// If parsing fails, it returns zero time.
func parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}
	if t, err := mail.ParseDate(dateStr); err == nil {
		return t
	}
	// common fallbacks
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC850,
		time.RFC3339,
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, dateStr); err == nil {
			return t
		}
	}
	return time.Time{}
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	if charset == "" {
		return input, nil
	}
	enc, err := ianaindex.IANA.Encoding(strings.ToLower(charset))
	if err != nil || enc == nil {
		return input, nil
	}
	return transform.NewReader(input, enc.NewDecoder()), nil
}

func splitAtFirstNewline(s string) (string, string) {
	if i := strings.Index(s, "\n"); i != -1 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

func splitHeadersFromBody(s string) (string, string) {
	// Find the first empty line that separates headers from body
	if i := strings.Index(s, "\n\n"); i != -1 {
		return s[:i+1], s[i+2:]
	}
	return s, ""
}

func updateStatusHeader(headers string, newStatus string) string {
	// Find existing Status header
	statusStart := strings.Index(headers, "Status: ")
	if statusStart != -1 {
		// Find end of this line
		statusEnd := strings.Index(headers[statusStart:], "\n")
		if statusEnd != -1 {
			statusEnd += statusStart
		} else {
			statusEnd = len(headers)
		}
		// Replace the line
		return headers[:statusStart] + "Status: " + newStatus + "\n" + headers[statusEnd:]
	}

	// No newline, just append
	if headers[len(headers)-1] != '\n' {
		headers += "\n"
	}

	return headers + "Status: " + newStatus + "\n"
}

func decodeAddressList(header string, decoder *mime.WordDecoder) string {
	if header == "" {
		return ""
	}
	addrs, err := mail.ParseAddressList(header)
	if err != nil {
		// Fallback: try to decode the whole header as an encoded-word
		if dec, e := decoder.DecodeHeader(header); e == nil {
			return dec
		}
		return header
	}
	var parts []string
	for _, a := range addrs {
		name := a.Name
		if name != "" {
			if dec, e := decoder.DecodeHeader(name); e == nil {
				name = dec
			}
			parts = append(parts, name+" <"+a.Address+">")
		} else {
			parts = append(parts, a.Address)
		}
	}
	return strings.Join(parts, ", ")
}
