package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExampleManifestsDecode(t *testing.T) {
	entries, err := os.ReadDir(filepath.Join("..", "..", "examples", "settings"))
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yml" {
			continue
		}
		count++
		t.Run(entry.Name(), func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("..", "..", "examples", "settings", entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Decode(data); err != nil {
				t.Fatalf("Decode(%s): %v", entry.Name(), err)
			}
		})
	}
	if count == 0 {
		t.Fatal("settings examples directory contains no .yml manifests")
	}
}
