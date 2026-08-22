package service_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/abyankamal/sidak/backend/config"
	"github.com/abyankamal/sidak/backend/internal/service"
)

type mockFile struct {
	*bytes.Reader
}

func (m *mockFile) Close() error {
	return nil
}

func TestStorageService_SaveUpload_PathTraversal(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "sidak_storage_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		StorageBasePath: tempDir,
		StoragePublicURL: "http://localhost:8080/uploads",
	}

	svc := service.NewStorageService(cfg)
	ctx := context.Background()

	t.Run("Valid upload", func(t *testing.T) {
		content := []byte("test content")
		fileReader := &mockFile{Reader: bytes.NewReader(content)}
		header := &multipart.FileHeader{
			Filename: "test.txt",
			Size:     int64(len(content)),
		}

		resp, err := svc.SaveUpload(ctx, header, fileReader, "lampiran", "TX123", "3201010101010001")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if !strings.Contains(resp.FilePath, "TX123_test.txt") {
			t.Errorf("unexpected file path: %s", resp.FilePath)
		}
	})

	t.Run("Path traversal in wargaNIK", func(t *testing.T) {
		content := []byte("malicious content")
		fileReader := &mockFile{Reader: bytes.NewReader(content)}
		header := &multipart.FileHeader{
			Filename: "evil.txt",
			Size:     int64(len(content)),
		}

		resp, err := svc.SaveUpload(ctx, header, fileReader, "lampiran", "TX123", "../../etc")
		if err != nil {
			t.Fatalf("expected no error due to sanitization, got: %v", err)
		}

		// Verify the file was saved inside tempDir and safeWargaNIK became "etc"
		absPath := filepath.Join(tempDir, strings.TrimPrefix(resp.FilePath, "uploads/"))
		if !strings.HasPrefix(absPath, tempDir) {
			t.Errorf("file escaped storage dir: %s", absPath)
		}
	})

	t.Run("Path traversal in transaksiID", func(t *testing.T) {
		content := []byte("malicious content 2")
		fileReader := &mockFile{Reader: bytes.NewReader(content)}
		header := &multipart.FileHeader{
			Filename: "evil2.txt",
			Size:     int64(len(content)),
		}

		resp, err := svc.SaveUpload(ctx, header, fileReader, "lampiran", "../../../tmp/evil", "3201010101010001")
		if err != nil {
			t.Fatalf("expected no error due to sanitization, got: %v", err)
		}

		absPath := filepath.Join(tempDir, strings.TrimPrefix(resp.FilePath, "uploads/"))
		if !strings.HasPrefix(absPath, tempDir) {
			t.Errorf("file escaped storage dir: %s", absPath)
		}
	})
}
