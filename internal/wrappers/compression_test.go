package wrappers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompressionWrapper(t *testing.T) {
	// Create a test directory
	tempDir, err := os.MkdirTemp("", "compress_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	sourceDir := filepath.Join(tempDir, "source")
	err = os.MkdirAll(sourceDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	testFilePath := filepath.Join(sourceDir, "testfile.txt")
	testContent := "This is test content for compression"
	err = os.WriteFile(testFilePath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create destination paths
	tarGzPath := filepath.Join(tempDir, "archive.tar.gz")
	extractDir := filepath.Join(tempDir, "extract")

	// Create compression wrapper
	cw, err := NewCompressionWrapper()
	if err != nil {
		t.Fatalf("Failed to create compression wrapper: %v", err)
	}

	// Test compression
	err = cw.Compress(sourceDir, tarGzPath, FormatTarGz)
	if err != nil {
		t.Fatalf("Compression failed: %v", err)
	}

	// Verify compressed file exists
	if _, err := os.Stat(tarGzPath); os.IsNotExist(err) {
		t.Fatalf("Compressed file doesn't exist: %s", tarGzPath)
	}

	// Test decompression
	err = cw.Decompress(tarGzPath, extractDir)
	if err != nil {
		t.Fatalf("Decompression failed: %v", err)
	}

	// Verify file was extracted correctly
	extractedFilePath := filepath.Join(extractDir, "testfile.txt")
	if _, err := os.Stat(extractedFilePath); os.IsNotExist(err) {
		t.Fatalf("Extracted file doesn't exist: %s", extractedFilePath)
	}

	// Verify content
	extractedContent, err := os.ReadFile(extractedFilePath)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}

	if string(extractedContent) != testContent {
		t.Errorf("Extracted content doesn't match original. Expected: %s, Got: %s",
			testContent, string(extractedContent))
	}
}