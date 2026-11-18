package demo

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteExportsStaysInTheDataDirectory(t *testing.T) {
	data := t.TempDir()
	smart, wg, err := WriteExports(data, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{smart, wg} {
		if filepath.Dir(filepath.Dir(dir)) != data {
			t.Fatalf("%s is not below the data directory %s", dir, data)
		}
		if entries, _ := os.ReadDir(dir); len(entries) == 0 {
			t.Fatalf("%s is empty", dir)
		}
	}
	// A restart rewrites the same directory instead of adding another one.
	if _, _, err := WriteExports(data, time.Now()); err != nil {
		t.Fatal(err)
	}
	if entries, _ := os.ReadDir(data); len(entries) != 1 {
		t.Fatalf("the data directory holds %d entries, want 1", len(entries))
	}
}
