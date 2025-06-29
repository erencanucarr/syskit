package cmd

import (
    "fmt"
    "sort"
    "strings"
    "time"

    "syskit/internal/ui"

    "github.com/spf13/cobra"
)

type event struct {
    When     time.Time
    Src      string
    Category string
    Msg      string
}

// gatherEvents collects logs for last 24h (same logic as previous implementation). Implementation kept in separate file.

func buildTimelineGroups(evs []event) map[string][]string {
    for i := range evs {
        e := &evs[i]
        switch e.Src {
        case "auth":
            if strings.Contains(e.Msg, "Failed password") {
                e.Category = "auth-failed"
            } else if strings.Contains(e.Msg, "Accepted password") {
                e.Category = "auth-ok"
            } else {
                e.Category = "auth-other"
            }
        case "dmesg":
            lower := strings.ToLower(e.Msg)
            if strings.Contains(lower, "error") {
                e.Category = "kernel-err"
            } else if strings.Contains(lower, "warn") {
                e.Category = "kernel-warn"
            } else {
                e.Category = "dmesg"
            }
        default:
            e.Category = e.Src
        }
    }
    groups := make(map[string][]string)
    for _, e := range evs {
        groups[e.Category] = append(groups[e.Category], fmt.Sprintf("[%s] %s", e.When.Format("15:04"), e.Msg))
    }
    // sort each slice chronologically
    for _, list := range groups {
        sort.Strings(list)
    }
    return groups
}

var timelineCmd = &cobra.Command{
    Use:   "timeline",
    Short: "Show system event timeline (24h)",
    RunE: func(cmd *cobra.Command, args []string) error {
        evs := gatherEvents()
        groups := buildTimelineGroups(evs)
        return ui.RunTimeline(groups)
    },
}
