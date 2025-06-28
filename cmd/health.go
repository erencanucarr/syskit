package cmd

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "runtime"
    "strconv"
    "strings"

    "syskit/internal/config"
    "syskit/internal/email"
    "syskit/internal/utils"

    "github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
    Use:   "health",
    Short: "Check system health and alert if thresholds exceeded",
    Run: func(cmd *cobra.Command, args []string) {
        checkHealth()
    },
}

func init() {
    // registered from root.go later
}

func checkHealth() {
    cfg := config.Load()
    exceeded := []string{}

    cpuPerc := cpuUsagePercent()
    if cpuPerc >= cfg.Thresholds.CPU {
        exceeded = append(exceeded, fmt.Sprintf("CPU %d%%", cpuPerc))
    }

    ramPerc := ramUsagePercent()
    if ramPerc >= cfg.Thresholds.RAM {
        exceeded = append(exceeded, fmt.Sprintf("RAM %d%%", ramPerc))
    }

    diskPerc := diskUsagePercent()
    if diskPerc >= cfg.Thresholds.Disk {
        exceeded = append(exceeded, fmt.Sprintf("Disk %d%%", diskPerc))
    }

    headers := []string{"Resource", "Usage%", "Threshold%"}
    rows := [][]string{
        {"CPU", fmt.Sprintf("%d", cpuPerc), fmt.Sprintf("%d", cfg.Thresholds.CPU)},
        {"RAM", fmt.Sprintf("%d", ramPerc), fmt.Sprintf("%d", cfg.Thresholds.RAM)},
        {"Disk", fmt.Sprintf("%d", diskPerc), fmt.Sprintf("%d", cfg.Thresholds.Disk)},
    }
    utils.Print(headers, rows)

    if len(exceeded) == 0 {
        fmt.Println("All OK")
        return
    }
    subject := "Syskit alert: " + strings.Join(exceeded, ", ")
    body := "The following resources exceeded thresholds:\n" + strings.Join(exceeded, "\n")

    if cfg.SMTP.Host != "" {
        if err := email.Send(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.To, subject, body); err != nil {
            fmt.Println("Email send failed:", err)
        } else {
            fmt.Println("Alert email sent")
        }
    }
}

// simplistic average 1 min load divided by cores, *100
func cpuUsagePercent() int {
    b, err := os.ReadFile("/proc/loadavg")
    if err != nil {
        return 0
    }
    parts := strings.Fields(string(b))
    load, _ := strconv.ParseFloat(parts[0], 64)
    perc := int(load / float64(runtime.NumCPU()) * 100)
    return perc
}

func ramUsagePercent() int {
    f, err := os.Open("/proc/meminfo")
    if err != nil {
        return 0
    }
    defer f.Close()
    var total, free int64
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        var key string
        var val int64
        fmt.Sscanf(scanner.Text(), "%s %d", &key, &val)
        switch key {
        case "MemTotal:":
            total = val
        case "MemAvailable:":
            free = val
        }
    }
    used := total - free
    return int(float64(used) / float64(total) * 100)
}

func diskUsagePercent() int {
    out, err := exec.Command("df", "/").Output()
    if err != nil {
        return 0
    }
    lines := strings.Split(strings.TrimSpace(string(out)), "\n")
    if len(lines) < 2 {
        return 0
    }
    fields := strings.Fields(lines[1])
    if len(fields) >= 5 {
        p := strings.TrimRight(fields[4], "%")
        val, _ := strconv.Atoi(p)
        return val
    }
    return 0
}
