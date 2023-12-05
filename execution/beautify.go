package execution

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/j3ssie/osmedeus/utils"
	"github.com/olekukonko/tablewriter"
)

func PrintCSV(filename string) {
	records, err := readCSV(filename)
	if err != nil {
		return
	}

	// Create a new table
	table := tablewriter.NewWriter(os.Stdout)
	for index, record := range records {
		if index == 0 {
			table.SetHeader(record)
			continue
		}
		table.Append(record)
	}
	table.SetRowLine(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetColWidth(100)
	table.SetHeaderLine(true)
	table.SetAutoWrapText(true)
	table.Render()
}

func BeautifyCSV(filename string, dest string) {
	records, err := readCSV(filename)
	if err != nil {
		return
	}

	// Create a new table
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	for index, record := range records {
		if index == 0 {
			table.SetHeader(record)
			continue
		}
		table.Append(record)
	}
	table.SetRowLine(false)
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: true})
	table.SetColWidth(100)
	table.SetHeaderLine(true)
	table.SetAutoWrapText(true)
	table.Render()

	tableOutput := buf.String()
	// write to file
	utils.WriteToFile(dest, tableOutput)
}

func readCSV(filename string) ([][]string, error) {
	filename = utils.NormalizePath(filename)
	if !utils.FileExists(filename) {
		utils.ErrorF("File %v not found", filename)
		return nil, fmt.Errorf("File %v not found", filename)
	}

	// Open the CSV file
	file, err := os.Open(filename)
	if err != nil {
		utils.ErrorF("%v", err)
		return nil, err
	}
	defer file.Close()
	// Create a new CSV reader
	reader := csv.NewReader(file)
	reader.LazyQuotes = true

	// Read all CSV records
	records, err := reader.ReadAll()
	if err != nil {
		utils.ErrorF("%v", err)
		return nil, err
	}
	return records, nil
}
