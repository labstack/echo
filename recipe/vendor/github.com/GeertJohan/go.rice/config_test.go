package rice

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/GeertJohan/go.rice/embedded"
)

// For all test code in this package, define a set of test boxes.
var eb1 *embedded.EmbeddedBox
var ab1, ab2 *appendedBox
var fsb1, fsb2, fsb3 string // paths to filesystem boxes
func init() {
	var err error

	// Box1 exists in all three locations.
	eb1 = &embedded.EmbeddedBox{Name: "box1"}
	embedded.RegisterEmbeddedBox(eb1.Name, eb1)
	ab1 = &appendedBox{Name: "box1"}
	appendedBoxes["box1"] = ab1
	fsb1, err = ioutil.TempDir("", "box1")
	if err != nil {
		panic(err)
	}

	// Box2 exists in only appended and FS.
	ab2 = &appendedBox{Name: "box2"}
	appendedBoxes["box2"] = ab2
	fsb2, err = ioutil.TempDir("", "box2")
	if err != nil {
		panic(err)
	}

	// Box3 exists only on disk.
	fsb3, err = ioutil.TempDir("", "box3")
	if err != nil {
		panic(err)
	}

	// Also, replace the default filesystem lookup path to directly support the
	// on-disk temp directories.
	resolveAbsolutePathFromCaller = func(name string, n int) (string, error) {
		if name == "box1" {
			return fsb1, nil
		} else if name == "box2" {
			return fsb2, nil
		} else if name == "box3" {
			return fsb3, nil
		}
		return "", fmt.Errorf("Unknown box name: %q", name)
	}
}

func TestDefaultLookupOrder(t *testing.T) {
	// Box1 exists in all three, so the default order should find the embedded.
	b, err := FindBox("box1")
	if err != nil {
		t.Fatalf("Expected to find box1, got error: %v", err)
	}
	if b.embed != eb1 {
		t.Fatalf("Expected to find embedded box, but got %#v", b)
	}

	// Box2 exists in appended and FS, so find the appended.
	b2, err := FindBox("box2")
	if err != nil {
		t.Fatalf("Expected to find box2, got error: %v", err)
	}
	if b2.appendd != ab2 {
		t.Fatalf("Expected to find appended box, but got %#v", b2)
	}

	// Box3 exists only on FS, so find it there.
	b3, err := FindBox("box3")
	if err != nil {
		t.Fatalf("Expected to find box3, got error: %v", err)
	}
	if b3.absolutePath != fsb3 {
		t.Fatalf("Expected to find FS box, but got %#v", b3)
	}
}

func TestConfigLocateOrder(t *testing.T) {
	cfg := Config{LocateOrder: []LocateMethod{LocateFS, LocateAppended, LocateEmbedded}}
	fsb := []string{fsb1, fsb2, fsb3}
	// All 3 boxes have a FS backend, so we should always find that.
	for i, boxName := range []string{"box1", "box2", "box3"} {
		b, err := cfg.FindBox(boxName)
		if err != nil {
			t.Fatalf("Expected to find %q, got error: %v", boxName, err)
		}
		if b.absolutePath != fsb[i] {
			t.Fatalf("Expected to find FS box, but got %#v", b)
		}
	}

	cfg.LocateOrder = []LocateMethod{LocateAppended, LocateFS, LocateEmbedded}
	{
		b, err := cfg.FindBox("box3")
		if err != nil {
			t.Fatalf("Expected to find box3, got error: %v", err)
		}
		if b.absolutePath != fsb3 {
			t.Fatalf("Expected to find FS box, but got %#v", b)
		}
	}
	{
		b, err := cfg.FindBox("box2")
		if err != nil {
			t.Fatalf("Expected to find box2, got error: %v", err)
		}
		if b.appendd != ab2 {
			t.Fatalf("Expected to find appended box, but got %#v", b)
		}
	}

	// What if we don't list all the locate methods?
	cfg.LocateOrder = []LocateMethod{LocateEmbedded}
	{
		b, err := cfg.FindBox("box2")
		if err == nil {
			t.Fatalf("Expected not to find box2, but something was found: %#v", b)
		}
	}
	{
		b, err := cfg.FindBox("box1")
		if err != nil {
			t.Fatalf("Expected to find box2, got error: %v", err)
		}
		if b.embed != eb1 {
			t.Fatalf("Expected to find embedded box, but got %#v", b)
		}
	}
}
