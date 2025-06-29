package cmd

import (
    "errors"
    "fmt"
    "strings"

    "syskit/internal/container"
    "syskit/internal/utils"

    "github.com/spf13/cobra"
)

var containersCmd = &cobra.Command{
    Use:   "containers",
    Short: "Manage Docker / Podman containers",
}

var containersListCmd = &cobra.Command{
    Use:   "list",
    Short: "List running containers",
    RunE: func(cmd *cobra.Command, args []string) error {
        list, err := container.List()
        if err != nil {
            return err
        }
        headers := []string{"ID", "NAME", "IMAGE", "STATUS", "PORTS"}
        rows := [][]string{}
        for _, c := range list {
            rows = append(rows, []string{c.ID, c.Name, c.Image, c.Status, strings.Join(c.Ports, ",")})
        }
        if len(rows) == 0 {
            fmt.Println("No containers")
            return nil
        }
        utils.Print(headers, rows)
        return nil
    },
}

var pruneAll bool

var containersPruneCmd = &cobra.Command{
    Use:   "prune",
    Short: "Remove stopped containers",
    RunE: func(cmd *cobra.Command, args []string) error {
        if !pruneAll {
            return errors.New("use --all to confirm prune")
        }
        return container.Prune()
    },
}

func init() {
    containersPruneCmd.Flags().BoolVar(&pruneAll, "all", false, "confirm prune")

    containersCmd.AddCommand(containersListCmd)
    containersCmd.AddCommand(containersPruneCmd)
}
