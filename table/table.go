package table

import (
	"os"

	"github.com/adarshschizo/pulsewatch-network-monitor/network"

	"github.com/jedib0t/go-pretty/table"
)

// Function to print out beautiful table along with network consumption chart
func PrintNetworkingTable(networkDataMap map[string]network.NetworkData, keyDesc []string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	// Set the header
	t.AppendHeader(table.Row{"Process Name", "Incoming data (MB)", "Outgoing data (MB)", "Time"})

	// Key is sorted by MBIn (incoming network)
	for _, processName := range keyDesc {
		data := networkDataMap[processName]
		dataLength := len(data.Time)
		if dataLength == 0 || len(data.ReceivedMB) < dataLength || len(data.SentMB) < dataLength {
			continue
		}

		// Get the latest network in data
		MBIn := data.ReceivedMB[dataLength-1]
		// Get the latest network out data
		MBOut := data.SentMB[dataLength-1]
		// Get the latest time recorded
		Time := data.Time[dataLength-1]
		// Append them all to the row
		t.AppendRow(table.Row{processName, MBIn, MBOut, Time})
	}
	t.AppendFooter(table.Row{"Table is sorted by Incoming network"})
	// Set auto index
	t.SetAutoIndex(true)
	// Set style
	t.SetStyle(table.StyleColoredBlackOnMagentaWhite)
	// Render the table to the terminal
	t.Render()
}
