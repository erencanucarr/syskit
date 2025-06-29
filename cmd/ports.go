package cmd

import (
    tea "github.com/charmbracelet/bubbletea"
    "syskit/internal/ui"

    "github.com/spf13/cobra"
)

var portFilter string
var watchPorts bool

var portsCmd = &cobra.Command{
    Use:   "ports",
    Short: "List open ports",
    Run: func(cmd *cobra.Command, args []string) {
        m := ui.NewPortsModel(portFilter, watchPorts)
        p := tea.NewProgram(m)
        _ = p.Start()
    },
}

func init() {
    portsCmd.Flags().StringVarP(&portFilter, "filter", "f", "", "filter substring (e.g. LISTEN)")
    portsCmd.Flags().BoolVarP(&watchPorts, "watch", "w", false, "watch mode (refresh)")
}
