package server

import (
	"bufio"
	"net/http"
	"os"
	"strings"
)

// ReadMessages reads all messages from an mbox file and returns them as a slice of strings.
// It opens the file in read-only mode.
func ReadMessages(mboxPath string, w http.ResponseWriter, r *http.Request) ([]string, bool) {
	f, err := os.Open(mboxPath)
	if err != nil {
		http.NotFound(w, r)
		return nil, false
	}
	defer f.Close()

	var messages []string
	scanner := bufio.NewScanner(f)
	var currentMessage strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "From ") && currentMessage.Len() > 0 {
			messages = append(messages, currentMessage.String())
			currentMessage.Reset()
		}
		currentMessage.WriteString(line)
		currentMessage.WriteString("\n")
	}
	if currentMessage.Len() > 0 {
		messages = append(messages, currentMessage.String())
	}
	return messages, true
}

func SplitAtFirstNewline(s string) (string, string) {
	if i := strings.Index(s, "\n"); i != -1 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

func SplitHeadersFromBody(s string) (string, string) {
	if i := strings.Index(s, "\n\n"); i != -1 {
		return s[:i+1], s[i+2:]
	}
	return s, ""
}

func updateStatusHeader(headers string, newStatus string) (string, bool) {
	searchKey := "Status: "
	searchIndex := 0

	for {
		index := strings.Index(headers[searchIndex:], searchKey)
		if index == -1 {
			break
		}
		actualIndex := searchIndex + index
		isLineStart := actualIndex == 0 || headers[actualIndex-1] == '\n'
		if !isLineStart {
			searchIndex = actualIndex + 1
			continue
		}

		statusEnd := strings.Index(headers[actualIndex:], "\n")
		if statusEnd == -1 {
			statusEnd = len(headers)
		} else {
			statusEnd += actualIndex
		}

		currentStatus := strings.TrimSpace(headers[actualIndex+len(searchKey) : statusEnd])
		if currentStatus == newStatus {
			return headers, false
		}

		if statusEnd < len(headers) && headers[statusEnd] == '\n' {
			statusEnd++
		}
		return headers[:actualIndex] + searchKey + newStatus + "\n" + headers[statusEnd:], true
	}

	if len(headers) > 0 && headers[len(headers)-1] != '\n' {
		headers += "\n"
	}
	return headers + searchKey + newStatus + "\n", true
}
