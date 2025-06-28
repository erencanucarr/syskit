package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "syskit/internal/config"
)

var langCmd = &cobra.Command{
    Use:   "lang <code>",
    Short: "Set default language code (e.g., en, tr, de, es)",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        code := args[0]
        cfg := config.Load()
        cfg.Lang = code
        if err := config.Save(); err != nil {
            fmt.Println("Failed to save config:", err)
            return
        }
        fmt.Printf("Default language set to %s\n", code)
    },
}
