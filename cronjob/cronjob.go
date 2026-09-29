package cronjob

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const windowsTaskName = "Pulsewatch Network Scan"

func IsScheduled() bool {
	if runtime.GOOS == "windows" {
		return exec.Command("schtasks", "/Query", "/TN", windowsTaskName).Run() == nil
	}
	output, err := exec.Command("sh", "-c", "crontab -l 2>/dev/null").Output()
	return err == nil && (strings.Contains(string(output), "scanning") || strings.Contains(string(output), "pulsewatch"))
}

// Calculate timestring in cronjob format, ex: * * * * *
func calculateTimeString(timeInMinutes int) (timeStringInCronjob string) {
	// Initialize the first and second position
	firstPosition := "*"
	secondPosition := "*"

	// If 24 hours
	if timeInMinutes == 1440 {
		firstPosition = "0"
		secondPosition = "0"

	}
	// If 60  minute
	if timeInMinutes == 60 {
		firstPosition = "0"
	}

	// If less than 60 minutes and not 1 minutes
	if timeInMinutes < 60 && timeInMinutes != 1 {
		firstPosition += "/" + strconv.Itoa(timeInMinutes)
	}

	// If more than 60 minutes and below 24 hours
	if timeInMinutes > 60 && timeInMinutes < 1440 {
		secondPosition += "/" + strconv.Itoa(timeInMinutes/60)
	}

	finalString := firstPosition + " " + secondPosition + " * * *"

	return finalString

}

// Ask for user time input, return time in minutes
func askForTimeInput() (timeInMinutes int) {
	var time int

	fmt.Println("Please choose how often you want to check your network latency")
	fmt.Println("Minimum: 1 minutes, maximum: 1 day")

	for {
		var inputTime string
		var timeMark string
		var min int
		fmt.Println("Possible input: 1 - 60 mins, 1 - 24 hrs. We don't support decimal hrs and mins")
		fmt.Printf("Your chosen time is: ")

		// Get user input
		fmt.Scanf("%s %s", &inputTime, &timeMark)

		// If user input in mins, convert that to min
		if strings.Contains(timeMark, "mins") {
			minString := strings.Split(inputTime, "mins")[0]
			min, _ = strconv.Atoi(minString)
			// If user input in hrs, convert that to min
		} else if strings.Contains(timeMark, "hrs") {
			minString := strings.Split(inputTime, "hrs")[0]
			hours, _ := strconv.Atoi(minString)
			min = hours * 60
			// If user input is not hrs or mins, prompt again
		} else {
			fmt.Println("It's either minutes or hours")
			continue
		}
		//  If min is between 1 and 60 or min is between 1 hr and 24 hrs and the number is whole
		if min >= 1 && min <= 60 || min > 60 && min%60 == 0 && min <= 1440 {
			time = min
			break
			// If the time is not within the limit, prompt user again
		} else {
			fmt.Println("Please remember the maximum time")
		}

	}

	return time
}

