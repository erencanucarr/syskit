package cmd

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "syskit/internal/utils"

    "github.com/spf13/cobra"
)

var pluginCmd = &cobra.Command{
    Use:   "plugin",
    Short: "Manage syskit plugins",
}

var pluginListCmd = &cobra.Command{
    Use:   "list",
    Short: "List installed plugins",
    Run: func(cmd *cobra.Command, args []string) {
        dir := PluginDir()
        entries, _ := os.ReadDir(dir)
        headers := []string{"Plugin"}
        rows := [][]string{}
        for _, e := range entries {
            if !e.IsDir() {
                rows = append(rows, []string{e.Name()})
            }
        }
        if len(rows) == 0 {
            fmt.Println("No plugins found")
            return
        }
        utils.Print(headers, rows)
    },
}

var pluginInstallCmd = &cobra.Command{
    Use:   "install [file]",
    Short: "Install a plugin binary or script",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        src := args[0]
        dst := filepath.Join(PluginDir(), filepath.Base(src))

        if err := copyFile(src, dst); err != nil {
            fmt.Println("install failed:", err)
            return
        }
        os.Chmod(dst, 0o755)
        fmt.Println("installed", dst)
    },
}

func PluginDir() string {
    home, _ := os.UserHomeDir()
    dir := filepath.Join(home, ".syskit", "plugins")
    os.MkdirAll(dir, 0o755)
    return dir
}

func copyFile(src, dst string) error {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()
    if _, err = io.Copy(out, in); err != nil {
        return err
    }
    return nil
}

func init() {
    pluginCmd.AddCommand(pluginListCmd)
    pluginCmd.AddCommand(pluginInstallCmd)
}
