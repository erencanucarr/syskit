package cmd

import (
    tea "github.com/charmbracelet/bubbletea"
    "syskit/internal/ui"

    "github.com/spf13/cobra"
)

var watchMem bool

var memCmd = &cobra.Command{
    Use:   "mem",
    Short: "Display memory usage",
    Run: func(cmd *cobra.Command, args []string) {
        m := ui.NewMemModel(watchMem)
        p := tea.NewProgram(m)
        _ = p.Start()
    },
}

func init() {
    memCmd.Flags().BoolVarP(&watchMem, "watch", "w", false, "watch mode")
}
