package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/adarshschizo/pulsewatch-network-monitor/cronjob"
	"github.com/adarshschizo/pulsewatch-network-monitor/network"
	"github.com/adarshschizo/pulsewatch-network-monitor/ping"
	"github.com/adarshschizo/pulsewatch-network-monitor/speedtest"
)

func main() {

	// Get working directory by env
	WORKING_DIR := os.Getenv("WORKING_DIR")

	// If env is not available, set the working directory to current working dir
	if WORKING_DIR == "" {
		workingDirectory, err := os.Getwd()
		if executable, executableErr := os.Executable(); executableErr == nil && filepath.Base(executable) == "pulsewatch.exe" {
			workingDirectory = filepath.Dir(executable)
		} else if err != nil {
			log.Fatal(err)
		}
		WORKING_DIR = workingDirectory
	}

	advancedMode := false
	collectOnce := false
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-a", "--advanced":
			advancedMode = true
		case "--collect", "--once":
			collectOnce = true
		case "-h", "--help":
			printUsage()
			return
		default:
			fmt.Printf("Invalid command line: %s\n\n", os.Args[1])
			printUsage()
			os.Exit(2)
		}
	}

	// Welcome message
	fmt.Println("Welcome to Pulsewatch")
	fmt.Println("---------------------")

	// Get into advanced mode by using -a
	if advancedMode {
		RunTerminal(WORKING_DIR)
	}
	if collectOnce {
		if err := collectReports(WORKING_DIR); err != nil {
			log.Fatal(err)
		}
		return
	}

	// If user do not input custom -a flag, then switch to collecting data mode

	// If we already set up cronjob, then we will collect data
	if cronjob.IsScheduled() {
		fmt.Println("We already set up cronjob")
		if err := collectReports(WORKING_DIR); err != nil {
			log.Fatal(err)
		}
		return

	} else {
		fmt.Println("Look like we haven't set up the cronjob, let's do it now!")
		err := cronjob.SetUpCronJob(WORKING_DIR)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func printUsage() {
	fmt.Println("Pulsewatch - local network monitoring")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  pulsewatch              Set up or run scheduled collection")
	fmt.Println("  pulsewatch --collect    Collect one report immediately")
	fmt.Println("  pulsewatch --once       Alias for --collect")
	fmt.Println("  pulsewatch -a           Open the advanced menu")
	fmt.Println("  pulsewatch --advanced   Open the advanced menu")
	fmt.Println("  pulsewatch -h           Show this help")
}

func collectReports(workingDir string) error {
	fmt.Println("Collecting latency, network traffic, and speed reports...")
	if err := ping.RecordPingData(workingDir); err != nil {
		return fmt.Errorf("collect latency report: %w", err)
	}
	if err := network.RecordNetworkData(workingDir); err != nil {
		return fmt.Errorf("collect network report: %w", err)
	}
	if err := speedtest.RecordSpeedTestData(workingDir); err != nil {
		return fmt.Errorf("collect speed report: %w", err)
	}
	fmt.Println("Collection complete")
	return nil
}
