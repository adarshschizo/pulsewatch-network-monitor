package ping

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type PingStats struct {
	Min        []string
	Avg        []string
	Max        []string
	Sttdev     []string
	TimeString []string
}

// Function to scan ping for network latency then save the stats to ping/ping.txt
func RecordPingData(workingDir string) (scanningErr error) {
	// Initialize path of the report file
	workingDirReport := filepath.Join(workingDir, "ping", "ping.txt")
	if err := os.MkdirAll(filepath.Dir(workingDirReport), 0755); err != nil {
		return err
	}
	// Testing log
	fmt.Printf("Working dir ping: %s\n", workingDirReport)

	var pingCommand *exec.Cmd
	if os.PathSeparator == '\\' {
		pingCommand = exec.Command("ping", "-n", "10", "google.com")
	} else {
		pingCommand = exec.Command("ping", "-c", "10", "google.com")
	}
	scanResult, pingScanningErr := pingCommand.Output()
	if pingScanningErr != nil {
		return pingScanningErr
	}

	result, parseErr := parsePingResult(string(scanResult))
	if parseErr != nil {
		return parseErr
	}

	file, openFileErr := os.OpenFile(workingDirReport, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openFileErr != nil {
		return openFileErr
	}
	defer file.Close()

	finalString := result + " | " + time.Now().Format("2006-01-02 15:04:05") + "\n"

	// Write the text to the file
	_, writeFileErr := file.WriteString(finalString)
	if writeFileErr != nil {
		return writeFileErr
	}

	// Return nil if no error
	return nil
}

func parsePingResult(output string) (string, error) {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "round-trip") || strings.Contains(line, "rtt min/avg/max") {
			return strings.TrimSpace(line), nil
		}
		if strings.Contains(line, "Minimum") && strings.Contains(line, "Average") {
			values := regexp.MustCompile(`Minimum = ([0-9]+)ms, Maximum = ([0-9]+)ms, Average = ([0-9]+)ms`).FindStringSubmatch(line)
			if len(values) == 4 {
				return fmt.Sprintf("round-trip min/avg/max/stddev = %s/%s/%s/0 ms", values[1], values[3], values[2]), nil
			}
		}
	}
	return "", fmt.Errorf("could not parse ping output")
}

// Function to read the report and return the stats, ready for chart building
func ReadPingReport(reportPath string) (pingStats PingStats, err error) {

	report, openFileErr := os.ReadFile(reportPath)
	if openFileErr != nil {
		return PingStats{}, openFileErr
	}

	// Extract data
	pingStats = PingStats{}

	for _, line := range strings.Split(strings.TrimSpace(string(report)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lineSlice := strings.Split(line, "|")
		if len(lineSlice) != 2 || !strings.Contains(lineSlice[0], "=") {
			continue
		}
		time := lineSlice[1]
		data := lineSlice[0]
		stats := strings.Split(strings.TrimSpace(strings.Split(data, "=")[1]), "/")
		if len(stats) < 4 {
			continue
		}

		pingStats.Min = append(pingStats.Min, stats[0])
		pingStats.Avg = append(pingStats.Avg, stats[1])
		pingStats.Max = append(pingStats.Max, stats[2])
		pingStats.Sttdev = append(pingStats.Sttdev, strings.Split(stats[3], " ")[0])
		pingStats.TimeString = append(pingStats.TimeString, time)
	}
	if len(pingStats.Avg) == 0 {
		return PingStats{}, fmt.Errorf("ping report contains no valid samples")
	}

	return pingStats, nil

}
