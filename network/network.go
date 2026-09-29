package network

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

type NetworkData struct {
	ProcessName string
	ReceivedMB  []string
	SentMB      []string
	Time        []string
}

// Function to record Network Data to file
func RecordNetworkData(WORKING_DIR string) error {

	workingDirReport := filepath.Join(WORKING_DIR, "network", "network.txt")
	if err := os.MkdirAll(filepath.Dir(workingDirReport), 0755); err != nil {
		return err
	}

	// Open the file for appending
	file, openFileErr := os.OpenFile(workingDirReport, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openFileErr != nil {
		return openFileErr
	}
	defer file.Close()

	if runtime.GOOS == "windows" {
		return recordWindowsNetworkData(file)
	}

	// Scan using nettop
	networkcmd, err := exec.Command("nettop", "-l", "1", "-P", "-x").Output()
	if err != nil {
		return err
	}
	networkcmd = []byte(strings.TrimSpace(string(networkcmd)))

	lines := strings.Split(string(networkcmd), "\n")
	lines = lines[1:]

	// Using regex to extract data
	re := regexp.MustCompile(`\d+:\d+:\d+\.\d+\s+([\w\s\(\)\.]+)\.(\d+)\s+(\d+)\s+(\d+)`)

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) != 5 {
			continue
		}

		// Get the network consumption in byte
		receivedByte, err := strconv.ParseFloat(matches[3], 64)
		if err != nil {
			fmt.Println(err)
		}

		// Get the sent network in byte
		sentByte, err := strconv.ParseFloat(matches[4], 64)
		if err != nil {
			fmt.Println(err)
		}

		// Convert from byte to MB
		receivedMB := receivedByte / float64(1000000)
		sentMB := sentByte / float64(1000000)

		// Convert to string
		receivedMBString := fmt.Sprintf("%.5f", receivedMB)
		sentMBString := fmt.Sprintf("%.5f", sentMB)

		finalResult := matches[1] + "." + matches[2] + " | " + receivedMBString + " | " + sentMBString + " | " + time.Now().Format("2006-01-02 15:04:05")

		// Write the string to the file
		_, writeFileErr := file.WriteString(finalResult + "\n")
		if writeFileErr != nil {
			return writeFileErr

		}
	}

	return nil
}

func recordWindowsNetworkData(file *os.File) error {
	output, err := exec.Command("powershell", "-NoProfile", "-Command", "Get-NetAdapterStatistics | ForEach-Object { \"$($_.ReceivedBytes),$($_.SentBytes)\" }").Output()
	if err != nil {
		return err
	}

	var receivedBytes, sentBytes float64
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		values := strings.Split(strings.TrimSpace(line), ",")
		if len(values) != 2 {
			continue
		}
		received, receivedErr := strconv.ParseFloat(strings.TrimSpace(values[0]), 64)
		sent, sentErr := strconv.ParseFloat(strings.TrimSpace(values[1]), 64)
		if receivedErr == nil && sentErr == nil {
			receivedBytes += received
			sentBytes += sent
		}
	}

	if receivedBytes == 0 && sentBytes == 0 {
		return fmt.Errorf("no Windows network adapter statistics were returned")
	}
	result := fmt.Sprintf("Windows network | %.5f | %.5f | %s\n", receivedBytes/1000000, sentBytes/1000000, time.Now().Format("2006-01-02 15:04:05"))
	_, err = file.WriteString(result)
	return err
}

