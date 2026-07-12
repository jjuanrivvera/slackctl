package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeSecret(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"bracketed paste wrappers stripped", "\x1b[200~KEY\x1b[201~\n", "KEY"},
		{"clean key unchanged", "KEY", "KEY"},
		{"surrounding whitespace trimmed", "  KEY  ", "KEY"},
		{"lone start marker stripped", "\x1b[200~KEY", "KEY"},
		{"lone end marker stripped", "KEY\x1b[201~", "KEY"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, sanitizeSecret(tc.in))
		})
	}
}
