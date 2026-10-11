package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
)

const (
	collectorIDLength = 4
	// collectorIDAlphabet has 32 characters, so one random byte maps onto it without bias.
	// It leaves out 0, 1, i, l and o, which are easily mistaken for each other.
	collectorIDAlphabet = "23456789abcdefghjkmnpqrstuvwxyz"
)

// ErrInvalidCollectorID is returned for an id that is not a well-formed collector id.
var ErrInvalidCollectorID = errors.New("invalid collector id")

// CollectorID identifies a collector. It is short so it can be used in URLs.
type CollectorID string

// NewCollectorID returns a random collector id. It is not guaranteed to be unused.
func NewCollectorID() (CollectorID, error) {
	buf := make([]byte, collectorIDLength)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate collector id: %w", err)
	}

	for i, b := range buf {
		buf[i] = collectorIDAlphabet[int(b)%len(collectorIDAlphabet)]
	}

	return CollectorID(buf), nil
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
