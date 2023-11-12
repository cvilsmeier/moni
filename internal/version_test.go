package internal

import (
	"os"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	readFile := func(filename string) string {
		data, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	cutout := func(s, from, to string) string {
		_, rest, found := strings.Cut(s, from)
		if !found {
			t.Fatalf("%q not found in %q", from, s)
		}
		s = rest
		rest, _, found = strings.Cut(s, to)
		if !found {
			t.Fatalf("%q not found in %q", to, s)
		}
		return rest
	}
	// version.go and README must have same version
	text := readFile("../README.md")
	readmeVersion := "v" + cutout(text, "### v", "\n")
	if readmeVersion != Version {
		t.Fatalf("version.go has %q but README.md has %q", Version, readmeVersion)
	}
}
