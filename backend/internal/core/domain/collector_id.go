package domain

import (
	"crypto/rand"
	"fmt"
)

const (
	collectorIDLength = 4
	// collectorIDAlphabet has 31 characters. Random bytes are mapped onto it by rejection
	// sampling (see NewCollectorID) to avoid modulo bias.
	// It leaves out 0, 1, i, l and o, which are easily mistaken for each other.
	collectorIDAlphabet = "23456789abcdefghjkmnpqrstuvwxyz"
)

// ErrInvalidCollectorID is returned for an id that is not a well-formed collector id.
var ErrInvalidCollectorID = invalid("invalid collector id")

// CollectorID identifies a collector. It is short so it can be used in URLs.
type CollectorID string

// NewCollectorID returns a random collector id. It is not guaranteed to be unused.
func NewCollectorID() (CollectorID, error) {
	// Bytes at or above this limit are rejected, so the remaining range is a multiple of the
	// alphabet size and every symbol is equally likely.
	const limit = 256 - 256%len(collectorIDAlphabet)

	id := make([]byte, 0, collectorIDLength)
	buf := make([]byte, collectorIDLength*2)

	for len(id) < collectorIDLength {
		if _, err := rand.Read(buf); err != nil {
			return "", fmt.Errorf("generate collector id: %w", err)
		}

		for _, b := range buf {
			if int(b) >= limit {
				continue
			}

			id = append(id, collectorIDAlphabet[int(b)%len(collectorIDAlphabet)])
			if len(id) == collectorIDLength {
				break
			}
		}
	}

	return CollectorID(id), nil
}

// ParseCollectorID validates s and returns it as a CollectorID.
func ParseCollectorID(s string) (CollectorID, error) {
	if len(s) != collectorIDLength {
		return "", ErrInvalidCollectorID
	}

	for i := range len(s) {
		if !isCollectorIDChar(s[i]) {
			return "", ErrInvalidCollectorID
		}
	}

	return CollectorID(s), nil
}

func isCollectorIDChar(c byte) bool {
	for i := range len(collectorIDAlphabet) {
		if collectorIDAlphabet[i] == c {
			return true
		}
	}

	return false
}

// String implements fmt.Stringer.
func (id CollectorID) String() string { return string(id) }
