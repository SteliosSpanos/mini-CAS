package path

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("directory does not exist: %s", path)
		return
	}
	if err == nil && !info.IsDir() {
		t.Errorf("expected directory, got file: %s", path)
	}
}

func TestInit(t *testing.T) {
	tempDir := t.TempDir()

	repo, err := Init(tempDir)
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	if repo == nil {
		t.Fatal("Init() returned nil repository")
	}

	expectedCasDir := filepath.Join(tempDir, CASDir)
	if repo.RootDir != expectedCasDir {
		t.Errorf("RootDir = %q, want %q", repo.RootDir, expectedCasDir)
	}

	assertDirExists(t, expectedCasDir)
	assertDirExists(t, filepath.Join(expectedCasDir, "storage"))
}

func TestInit_AlreadyExists(t *testing.T) {
	tempDir := t.TempDir()

	_, err := Init(tempDir)
	if err != nil {
		t.Fatalf("first Init() error: %v", err)
	}

	_, err = Init(tempDir)
	if err == nil {
		t.Error("second Init() should fail, got nil error")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error should mention 'already exists', got: %v", err)
	}
}

func TestOpen(t *testing.T) {
	tempDir := t.TempDir()

	_, err := Init(tempDir)
	if err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	repo, err := Open(tempDir)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}

	if repo == nil {
		t.Fatal("Open() returned nil repository")
	}

	expectedCasDir := filepath.Join(tempDir, CASDir)
	if repo.RootDir != expectedCasDir {
		t.Errorf("RootDir = %q, want %q", repo.RootDir, expectedCasDir)
	}
}

func TestOpen_NotExists(t *testing.T) {
	tempDir := t.TempDir()

	_, err := Open(tempDir)
	if err == nil {
		t.Error("Open() should fail for non-existent repository")
	}

	if !strings.Contains(err.Error(), "no CAS repository") {
		t.Errorf("error should mention 'no CAS repository', got %v", err)
	}
}

func TestInit_EmptyPath(t *testing.T) {
	tempDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error: %v", err)
	}

	os.Chdir(tempDir)

	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	repo, err := Init("")
	if err != nil {
		t.Fatalf("Init('') error: %v", err)
	}

	expectedCasDir := filepath.Join(tempDir, CASDir)
	if repo.RootDir != expectedCasDir {
		t.Errorf("RootDir = %q, want %q", repo.RootDir, expectedCasDir)
	}
}
