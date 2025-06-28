package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
	"syskit/internal/i18n"
)

var format = "table"

// SetFormat sets global output format
func SetFormat(f string) { format = f }

// PrintTable prints 2D string slice as table
func PrintTable(headers []string, rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	for _, r := range rows {
		table.Append(r)
	}
	table.Render()
}

// Print prints data in given format
func Print(headers []string, rows [][]string) {
    for i, h := range headers {
        headers[i] = i18n.T(h)
    }
	switch format {
	case "json":
		out := make([]map[string]string, 0, len(rows))
		for _, r := range rows {
			rowMap := map[string]string{}
			for i, h := range headers {
				rowMap[h] = r[i]
			}
			out = append(out, rowMap)
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
	case "yaml":
		out := make([]map[string]string, 0, len(rows))
		for _, r := range rows {
			rowMap := map[string]string{}
			for i, h := range headers {
				rowMap[h] = r[i]
			}
			out = append(out, rowMap)
		}
		b, _ := yaml.Marshal(out)
		fmt.Println(string(b))
	default:
		PrintTable(headers, rows)
	}
}
