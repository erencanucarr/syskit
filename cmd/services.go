package cmd

import (
    "fmt"

    "syskit/internal/service"
    "syskit/internal/utils"

    "github.com/spf13/cobra"
)

var servicesCmd = &cobra.Command{
    Use:   "services",
    Short: "Manage systemd services (Linux)",
}

var servicesListCmd = &cobra.Command{
    Use:   "list",
    Short: "List services and status",
    RunE: func(cmd *cobra.Command, args []string) error {
        units, err := service.List()
        if err != nil {
            return err
        }
        headers := []string{"NAME", "LOAD", "ACTIVE", "SUB", "DESCRIPTION"}
        rows := [][]string{}
        for _, u := range units {
            rows = append(rows, []string{u.Name, u.Load, u.Active, u.Sub, u.Description})
        }
        if len(rows) == 0 {
            fmt.Println("No services found")
            return nil
        }
        utils.Print(headers, rows)
        return nil
    },
}

func newServiceControlCmd(action string) *cobra.Command {
    return &cobra.Command{
        Use:   fmt.Sprintf("%s [name]", action),
        Short: fmt.Sprintf("%s a service", action),
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            name := args[0]
            if err := service.Control(name, action); err != nil {
                return err
            }
            fmt.Printf("%s %s\n", action, name)
            return nil
        },
    }
}

func init() {
    servicesCmd.AddCommand(servicesListCmd)
    servicesCmd.AddCommand(newServiceControlCmd("start"))
    servicesCmd.AddCommand(newServiceControlCmd("stop"))
    servicesCmd.AddCommand(newServiceControlCmd("restart"))
}
