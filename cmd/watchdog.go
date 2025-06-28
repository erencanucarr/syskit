package cmd

import (
    "fmt"
    "os/exec"
    "strings"
    "syskit/internal/i18n"
    "time"

    "github.com/spf13/cobra"
)

var wdService string
var wdLoop bool

var watchdogCmd = &cobra.Command{
    Use:   "watchdog",
    Short: "Smart service watchdog with basic root-cause hints",
    Run: func(cmd *cobra.Command, args []string) {
        if wdService == "" {
            fmt.Println(i18n.T("service_required"))
            return
        }
        for {
            onceWatch(wdService)
            if !wdLoop {
                return
            }
            time.Sleep(30 * time.Second)
        }
    },
}

func init() {
    watchdogCmd.Flags().StringVar(&wdService, "service", "", "systemd service name to watch")
    watchdogCmd.Flags().BoolVar(&wdLoop, "loop", false, "run continuously (30s interval)")
}

func onceWatch(svc string) {
    if isActive(svc) {
        fmt.Println(fmt.Sprintf(i18n.T("service_ok"), svc))
        return
    }
    fmt.Println(fmt.Sprintf(i18n.T("restarting_service"), svc))
    exec.Command("systemctl", "restart", svc).Run()
    time.Sleep(3 * time.Second)
    if isActive(svc) {
        fmt.Println(i18n.T("restart_success"))
        return
    }
    fmt.Println(i18n.T("restart_failed"))

    hint := analyseFailure(svc)
    fmt.Println(i18n.T("possible_reason"), hint)
}

func isActive(svc string) bool {
    err := exec.Command("systemctl", "is-active", "--quiet", svc).Run()
    return err == nil
}

func analyseFailure(svc string) string {
    // check last 20 journal lines
    out, _ := exec.Command("journalctl", "-u", svc, "-n", "20", "--no-pager").Output()
    text := strings.ToLower(string(out))
    switch {
    case strings.Contains(text, "memory") || strings.Contains(text, "oom"):
        return "Service killed due to memory pressure – check RAM or limits"
    case strings.Contains(text, "permission"):
        return "Permission issues – verify file ownership/sudo"
    case strings.Contains(text, "config") || strings.Contains(text, "syntax"):
        // for nginx, run -t
        if svc == "nginx" {
            cfgOut, err := exec.Command("nginx", "-t").CombinedOutput()
            if err != nil {
                return "Config test failed: " + string(cfgOut)
            }
        }
        return "Possible configuration error – validate config files"
    default:
        return "See journalctl -u " + svc + " for details"
    }
}
