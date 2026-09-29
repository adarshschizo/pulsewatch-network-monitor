package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/adarshschizo/pulsewatch-network-monitor/chart"
	"github.com/adarshschizo/pulsewatch-network-monitor/cronjob"
)

func RunTerminal(workingDir string) {
	for {
		clearTerminal()
		choice := displayMenu("Pulsewatch", []string{
			"Cronjob options",
			"Show process network usage chart",
			"Show network latency chart",
			"Quit",
		})

		switch choice {
		case 1:
			cronJobOptions(workingDir)
		case 2:
			if err := chart.CreateNetworkChart(workingDir); err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		case 3:
			if err := chart.CreateSpeedtestChart(workingDir); err != nil {
				log.Fatal(err)
			}
			if err := chart.CreatePingChart(workingDir); err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		case 4:
			fmt.Println("Goodbye! See you later")
			os.Exit(0)
		}
	}
}

func cronJobOptions(workingDir string) {
	clearTerminal()
	choice := displayMenu("Cronjob options", []string{
		"Edit cronjob time",
		"Remove cronjob completely",
		"Come back",
	})

	switch choice {
	case 1:
		if err := cronjob.SaveCronJob("", workingDir, "remove"); err != nil {
			log.Fatal(err)
		}
		if err := cronjob.SetUpCronJob(workingDir); err != nil {
			log.Fatal(err)
		}
	case 2:
		if err := cronjob.SaveCronJob("", workingDir, "remove"); err != nil {
			log.Fatal(err)
		}
	case 3:
		fmt.Println("Coming back")
	}
}

func displayMenu(title string, options []string) int {
	for {
		fmt.Println(title)
		for index, option := range options {
			fmt.Printf("  %d. %s\n", index+1, option)
		}
		fmt.Print("Select an option: ")

		var choice int
		if _, err := fmt.Scanln(&choice); err == nil && choice >= 1 && choice <= len(options) {
			return choice
		}
		fmt.Println("Please enter one of the listed numbers.")
	}
}

func clearTerminal() {
	command := "clear"
	if os.PathSeparator == '\\' {
		command = "cls"
	}
	clear := exec.Command(command)
	clear.Stdout = os.Stdout
	_ = clear.Run()
}
