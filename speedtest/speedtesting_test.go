package speedtest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadSpeedTestReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "speedtest.txt")
	content := "bad row\n144.11 MB/s | 171.95 MB/s | 2026-09-29 17:46:41\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	download, upload, timestamps, err := ReadSpeedTestReport(path)
	if err != nil {
		t.Fatalf("ReadSpeedTestReport returned error: %v", err)
	}
	if len(download) != 1 || download[0] != "144.11" || upload[0] != "171.95" || len(timestamps) != 1 {
		t.Fatalf("unexpected report values: %v, %v, %v", download, upload, timestamps)
	}
}

func TestReadSpeedTestReportRejectsEmptyReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "speedtest.txt")
	if err := os.WriteFile(path, []byte("bad row\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := ReadSpeedTestReport(path); err == nil {
		t.Fatal("expected invalid report error")
	}
}