// Save the cronjob to the system, has 2 mode: add and remove
// add mode adds the cronjob to the system
// remove mode removes the cronjob from the system
func SaveCronJob(timeStringInCronjob string, WORKING_DIR string, mode string) error {
	if runtime.GOOS == "windows" {
		if mode == "remove" {
			_ = exec.Command("schtasks", "/Delete", "/TN", windowsTaskName, "/F").Run()
			return nil
		}
		minutes := minutesFromCronString(timeStringInCronjob)
		return saveWindowsTask(minutes, WORKING_DIR)
	}

	// Initialize the path of txt.file
	cronTXTPath := filepath.Join(WORKING_DIR, "cronjob", "cron.txt")

	// IF the file exist, then delete it
	if _, err := os.Stat(cronTXTPath); err == nil {
		e := os.Remove(cronTXTPath)
		if e != nil {
			return e
		}
	}

	// Create a new file
	file, openFileErr := os.OpenFile(cronTXTPath, os.O_RDWR|os.O_CREATE, 0777)
	if openFileErr != nil {
		return openFileErr
	}
	defer file.Close()

	// List existing cronjobs
	crontabJobs, setupCronJobErr := exec.Command("crontab", "-l").Output()
	if setupCronJobErr != nil {
		return setupCronJobErr
	}

	if mode == "remove" {
		// Convert []byte to string array
		cronjobArray := strings.Split(string(crontabJobs), "\n")
		// Scanning target to remove
		executable, executableErr := os.Executable()
		if executableErr != nil {
			return executableErr
		}
		envEnvironment := "WORKING_DIR=" + WORKING_DIR
		// Remove the line that contain the scanning target
		for index, cronjob := range cronjobArray {
			if strings.Contains(cronjob, executable) || strings.Contains(cronjob, "scanning") {
				cronjobArray = append(cronjobArray[:index], cronjobArray[index+1:]...)
			}

			if strings.Contains(cronjob, envEnvironment) {
				cronjobArray = append(cronjobArray[:index], cronjobArray[index+1:]...)
			}
		}
		// Convert string array to []byte
		joinedString := strings.Join(cronjobArray, "\n")
		// Assign the []byte to the existing cronjobs
		crontabJobs = []byte(joinedString)

	}

	if mode == "add" {
		_, writeENV := file.WriteString("WORKING_DIR=" + WORKING_DIR + "\n")
		if writeENV != nil {
			return writeENV
		}

	}

	// Write existing cronjobs to the file
	_, writeExistingCronJobErr := file.Write(crontabJobs)
	if writeExistingCronJobErr != nil {
		return writeExistingCronJobErr
	}

	if mode == "add" {
		executable, executableErr := os.Executable()
		if executableErr != nil {
			return executableErr
		}
		// Create a new cronjob string
		cronjob := timeStringInCronjob + " \"" + executable + "\" >> /tmp/pulsewatch.out 2>> /tmp/pulsewatch.err" + "\n"
		// Write the new cronjob to the file
		_, writeNewCronJobErr := file.WriteString(cronjob)
		if writeNewCronJobErr != nil {
			return writeNewCronJobErr
		}

	}

	// Set up the cronjob to the system
	setupCronJob := exec.Command("crontab", cronTXTPath)
	setupCronJobErr = setupCronJob.Run()
	if setupCronJobErr != nil {
		return setupCronJobErr
	}

	// If no error then return nil
	return nil
}

// Set up cronjob
func SetUpCronJob(WORKING_DIR string) error {
	// Ask for time in minutes
	timeInMinutes := askForTimeInput()
	fmt.Printf("You chose %d minutes\n", timeInMinutes)
	if runtime.GOOS == "windows" {
		return saveWindowsTask(timeInMinutes, WORKING_DIR)
	}

	// Calculate time string in cronjob
	timeStringInCronjob := calculateTimeString(timeInMinutes)
	fmt.Printf("Your cronjob timestring is: %s\n", timeStringInCronjob)
	fmt.Println("Please allow the script to set the cronjob by clicking allow")

	// Save the cronjob
	err := SaveCronJob(timeStringInCronjob, WORKING_DIR, "add")

	return err

}

func minutesFromCronString(cronString string) int {
	if cronString == "0 0 * * *" {
		return 1440
	}
	if strings.HasPrefix(cronString, "0 ") {
		return 60
	}
	parts := strings.Fields(cronString)
	if len(parts) > 0 && strings.Contains(parts[0], "/") {
		value, _ := strconv.Atoi(strings.TrimPrefix(parts[0], "*/"))
		return value
	}
	return 1
}

func saveWindowsTask(minutes int, workingDir string) error {
	if minutes < 1 {
		minutes = 1
	}
	executable := filepath.Join(workingDir, "pulsewatch.exe")
	if _, err := os.Stat(executable); err != nil {
		executable, err = os.Executable()
		if err != nil {
			return err
		}
	}
	cronTXTPath := filepath.Join(workingDir, "cronjob", "cron.txt")
	if err := os.WriteFile(cronTXTPath, []byte(fmt.Sprintf("Windows Task Scheduler: every %d minute(s)\n", minutes)), 0644); err != nil {
		return err
	}
	return exec.Command("schtasks", "/Create", "/SC", "MINUTE", "/MO", strconv.Itoa(minutes), "/TN", windowsTaskName, "/TR", fmt.Sprintf("\"%s\"", executable), "/F").Run()
}
