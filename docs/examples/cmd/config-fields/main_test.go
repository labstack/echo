// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractOnlyExportedConfigFields(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "middleware")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	source := `package middleware
type ExampleConfig struct {
	// Deprecated: use New instead.
	Old bool
	New string
	private int
}
type hiddenConfig struct { Visible bool }
type OtherType struct { Visible bool }
`
	if err := os.WriteFile(filepath.Join(dir, "example.go"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := extract(root, "abc123")
	if err != nil {
		t.Fatal(err)
	}
	if got.Revision != "abc123" || len(got.Configs) != 1 {
		t.Fatalf("unexpected manifest: %#v", got)
	}
	fields := got.Configs[0].Fields
	if got.Configs[0].File != "middleware/example.go" || len(fields) != 2 {
		t.Fatalf("unexpected config: %#v", got.Configs[0])
	}
	if fields[0].Name != "Old" || fields[0].Type != "bool" || !fields[0].Deprecated || fields[1].Name != "New" {
		t.Fatalf("unexpected fields: %#v", fields)
	}
}
