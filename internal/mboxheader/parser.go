package mboxheader

import (
	"bufio"
	"strings"
)

type KeyFieldSet struct {
	index int    // field-index
	name  string // original field-name
}

type ParsedHeaderField struct {
	name   string   // original field-name
	values []string // folded lines
}

type ParsedMailHeaders struct {
	keys   map[string]KeyFieldSet // lowercased field-name set: keys[lowercased field-name] = KeyFieldSet
	fields []ParsedHeaderField    // Preserve header field order
}

func NewParsedMailHeaders(headers string) ParsedMailHeaders {
	keys := map[string]KeyFieldSet{}
	parsedFields := parseField(headers)

	for i, field := range parsedFields {
		key := strings.ToLower(field.name)
		if _, exists := keys[key]; !exists {
			keys[key] = KeyFieldSet{
				index: i,
				name:  field.name,
			}
		}
	}

	return ParsedMailHeaders{
		keys:   keys,
		fields: parsedFields,
	}
}

func parseField(headers string) (fields []ParsedHeaderField) {
	var currentField *ParsedHeaderField
	scanner := bufio.NewScanner(strings.NewReader(headers))

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this is a folded line (starts with SP or HT)
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			// This is a folded line
			if currentField != nil {
				// Append to the last field's values
				currentField.values = append(currentField.values, strings.TrimLeft(line, " \t"))
			}
		} else if i := strings.Index(line, ":"); i != -1 {
			// This is a new header line
			name := strings.TrimSpace(line[:i])
			value := strings.TrimSpace(line[i+1:])

			fields = append(fields, ParsedHeaderField{
				name:   name,
				values: []string{value},
			})
			currentField = &fields[len(fields)-1]
		} else {
			// Invalid header line, ignore
			currentField = nil
		}
	}

	return
}

func (h ParsedMailHeaders) GetFieldValue(key string) (string, bool) {
	keySet, exists := h.keys[key]
	if !exists {
		return "", false
	}
	index := keySet.index
	if index < 0 || index >= len(h.fields) {
		return "", false
	}
	return strings.TrimSpace(strings.Join(h.fields[index].values, " ")), true
}
