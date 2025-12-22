package mboxheader

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// makeUUID generates a random UUID version 4 (RFC 4122)
func makeUUID() string {
	uuid := make([]byte, 16)
	_, _ = rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant RFC 4122
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

// makeUUIDv7 generates a UUID version 7 (draft RFC 9562) based on a timestamp
func makeUUIDv7(timestamp time.Time) string {
	uuid := make([]byte, 16)
	_, _ = rand.Read(uuid)

	// Set timestamp in big-endian format (bytes 0-5)
	ms := timestamp.UnixMilli()
	binary.BigEndian.PutUint64(uuid[0:8], uint64(ms)<<16) // 48 bits of timestamp

	// Version 7 (0111)
	uuid[6] = (uuid[6] & 0x0f) | 0x70

	// Variant RFC 4122
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
