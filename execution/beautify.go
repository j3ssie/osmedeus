package execution

import (
	"encoding/csv"
	"os"

	"github.com/j3ssie/osmedeus/utils"
	"github.com/olekukonko/tablewriter"
)

func PrintCSV(filename string) {
	filename = utils.NormalizePath(filename)
	if !utils.FileExists(filename) {
		utils.ErrorF("File %v not found", filename)
		return
	}

	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		utils.ErrorF("%v", err)
		return
	}
	defer file.Close()
	// Create a new CSV reader
	reader := csv.NewReader(file)
	reader.LazyQuotes = true

	// Read all CSV records
	records, err := reader.ReadAll()
	if err != nil {
		utils.ErrorF("%v", err)
		return
	}

	// Create a new table
	table := tablewriter.NewWriter(os.Stdout)
	for _, record := range records {
		table.Append(record)
	}
	table.SetRowLine(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetColWidth(100)
	table.SetHeaderLine(true)
	table.SetAutoWrapText(true)
	table.Render()
}
