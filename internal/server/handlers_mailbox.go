package server

import (
	"encoding/json"
	"io"
	"log"
	"mime"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/emersion/go-imap/utf7"
	"github.com/emersion/go-mbox"
)

func updateStatusHandler(w http.ResponseWriter, r *http.Request, mailboxName string, emailIdStr string, status string) {
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
	messages, ok := ReadMessages(mboxPath, w, r)
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
	envelopeLine, rest := SplitAtFirstNewline(targetMsgStr)
	// Find header part (before body)
	headers, body := SplitHeadersFromBody(rest)
	// Update status header
	newHeaders, updated := updateStatusHeader(headers, status)
	if !updated {
		w.WriteHeader(http.StatusOK)
		return
	}

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

func markEmailReadHandler(w http.ResponseWriter, r *http.Request, mailboxName string, emailIdStr string) {
	updateStatusHandler(w, r, mailboxName, emailIdStr, "RO")
}

func deleteEmailHandler(w http.ResponseWriter, r *http.Request, mailboxName string, emailIdStr string) {
	updateStatusHandler(w, r, mailboxName, emailIdStr, "D")
}

// readMessages is provided in helpers_mbox.go

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
			i++
			continue
		}

		// Parse the message headers using mail.ReadMessage
		mr, err := mail.ReadMessage(mrReader)
		if err != nil {
			log.Printf("Failed to parse message headers in %s: %v", mailboxName, err)
			i++
			continue
		}

		header := mr.Header
		status := header.Get("Status")
		if status == "D" {
			i++
			continue
		}
		if status == "" {
			// ヘッダが無い場合は新着扱い
			status = "N"
		}

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
