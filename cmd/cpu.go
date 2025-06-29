package cmd

import (
    tea "github.com/charmbracelet/bubbletea"
    "syskit/internal/ui"

    "github.com/spf13/cobra"
)

var watchCPU bool

var cpuCmd = &cobra.Command{
	Use:   "cpu",
	Short: "Display CPU information",
	Run: func(cmd *cobra.Command, args []string) {
        m := ui.NewCpuModel(watchCPU)
        p := tea.NewProgram(m)
        _ = p.Start()
    },
}

func init() {
    cpuCmd.Flags().BoolVarP(&watchCPU, "watch", "w", false, "watch mode")
}
