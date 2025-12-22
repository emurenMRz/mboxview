package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/emurenMRz/mboxview/internal/mboxheader"
	"github.com/emurenMRz/mboxview/internal/server"
)

func main() {
	var (
		mode          = flag.String("mode", "validate", "Operation mode: validate, fix, show")
		inplace       = flag.Bool("inplace", false, "Modify input file in-place (for fix mode)")
		outPath       = flag.String("out", "", "Output file path (for fix mode)")
		dryRun        = flag.Bool("dry-run", false, "Simulate fix operation without writing (for fix mode)")
		removeDeleted = flag.Bool("remove-deleted", false, "Remove messages with Status: D (for fix mode)")
		quiet         = flag.Bool("quiet", false, "Suppress non-error output (for fix mode)")
		msgIndex      = flag.Int("msg", -1, "Message index (for show mode)")
		inputPath     = flag.String("path", "", "Input mbox file path (required)")
	)
	flag.Parse()

	if *inputPath == "" {
		log.Fatal("Error: -path is required")
	}

	// Read messages from mbox file
	messages, ok := server.ReadMessages(*inputPath, nil, nil)
	if !ok {
		log.Fatal("Failed to read mbox file")
	}

	// Process messages based on mode
	switch *mode {
	case "validate":
		validateMessages(messages)
	case "fix":
		fixMessages(messages, *inputPath, *inplace, *outPath, *dryRun, *removeDeleted, *quiet)
	case "show":
		showMessage(messages, *msgIndex)
	default:
		log.Fatal("Error: Unknown mode. Use validate, fix, or show")
	}
}

func validateMessages(messages []string) {
	var allResults []mboxheader.ValidationResult

	for i, message := range messages {
		// Split headers and body
		_, rest := server.SplitAtFirstNewline(message)
		headers, _ := server.SplitHeadersFromBody(rest)

		// Validate headers
		results := mboxheader.ValidateHeaders(headers, i)
		allResults = append(allResults, results...)
	}

	// Output results
	outputText(allResults)
}

func fixMessages(messages []string, inputPath string, inplace bool, outPath string, dryRun, removeDeleted, quiet bool) {
	// Filter out deleted messages if requested
	if removeDeleted {
		var filteredMessages []string
		for _, message := range messages {
			// Split headers and body
			_, rest := server.SplitAtFirstNewline(message)
			headers, _ := server.SplitHeadersFromBody(rest)

			// Check for Status: D
			parsedHeaders := mboxheader.NewParsedMailHeaders(headers)
			if status, exists := parsedHeaders.GetFieldValue("status"); exists && status == "D" {
				// Skip this message
				continue
			}
			filteredMessages = append(filteredMessages, message)
		}
		messages = filteredMessages
	}

	// Normalize messages
	var normalizedMessages []string
	var allResults []mboxheader.ValidationResult

	for i, message := range messages {
		// Split headers and body
		envelopeLine, rest := server.SplitAtFirstNewline(message)
		headers, body := server.SplitHeadersFromBody(rest)

		// Normalize headers
		normalizedMessage, results := mboxheader.NormalizeHeaders(headers, i)
		normalizedMessages = append(normalizedMessages, envelopeLine+"\n"+normalizedMessage+"\n"+body)
		allResults = append(allResults, results...)
	}

	// Output results
	if !quiet {
		outputText(allResults)
	}

	// Write to file if not dry-run
	if !dryRun {
		if inplace {
			// Write back to input file
			writeMessagesToFile(normalizedMessages, inputPath)
		} else if outPath != "" {
			// Write to output file
			writeMessagesToFile(normalizedMessages, outPath)
		} else {
			// Write to stdout
			for _, msg := range normalizedMessages {
				fmt.Println(msg)
			}
		}
	}
}

func showMessage(messages []string, msgIndex int) {
	if msgIndex < 0 || msgIndex >= len(messages) {
		log.Fatal("Error: Invalid message index")
	}

	message := messages[msgIndex]
	_, rest := server.SplitAtFirstNewline(message)
	headers, _ := server.SplitHeadersFromBody(rest)

	fmt.Printf("Message %d:\n", msgIndex)
	fmt.Println(mboxheader.NewParsedMailHeaders(headers))
}

func writeMessagesToFile(messages []string, path string) {
	file, err := os.Create(path)
	if err != nil {
		log.Fatal("Error creating output file:", err)
	}
	defer file.Close()

	for _, msg := range messages {
		_, err := file.WriteString(msg)
		if err != nil {
			log.Fatal("Error writing to output file:", err)
		}
	}
}

func outputText(results []mboxheader.ValidationResult) {
	if len(results) == 0 {
		fmt.Println("No validation errors found.")
		return
	}

	for _, result := range results {
		switch result.Status {
		case mboxheader.StatusMissing:
			fmt.Printf("Message %d: %s header is missing\n", result.MsgIndex, result.Field)
		case mboxheader.StatusInvalid:
			fmt.Printf("Message %d: %s header is invalid (%s)\n", result.MsgIndex, result.Field, result.Detail)
		case mboxheader.StatusDeleted:
			fmt.Printf("Message %d: Status = D (will be removed)\n", result.MsgIndex)
		}
	}
}
