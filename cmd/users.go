package cmd

import (
    tea "github.com/charmbracelet/bubbletea"
    "syskit/internal/ui"

    "github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Show active users",
	Run: func(cmd *cobra.Command, args []string) {
        m := ui.NewUsersModel(false)
        p := tea.NewProgram(m)
        _ = p.Start()
    },
}
