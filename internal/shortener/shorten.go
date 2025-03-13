package shortener

import (
	"errors"
	"slices"
	"strings"
)

const (
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// EncodeBase62 converts an integer ID to a Base62 string.
func EncodeBase62(num uint64) string {
	if num == 0 {
		return "0"
	}

	var sb strings.Builder
	for num > 0 {
		rem := num % 62
		sb.WriteByte(base62Chars[rem])
		num /= 62
	}

	// Reverse the string (since we build it backwards)
	encoded := sb.String()
	runes := []rune(encoded)
	slices.Reverse(runes)
	return string(runes)
}

// DecodeBase62 decodes a Base62 string back to a number
func DecodeBase62(s string) (uint64, error) {
	var decoded uint64
	for _, r := range s {
		pos := strings.IndexRune(base62Chars, r)
		if pos == -1 {
			return 0, errors.New("invalid character in base62 string")
		}
		decoded = decoded*62 + uint64(pos)
	}
	return decoded, nil
}
