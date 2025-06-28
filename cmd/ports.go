package cmd

import (
	"os/exec"
	"strings"
	"syskit/internal/i18n"

	"syskit/internal/utils"

	"github.com/spf13/cobra"
)

var portFilter string

var portsCmd = &cobra.Command{
	Use:   "ports",
	Short: "List open ports",
	Run: func(cmd *cobra.Command, args []string) {
		printPorts()
	},
}

func init() {
	portsCmd.Flags().StringVarP(&portFilter, "filter", "f", "", "filter like LISTEN")
}

func printPorts() {
	cmd := exec.Command("ss", "-tuln")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	rows := [][]string{}
	for _, l := range lines[1:] { // skip header
		if portFilter != "" && !strings.Contains(l, portFilter) {
			continue
		}
		fields := strings.Fields(l)
		if len(fields) < 5 {
			continue
		}
		rows = append(rows, []string{fields[0], fields[3], fields[4]})
	}
	utils.Print([]string{i18n.T("proto"), i18n.T("local"), i18n.T("peer")}, rows)
}
