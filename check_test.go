package main

import (
	"io/fs"
	"os"
	"strings"
	"testing"
)

// TestCheck checks all source files.
func TestCheck(t *testing.T) {
	readTextFile := func(name string) string {
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		text = strings.ReplaceAll(text, "\r", "")
		return text
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
	// usage in README.md must be in sync
	t.Run("UsageMustBeInSync", func(t *testing.T) {
		var sb strings.Builder
		printUsage(&sb)
		usageText := sb.String()
		readmeText := readTextFile("README.md")
		if !strings.Contains(readmeText, usageText) {
			t.Fatalf("wrong usage in README.md,\nwant %q\nhave %q", usageText, readmeText)
		}
	})
	// README.md version must be in sync
	t.Run("VersionMustBeInSync", func(t *testing.T) {
		v := "v" + cutout(readTextFile("README.md"), "### v", "\n")
		if v != Version {
			t.Fatal("wrong readme version", v)
		}
	})
	// find markers in source code
	t.Run("FixmeMarkers", func(t *testing.T) {
		readTextFile("go.mod") // make sure we're in root dir
		err := fs.WalkDir(os.DirFS("."), ".", func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".go") {
				text := readTextFile(path)
				if strings.Contains(text, "cv"+"vvv") || strings.Contains(text, "FIX"+"ME") {
					t.Fatal("found cv"+"vvv/FIX"+"ME in ", path)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal("fs.WalkDir", err)
		}
	})
}
