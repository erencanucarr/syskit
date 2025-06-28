package cmd

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/spf13/cobra"
)

var timelineCmd = &cobra.Command{
    Use:   "timeline",
    Short: "Show last 24h system events (journal, auth, dmesg, audit)",
    Run: func(cmd *cobra.Command, args []string) {
        events := gatherEvents()
        sort.Slice(events, func(i, j int) bool { return events[i].When.Before(events[j].When) })
        for _, e := range events {
            fmt.Printf("[%s] %s\n", e.When.Format("15:04"), e.Msg)
        }
    },
}

type event struct {
    When time.Time
    Msg  string
}

func gatherEvents() []event {
    cutoff := time.Now().Add(-24 * time.Hour)
    var evs []event

    // journalctl
    if out, err := exec.Command("journalctl", "--since", "24h", "--no-pager", "--output", "short-iso").Output(); err == nil {
        scanner := bufio.NewScanner(strings.NewReader(string(out)))
        for scanner.Scan() {
            line := scanner.Text()
            if len(line) < 20 {
                continue
            }
            tsStr := line[:19]
            if ts, err := time.Parse("2006-01-02T15:04:05", tsStr); err == nil {
                if ts.After(cutoff) {
                    evs = append(evs, event{When: ts, Msg: line[20:]})
                }
            }
        }
    }

    // auth.log
    authPath := "/var/log/auth.log"
    if lines := readLogFile(authPath); len(lines) > 0 {
        locNow := time.Now()
        for _, l := range lines {
            if len(l) < 15 {
                continue
            }
            // auth.log format: "Jan  2 15:04:05 hostname ..."
            tsStr := l[:15]
            ts, err := time.Parse("Jan 2 15:04:05", tsStr)
            if err != nil {
                continue
            }
            // add current year
            ts = ts.AddDate(locNow.Year(), 0, 0)
            if ts.After(cutoff) {
                evs = append(evs, event{When: ts, Msg: l[16:]})
            }
        }
    }

    // dmesg (iso)
    if out, err := exec.Command("dmesg", "--time-format=iso").Output(); err == nil {
        scanner := bufio.NewScanner(strings.NewReader(string(out)))
        for scanner.Scan() {
            line := scanner.Text()
            if len(line) < 24 || line[0] != '[' {
                continue
            }
            // pattern: [2025-06-28T15:04:05] msg
            idx := strings.Index(line, "]")
            if idx == -1 {
                continue
            }
            tsStr := line[1:idx]
            ts, err := time.Parse("2006-01-02T15:04:05", tsStr)
            if err != nil {
                continue
            }
            if ts.After(cutoff) {
                evs = append(evs, event{When: ts, Msg: line[idx+2:]})
            }
        }
    }

    // audit.log
    auditPath := "/var/log/audit/audit.log"
    if lines := readLogFile(auditPath); len(lines) > 0 {
        for _, l := range lines {
            if !strings.HasPrefix(l, "type=") {
                continue
            }
            // audit lines: msg=audit(1625930275.123:12345)
            if idx := strings.Index(l, "audit("); idx != -1 {
                start := idx + 6
                end := strings.Index(l[start:], ")")
                if end != -1 {
                    fields := strings.Split(l[start:start+end], ":")
                    parts := strings.Split(fields[0], ".")
                    epoch, _ := strconv.ParseInt(parts[0], 10, 64)
                    ts := time.Unix(epoch, 0)
                    if ts.After(cutoff) {
                        evs = append(evs, event{When: ts, Msg: l})
                    }
                }
            }
        }
    }

    return evs
}

func readLogFile(path string) []string {
    f, err := os.Open(path)
    if err != nil {
        return nil
    }
    defer f.Close()
    var lines []string
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }
    return lines
}
