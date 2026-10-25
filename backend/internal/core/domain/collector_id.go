package domain

const (
	// CollectorIDLength is the number of characters of a collector id.
	CollectorIDLength = 4
	// CollectorIDAlphabet lists the characters a collector id may consist of, 31 in total. It leaves out
	// 0, 1, i, l and o, which are easily mistaken for each other. Generators of ids must draw from it.
	CollectorIDAlphabet = "23456789abcdefghjkmnpqrstuvwxyz"
)

// ErrInvalidCollectorID is returned for an id that is not a well-formed collector id.
var ErrInvalidCollectorID = invalid("invalid collector id")

// CollectorID identifies a collector. It is short so it can be used in URLs.
type CollectorID string

// ParseCollectorID validates s and returns it as a CollectorID.
func ParseCollectorID(s string) (CollectorID, error) {
	if len(s) != CollectorIDLength {
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
	for i := range len(CollectorIDAlphabet) {
		if CollectorIDAlphabet[i] == c {
			return true
		}
	}

	return false
}

// String implements fmt.Stringer.
func (id CollectorID) String() string { return string(id) }