// Function to read the network report and return the stats, ready for chart building
func ReadNetworkData(WORKING_DIR string) (networkMap map[string]NetworkData, err error) {
	filePath := filepath.Join(WORKING_DIR, "network", "network.txt")

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Get all the lines in slice format
	var networkDataMap = make(map[string]NetworkData)

	for _, line := range strings.Split(strings.TrimSpace(string(file)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		slice := strings.Split(line, " | ")
		if len(slice) != 4 {
			continue
		}
		processName := slice[0]

		// If the process is already in the map, update its network data
		if networkData, ok := networkDataMap[processName]; ok {
			// Update existing slice
			receivedMB := append(networkData.ReceivedMB, slice[1])
			sentMB := append(networkData.SentMB, slice[2])
			time := append(networkData.Time, slice[3])

			// Replace the old network data by a new network data
			networkDataMap[processName] = NetworkData{
				ProcessName: processName,
				ReceivedMB:  receivedMB,
				SentMB:      sentMB,
				Time:        time,
			}
			// If the process is not in the map, add it
		} else {
			networkDataMap[processName] = NetworkData{
				ProcessName: processName,
				ReceivedMB:  []string{slice[1]},
				SentMB:      []string{slice[2]},
				Time:        []string{slice[3]},
			}
		}
	}

	networkDataMap = RemoveUnactivatedNetworkData(networkDataMap)

	return networkDataMap, nil

}

// Sort the map in descending order
func SortNetworkDataMap(networkDataMap map[string]NetworkData, sortedByReceivedData bool) (keysSortedInDesc []string) {

	// Initialize a slice of string containing all the keys sorted in descending order
	keysDesc := make([]string, 0, len(networkDataMap))

	// Add all the key of the map to the slice
	for key := range networkDataMap {
		keysDesc = append(keysDesc, key)
	}

	// Sorted the key based on the requirement
	sort.SliceStable(keysDesc, func(i, j int) bool {
		var (
			totalMBI float64
			totalMBJ float64
			MBSliceI []string
			MBSliceJ []string
		)

		// if sorted by incoming network is true, sort by incoming network
		if sortedByReceivedData {
			MBSliceI = networkDataMap[keysDesc[i]].ReceivedMB
			MBSliceJ = networkDataMap[keysDesc[j]].ReceivedMB
			// if sorted by incoming network is false, sort by outgoing network
		} else {
			MBSliceI = networkDataMap[keysDesc[i]].SentMB
			MBSliceJ = networkDataMap[keysDesc[j]].SentMB
		}

		for _, value := range MBSliceI {
			valueFloat, _ := strconv.ParseFloat(value, 64)
			totalMBI += valueFloat
		}

		for _, value := range MBSliceJ {
			valueFloat, _ := strconv.ParseFloat(value, 64)
			totalMBJ += valueFloat
		}

		avgMBI := totalMBI / float64(len(MBSliceI))
		avgMBJ := totalMBJ / float64(len(MBSliceJ))

		return avgMBI > avgMBJ
	})

	return keysDesc

}

// Function to get the top N keys in descending order
func GetTopDesc(keysSorted []string, topNumber int) (topKeysInDesc []string) {
	if topNumber > len(keysSorted) {
		topNumber = len(keysSorted)
	}
	if topNumber < 0 {
		topNumber = 0
	}
	topKeys := make([]string, 0, topNumber)
	for i := 0; i < topNumber; i++ {
		topKeys = append(topKeys, keysSorted[i])
	}

	return topKeys

}

func RemoveUnactivatedNetworkData(networkDataMap map[string]NetworkData) (networkDataMapCleaned map[string]NetworkData) {

	for key, networkdata := range networkDataMap {
		allZeroSent := CheckFullZero(networkdata.SentMB)
		allZeroReceived := CheckFullZero(networkdata.ReceivedMB)
		if allZeroSent && allZeroReceived {
			delete(networkDataMap, key)
		}
	}

	return networkDataMap

}

func CheckFullZero(slice []string) bool {
	for _, value := range slice {
		if value != "0.00000" {
			return false
		}
	}
	return true
}

// Some processes might have different length of time, this function will make them have the same length
func EqualizeTopKey(networkDataMap map[string]NetworkData, TopDesc []string, processNameLongestTime string) (networkDataMapCleaned map[string]NetworkData) {

	for index, time := range networkDataMap[processNameLongestTime].Time {
		for _, processName := range TopDesc {
			if index == len(networkDataMap[processName].Time) {
				networkData := networkDataMap[processName]
				networkData.Time = slices.Insert(networkData.Time, index, time)
				networkData.ReceivedMB = slices.Insert(networkData.ReceivedMB, index, "0.00000")
				networkData.SentMB = slices.Insert(networkData.SentMB, index, "0.00000")
				networkDataMap[processName] = networkData

			}
			if time != networkDataMap[processName].Time[index] {
				networkData := networkDataMap[processName]
				networkData.Time = slices.Insert(networkData.Time, index, time)
				networkData.ReceivedMB = slices.Insert(networkData.ReceivedMB, index, "0.00000")
				networkData.SentMB = slices.Insert(networkData.SentMB, index, "0.00000")
				networkDataMap[processName] = networkData
			}

		}

	}

	return networkDataMap

}

// Find the processName that has the longest time slice
func FindLongestTime(TopDesc []string, networkDataMap map[string]NetworkData) (processName string) {
	if len(TopDesc) == 0 {
		return ""
	}
	longest := 0
	for index := 1; index < len(TopDesc); index++ {
		if len(networkDataMap[TopDesc[index]].Time) > len(networkDataMap[TopDesc[longest]].Time) {
			longest = index
		}
	}

	return TopDesc[longest]

}
