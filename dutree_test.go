package dutree

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestScanComputesDirectorySizes(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), 100)
	writeFile(t, filepath.Join(root, "sub", "b.txt"), 250)

	e, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if e.Size != 350 {
		t.Fatalf("root size = %d, want 350", e.Size)
	}
	if len(e.Children) != 2 {
		t.Fatalf("children = %d, want 2", len(e.Children))
	}
	// Children are sorted largest first, so "sub" (250 bytes) comes before
	// "a.txt" (100 bytes).
	if e.Children[0].Name != "sub" {
		t.Fatalf("children[0] = %q, want %q", e.Children[0].Name, "sub")
	}
}

func TestScanConcurrentSubdirectories(t *testing.T) {
	root := t.TempDir()
	const dirs = 32
	var want int64
	for i := 0; i < dirs; i++ {
		size := 100 + i
		writeFile(t, filepath.Join(root, fmt.Sprintf("d%d", i), "f.txt"), size)
		want += int64(size)
	}

	e, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if e.Size != want {
		t.Fatalf("root size = %d, want %d", e.Size, want)
	}
	if len(e.Children) != dirs {
		t.Fatalf("children = %d, want %d", len(e.Children), dirs)
	}
}

func TestScanContextAlreadyCancelled(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), 10)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	e, err := ScanContext(ctx, root, Options{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if e != nil {
		t.Fatalf("entry = %v, want nil", e)
	}
}

func TestScanContextCancelledMidScan(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 8; i++ {
		writeFile(t, filepath.Join(root, fmt.Sprintf("d%d", i), "f.txt"), 10)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := ScanContext(ctx, root, Options{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestWriteJSONRoundTrips(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "a.txt"), 42)

	e, err := Scan(root, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	var buf bytes.Buffer
	if err := WriteJSON(&buf, e); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var got Entry
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.Size != e.Size {
		t.Fatalf("decoded size = %d, want %d", got.Size, e.Size)
	}
}

func TestHumanSize(t *testing.T) {
	cases := map[int64]string{
		0:                 "0B",
		1023:              "1023B",
		1024:              "1.0KiB",
		1536:              "1.5KiB",
		10 * 1024 * 1024:  "10.0MiB",
	}
	for in, want := range cases {
		if got := HumanSize(in); got != want {
			t.Errorf("HumanSize(%d) = %q, want %q", in, got, want)
		}
	}
}

func writeFile(t *testing.T, path string, size int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
		t.Fatal(err)
	}
}
