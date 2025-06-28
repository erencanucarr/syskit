package cmd

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"
    "time"

    "syskit/internal/config"
    "syskit/internal/email"
    "syskit/internal/utils"
    "syskit/internal/i18n"

    "github.com/spf13/cobra"
)

var (
    pwKeywords   string
    pwCpuLimit   int
    pwRamLimit   int
)

var processWatchCmd = &cobra.Command{
    Use:   "process-watch",
    Short: "Watch for suspicious or resource-hungry processes",
    Run: func(cmd *cobra.Command, args []string) {
        runProcessWatch()
    },
}

func init() {
    processWatchCmd.Flags().StringVar(&pwKeywords, "alert-on", "crypto,mining,reverse_shell", "comma separated suspicious keywords")
    processWatchCmd.Flags().IntVar(&pwCpuLimit, "cpu", 50, "CPU % threshold")
    processWatchCmd.Flags().IntVar(&pwRamLimit, "ram", 30, "RAM % threshold")
}

func runProcessWatch() {
    keywords := []string{}
    for _, k := range strings.Split(pwKeywords, ",") {
        k = strings.TrimSpace(k)
        if k != "" {
            keywords = append(keywords, strings.ToLower(k))
        }
    }

    out, err := exec.Command("ps", "-eo", "pid,comm,%cpu,%mem", "--no-headers").Output()
    if err != nil {
        fmt.Println("ps error:", err)
        return
    }

    scanner := bufio.NewScanner(strings.NewReader(string(out)))
    var suspectRows [][]string
    for scanner.Scan() {
        fields := strings.Fields(scanner.Text())
        if len(fields) < 4 {
            continue
        }
        pid := fields[0]
        cmdName := fields[1]
        cpuStr := fields[2]
        memStr := fields[3]
        cpuVal, _ := strconv.ParseFloat(strings.TrimSpace(cpuStr), 64)
        memVal, _ := strconv.ParseFloat(strings.TrimSpace(memStr), 64)

        lower := strings.ToLower(cmdName)
        matchKeyword := false
        for _, kw := range keywords {
            if strings.Contains(lower, kw) {
                matchKeyword = true
                break
            }
        }
        matchResource := int(cpuVal) >= pwCpuLimit || int(memVal) >= pwRamLimit
        if matchKeyword || matchResource {
            suspectRows = append(suspectRows, []string{pid, cmdName, fmt.Sprintf("%.1f", cpuVal), fmt.Sprintf("%.1f", memVal)})
            logEntry(pid, cmdName, cpuVal, memVal)
        }
    }

    if len(suspectRows) == 0 {
        fmt.Println(i18n.T("no_suspicious"))
        return
    }

    utils.Print([]string{"PID", "CMD", "CPU%", "MEM%"}, suspectRows)

    // email alert
    cfg := config.Load()
    if cfg.SMTP.Host != "" {
        body := "Suspicious processes detected:\n"
        for _, r := range suspectRows {
            body += strings.Join(r, " ") + "\n"
        }
        email.Send(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.To, "Syskit process alert", body)
    }
}

func logEntry(pid, name string, cpu, mem float64) {
    home, _ := os.UserHomeDir()
    logPath := filepath.Join(home, ".syskit", "process_watch.log")
    os.MkdirAll(filepath.Dir(logPath), 0o755)
    f, _ := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
    defer f.Close()
    timestamp := time.Now().Format(time.RFC3339)
    fmt.Fprintf(f, "%s pid=%s cmd=%s cpu=%.1f mem=%.1f\n", timestamp, pid, name, cpu, mem)
}
