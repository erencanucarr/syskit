package cmd

import (
    "github.com/spf13/cobra"
    tea "github.com/charmbracelet/bubbletea"
    "syskit/internal/pulse"
)

var pulseCmd = &cobra.Command{
    Use:   "pulse",
    Short: "Real-time terminal dashboard",
    Run: func(cmd *cobra.Command, args []string) {
        p := tea.NewProgram(pulse.Initial())
        p.Run()
    },
}
