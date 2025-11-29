package server

import (
	"encoding/base64"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/text/encoding/ianaindex"
	"golang.org/x/text/transform"
)

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

		if ctype == "text/plain" || ctype == "text/html" {
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

			var decodedBody []byte
			if err != nil {
				decodedBody = bodyBytes
			} else {
				decodedBody, _ = encoding.NewDecoder().Bytes(bodyBytes)
			}

			// Store in appropriate field
			switch ctype {
			case "text/html":
				if content.BodyHTML == "" {
					content.BodyHTML = string(decodedBody)
				}
				content.BodyType = "text/html"
			case "text/plain":
				if content.BodyText == "" {
					content.BodyText = string(decodedBody)
				}
				content.BodyType = "text/plain"
			}

			// Track if both text and HTML are available
			if content.BodyText != "" && content.BodyHTML != "" {
				content.HasAlternate = true
				// Prefer PlainText for primary display
				content.BodyType = "text/plain"
			}
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
