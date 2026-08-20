package main

import (
	"archive/tar"
	"bytes"
	"io/fs"
	"strings"
	"testing"
)

func TestRunRequiresBaseRef(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(nil, &stdout, &stderr); code != 2 {
		t.Fatalf("expected usage exit code, got %d", code)
	}
	if !strings.Contains(stderr.String(), "--base-ref") {
		t.Fatalf("expected actionable usage, got %q", stderr.String())
	}
}

func TestTarFSLoadsRegularFilesAndDirectories(t *testing.T) {
	var archive bytes.Buffer
	writer := tar.NewWriter(&archive)
	if err := writer.WriteHeader(&tar.Header{Name: "courses/", Typeflag: tar.TypeDir, Mode: 0o755}); err != nil {
		t.Fatal(err)
	}
	content := []byte("sources: {}\n")
	if err := writer.WriteHeader(&tar.Header{Name: "sources.yaml", Typeflag: tar.TypeReg, Mode: 0o644, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	archiveFS, err := tarFS(archive.Bytes())
	if err != nil {
		t.Fatalf("load tar filesystem: %v", err)
	}
	got, err := fs.ReadFile(archiveFS, "sources.yaml")
	if err != nil || string(got) != string(content) {
		t.Fatalf("read archived file: %q, %v", got, err)
	}
}

func TestTarFSRejectsNonRegularEntries(t *testing.T) {
	var archive bytes.Buffer
	writer := tar.NewWriter(&archive)
	if err := writer.WriteHeader(&tar.Header{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "target"}); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := tarFS(archive.Bytes()); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported entry error, got %v", err)
	}
}
