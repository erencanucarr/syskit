package cmd

import (
	"os/exec"
	"strings"
	"syskit/internal/i18n"

	"syskit/internal/utils"

	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Show active users",
	Run: func(cmd *cobra.Command, args []string) {
		showUsers()
	},
}

func showUsers() {
	out, err := exec.Command("who").Output()
	if err != nil {
		return
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	rows := [][]string{}
	for _, l := range lines {
		fields := strings.Fields(l)
		if len(fields) >= 2 {
			rows = append(rows, []string{fields[0], fields[1]})
		}
	}
	utils.Print([]string{i18n.T("user"), i18n.T("terminal")}, rows)
}
