package echo

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkContext_File_RealServer(b *testing.B) {
	if os.Getenv("ECHO_HEAVY_BENCHMARK") != "true" {
		b.Skip("skipping heavy benchmark; set ECHO_HEAVY_BENCHMARK=true to run")
	}
	e := New()
	tmpDir := b.TempDir()
	const benchFileName = "real_bench_data.bin"
	// 100MB file to observe kernel-level copy savings via sendfile
	fileSize := 100 * 1024 * 1024
	content := make([]byte, fileSize)
	for i := range content {
		content[i] = byte(i % 256)
	}
	_ = os.WriteFile(filepath.Join(tmpDir, benchFileName), content, 0644)
	e.Filesystem = os.DirFS(tmpDir)

	// Route 1: Optimized path (Standard Echo handles this automatically)
	e.GET("/optimized", func(c *Context) error {
		return c.File(benchFileName)
	})

	// Route 2: Non-optimized path (Disables optimization by registering an After hook)
	e.GET("/standard", func(c *Context) error {
		c.Response().(*Response).After(func() {})
		return c.File(benchFileName)
	})

	// Use a real TCP server to exercise the kernel's sendfile(2) path through ReadFrom.
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := ts.Client()

	b.Run("Zero-Copy-Optimized", func(b *testing.B) {
		url := ts.URL + "/optimized"
		b.ReportAllocs()
		b.SetBytes(int64(fileSize))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := client.Get(url)
			if err != nil {
				b.Fatal(err)
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})

	b.Run("User-Space-Standard", func(b *testing.B) {
		url := ts.URL + "/standard"
		b.ReportAllocs()
		b.SetBytes(int64(fileSize))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			resp, err := client.Get(url)
			if err != nil {
				b.Fatal(err)
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	})
}
