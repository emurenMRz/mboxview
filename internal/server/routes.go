package server

import (
	"log"
	"net/http"
	"strings"
)

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
