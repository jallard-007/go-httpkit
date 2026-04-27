package negotiate_test

import (
	"fmt"
	"testing"

	"github.com/jallard-007/go-httpkit/negotiate"
)

func TestEncoding(t *testing.T) {
	available := []string{"br", "gzip", "identity"}
	tests := []struct {
		acceptEncodings []string
		wantEncoding    string
	}{
		// baseline
		{[]string{}, "identity"},
		{[]string{""}, "identity"},
		{[]string{"br"}, "br"},
		{[]string{"gzip"}, "gzip"},
		{[]string{"identity"}, "identity"},
		{[]string{"foobar"}, "identity"},

		// multiple options
		{[]string{"br, gzip, identity"}, "br"},
		{[]string{"gzip, br, foobar"}, "br"},
		{[]string{"gzip, identity"}, "gzip"},

		// wildcard
		{[]string{"*"}, "br"},
		{[]string{"*, identity"}, "br"},
		{[]string{"identity, *"}, "br"},
		{[]string{"*;q=0, identity"}, "identity"},
		{[]string{"*;q=0.9, identity"}, "identity"},
		{[]string{"*;q=0.9, identity;q=0.8"}, "br"},
		{[]string{"*;q=0.9, identity;q=0.8", "br;q=0.7"}, "gzip"},
		{[]string{"identity, *;q=0"}, "identity"},

		// special cases
		{[]string{"foobar, identity;q=0"}, ""},
		{[]string{"identity;q=0"}, ""},
		{[]string{"*;q=0"}, ""},
		{[]string{"br;q=0"}, "identity"},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i+1), func(t *testing.T) {
			encoding := negotiate.Encoding(tt.acceptEncodings, available)
			if encoding != tt.wantEncoding {
				t.Errorf("got %q, wanted %q", encoding, tt.wantEncoding)
			}
		})
	}

	t.Run("nil available", func(t *testing.T) {
		encoding := negotiate.Encoding([]string{"br;q=0.9, gzip, identity, *;q=0.8"}, nil)
		if encoding != "" {
			t.Errorf("got %q, wanted %q", encoding, "")
		}
	})
}

func BenchmarkEncoding(b *testing.B) {
	for b.Loop() {
		_ = negotiate.Encoding([]string{"br;q=0.9, gzip, identity, *;q=0.8"}, []string{"br"})
	}
}
