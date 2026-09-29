package ping

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePingResultWindows(t *testing.T) {
	result, err := parsePingResult("Minimum = 7ms, Maximum = 11ms, Average = 8ms")
	if err != nil {
		t.Fatalf("parsePingResult returned error: %v", err)
	}
	want := "round-trip min/avg/max/stddev = 7/8/11/0 ms"
	if result != want {
		t.Fatalf("got %q, want %q", result, want)
	}
}

func TestReadPingReportRejectsInvalidSamples(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "ping.txt")
	content := "not a report\nround-trip min/avg/max/stddev = 7/8/11/0 ms | 2026-09-29 17:45:55\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	stats, err := ReadPingReport(path)
	if err != nil {
		t.Fatalf("ReadPingReport returned error: %v", err)
	}
	if len(stats.Avg) != 1 || stats.Avg[0] != "8" {
		t.Fatalf("unexpected parsed stats: %#v", stats)
	}
}

func TestReadPingReportRejectsEmptyReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ping.txt")
	if err := os.WriteFile(path, []byte("invalid\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadPingReport(path); err == nil {
		t.Fatal("expected invalid report error")
	}
}
