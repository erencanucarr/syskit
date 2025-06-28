package cmd

import (
    "errors"
    "fmt"
    "strings"

    "syskit/internal/schedule"
    "syskit/internal/utils"

    "github.com/spf13/cobra"
)

var scheduleCmd = &cobra.Command{Use: "schedule", Short: "Manage cron schedules"}

var scName, scCmd, scEvery, scAt, scOn, scCron string

var scheduleAddCmd = &cobra.Command{
    Use:   "add",
    Short: "Add a scheduled job",
    RunE: func(cmd *cobra.Command, args []string) error {
        if scName == "" || scCmd == "" {
            return errors.New("--name and --cmd required")
        }
        spec := scCron
        if spec == "" {
            spec = buildSpec()
        }
        if spec == "" {
            return errors.New("invalid schedule spec")
        }
        lines, _ := schedule.Read()
        // remove any existing with same name
        lines, _ = schedule.RemoveEntry(lines, scName)
        lines = schedule.AddEntry(lines, schedule.Entry{Name: scName, Spec: spec, Command: scCmd})
        if err := schedule.Write(lines); err != nil {
            return err
        }
        fmt.Println("added", scName)
        return nil
    },
}

var scheduleListCmd = &cobra.Command{
    Use:   "list",
    Short: "List scheduled jobs",
    RunE: func(cmd *cobra.Command, args []string) error {
        lines, _ := schedule.Read()
        entries := schedule.ParseEntries(lines)
        headers := []string{"NAME", "SCHEDULE", "COMMAND"}
        rows := [][]string{}
        for _, e := range entries {
            rows = append(rows, []string{e.Name, e.Spec, e.Command})
        }
        if len(rows) == 0 {
            fmt.Println("No jobs")
            return nil
        }
        utils.Print(headers, rows)
        return nil
    },
}

var scheduleRemoveCmd = &cobra.Command{
    Use:   "remove [name]",
    Short: "Remove scheduled job",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]
        lines, _ := schedule.Read()
        newLines, removed := schedule.RemoveEntry(lines, name)
        if !removed {
            return errors.New("not found")
        }
        return schedule.Write(newLines)
    },
}

func buildSpec() string {
    switch scEvery {
    case "day", "daily":
        if scAt == "" {
            scAt = "00:00"
        }
        parts := strings.Split(scAt, ":")
        if len(parts) != 2 {
            return ""
        }
        return fmt.Sprintf("%s %s * * *", parts[1], parts[0])
    case "hour":
        return fmt.Sprintf("0 * * * *")
    case "week", "weekly":
        if scOn == "" {
            scOn = "Sun"
        }
        if scAt == "" {
            scAt = "00:00"
        }
        parts := strings.Split(scAt, ":")
        if len(parts) != 2 {
            return ""
        }
        return fmt.Sprintf("%s %s * * %s", parts[1], parts[0], scOn[:3])
    }
    return ""
}

func init() {
    scheduleAddCmd.Flags().StringVar(&scName, "name", "", "job name")
    scheduleAddCmd.Flags().StringVar(&scCmd, "cmd", "", "command to run")
    scheduleAddCmd.Flags().StringVar(&scEvery, "every", "", "day|week|hour")
    scheduleAddCmd.Flags().StringVar(&scAt, "at", "", "HH:MM time")
    scheduleAddCmd.Flags().StringVar(&scOn, "on", "", "weekday (Mon..Sun)")
    scheduleAddCmd.Flags().StringVar(&scCron, "cron", "", "custom cron expression")

    scheduleCmd.AddCommand(scheduleAddCmd)
    scheduleCmd.AddCommand(scheduleListCmd)
    scheduleCmd.AddCommand(scheduleRemoveCmd)
}
