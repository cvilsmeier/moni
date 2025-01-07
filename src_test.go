package main

import (
	"crypto/md5"
	"encoding/hex"
	"io/fs"
	"os"
	"strings"
	"testing"
)

// TestSrc ensures all source files are in good shape.
func TestSrc(t *testing.T) {
	readFile := func(name string) string {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	cutout := func(s, from, to string) string {
		_, after, ok := strings.Cut(s, from)
		if !ok {
			t.Fatal("cutout not ok")
		}
		before, _, ok := strings.Cut(after, to)
		if !ok {
			t.Fatal("cutout not ok")
		}
		return before
	}
	md5sum := func(s string) string {
		sum := md5.Sum([]byte(s))
		return hex.EncodeToString(sum[:])
	}
	// calc checksums of usage docs in moni.go and README.md, so
	// that if one changes, we get nagged to change the other.
	t.Run("UsageMustBeInSync", func(t *testing.T) {
		// checksum moni.go
		{
			text := cutout(readFile("moni.go"), "func usage() {", "// end usage")
			checksum := md5sum(text)
			if checksum != "e56bed1242559ad0b613f565cff9bb4e" {
				t.Fatal("wrong", checksum)
			}
		}
		// checksum README.md
		{
			text := readFile("README.md")
			text = cutout(text, "## Usage", "## Changelog")
			text = cutout(text, "```", "```")
			checksum := md5sum(text)
			if checksum != "af1f7e2aea75cfac446d365cb33238b9" {
				t.Fatal("wrong", checksum)
			}
		}
	})
	// version.go and README.md must have same version
	t.Run("VersionMustBeInSync", func(t *testing.T) {
		v := "v" + cutout(readFile("README.md"), "### v", "\n")
		if v != Version {
			t.Fatal("wrong readme version", v)
		}
	})
	// find FIXME markers in source code
	t.Run("FixmeMarkers", func(t *testing.T) {
		_, err := os.Stat("go.mod")
		if err != nil {
			t.Fatal("must be in src/")
		}
		err = fs.WalkDir(os.DirFS("."), ".", func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				t.Fatal("WalkDir", err)
			}
			if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "src_test.go") {
				text := readFile(path)
				if strings.Contains(text, "cvvvv") || strings.Contains(text, "FIXME") {
					t.Fatal("found cvvvv/FIXME ", path)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal("WalkDir", err)
		}
	})
}
