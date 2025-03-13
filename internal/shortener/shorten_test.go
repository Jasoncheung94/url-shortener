package shortener

import (
	"testing"
)

func TestBase62EncodeDecode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{61, "Z"},
		{62, "10"},
		{12345, "3d7"},
		{987654321, "14Q60p"},
	}

	for _, test := range tests {
		// Test encoding
		encoded := EncodeBase62(test.input)
		if encoded != test.expected {
			t.Errorf("Base62Encode failed for %d: expected %s, got %s", test.input, test.expected, encoded)
		}

		// Test decoding
		decoded, err := DecodeBase62(encoded)
		if err != nil {
			t.Errorf("Base62Decode failed for %s: %v", encoded, err)
		}
		if decoded != test.input {
			t.Errorf("Base62Decode failed for %s: expected %d, got %d", encoded, test.input, decoded)
		}
	}
}

func TestBase62DecodeInvalid(t *testing.T) {
	t.Parallel()

	// Test with invalid characters in Base62 string
	invalidStrings := []string{"~", "@", "123abc#", "xyz!"}

	for _, str := range invalidStrings {
		_, err := DecodeBase62(str)
		if err == nil {
			t.Errorf("Base62Decode should have failed for invalid input %s", str)
		}
	}
}

func TestCounterHashUnique(t *testing.T) {
	t.Parallel()
	const million = 1000000
	resultMap := map[string]bool{}
	for i := range million {
		hashStr := EncodeBase62(uint64(i))
		if _, ok := resultMap[hashStr]; ok {
			t.Errorf("hash already exists")
		}
		resultMap[hashStr] = true
	}
}
