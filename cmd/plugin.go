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
    Long: `Plugins are standalone executables placed under $(HOME)/.syskit/plugins.
They are invoked automatically when you run:  syskit <plugin-name> [args...]
The CLI searches the plugin directory first and, if an executable with the given
name exists, it is executed instead of a built-in command.

Writing a plugin is trivial – any language that can write to STDOUT works.
A simple Bash example (save as hello, chmod +x):

  #!/usr/bin/env bash
  echo "Hello from Syskit plugin!"

A Go example:

  package main
  import "fmt"
  func main(){ fmt.Println("Hello from Syskit plugin") }

Plugins receive CLI arguments unchanged and should print their own help when
called with -h or --help.`,
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

var pluginCreateCmd = &cobra.Command{
    Use:   "create [name]",
    Short: "Scaffold a Go plugin in current directory",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]
        file := name + ".go"
        if _, err := os.Stat(file); err == nil {
            return fmt.Errorf("%s already exists", file)
        }
        tpl := fmt.Sprintf(`package main

import "fmt"

// %s plugin – replace with your logic.
func main() {
    fmt.Println("Hello from %s plugin!")
}
`, name, name)
        return os.WriteFile(file, []byte(tpl), 0o644)
    },
}

func init() {
    pluginCmd.AddCommand(pluginListCmd)
    pluginCmd.AddCommand(pluginInstallCmd)
    pluginCmd.AddCommand(pluginCreateCmd)
}
