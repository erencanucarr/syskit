package cmd

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "os/exec"
    "runtime"
    "strconv"
    "strings"
    "time"

    "syskit/internal/utils"

    "github.com/spf13/cobra"
)

var (
    infoQuiet   bool
    infoVerbose bool
)

var infoCmd = &cobra.Command{
    Use:   "info",
    Short: "Display system information",
    Run: func(cmd *cobra.Command, args []string) {
        printBasic()
        if infoVerbose {
            printDetailed()
        }
    },
}

func init() {
    infoCmd.Flags().BoolVarP(&infoQuiet, "quiet", "q", false, "only show basic info")
    infoCmd.Flags().BoolVarP(&infoVerbose, "verbose", "v", false, "show detailed info")
}

func printBasic() {
    if infoQuiet {
        headers := []string{"Hostname", "OS", "Kernel", "Uptime"}
        rows := [][]string{{hostname(), osRelease(), kernel(), uptime()}}
        utils.Print(headers, rows)
        return
    }

    fmt.Println("# Basic System Information")
    utils.Print([]string{"Hostname", "OS"}, [][]string{{hostname(), osRelease()}})

    utils.Print([]string{"Kernel", "Uptime"}, [][]string{{kernel(), uptime()}})

    cpuModel, cores, freq, load := cpuInfo()
    utils.Print([]string{"CPU", "Cores", "Freq", "Load"}, [][]string{{cpuModel, cores, freq, load}})

    total, avail, swap := memInfo()
    utils.Print([]string{"MemTotal", "MemAvail", "SwapUsed"}, [][]string{{total, avail, swap}})

    size, used := rootDisk()
    utils.Print([]string{"Disk", "Used%"}, [][]string{{size, used}})

    gpu := gpuInfo()
    utils.Print([]string{"GPU"}, [][]string{{gpu}})

    ifaceRows := netInfo()
    utils.Print([]string{"Iface", "IP"}, ifaceRows)
}

func printDetailed() {
    fmt.Println("\n# Detailed Info")
    // placeholder: could extend with more detailed sections
}

func hostname() string {
    h, _ := os.Hostname()
    return h
}

func osRelease() string {
    f, err := os.Open("/etc/os-release")
    if err != nil {
        return "-"
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    var name, ver string
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "NAME=") && name == "" {
            name = strings.Trim(line[5:], "\"")
        }
        if strings.HasPrefix(line, "VERSION_ID=") && ver == "" {
            ver = strings.Trim(line[11:], "\"")
        }
    }
    return fmt.Sprintf("%s %s", name, ver)
}

func kernel() string {
    out, err := exec.Command("uname", "-r").Output()
    if err != nil {
        return "-"
    }
    return strings.TrimSpace(string(out))
}

func uptime() string {
    b, err := os.ReadFile("/proc/uptime")
    if err != nil {
        return "-"
    }
    parts := strings.Fields(string(b))
    sec, _ := strconv.ParseFloat(parts[0], 64)
    d := time.Duration(sec) * time.Second
    return d.Truncate(time.Minute).String()
}

func cpuInfo() (model, cores, freq, load string) {
    f, err := os.Open("/proc/cpuinfo")
    if err == nil {
        scanner := bufio.NewScanner(f)
        for scanner.Scan() {
            line := scanner.Text()
            if strings.HasPrefix(line, "model name") && model == "" {
                parts := strings.SplitN(line, ":", 2)
                if len(parts) == 2 {
                    model = strings.TrimSpace(parts[1])
                }
            }
            if strings.HasPrefix(line, "cpu MHz") && freq == "" {
                parts := strings.SplitN(line, ":", 2)
                if len(parts) == 2 {
                    freq = fmt.Sprintf("%s MHz", strings.TrimSpace(parts[1]))
                }
            }
        }
        f.Close()
    }
    cores = fmt.Sprintf("%d", runtime.NumCPU())

    b, _ := os.ReadFile("/proc/loadavg")
    load = strings.Join(strings.Fields(string(b))[:3], " ")
    return
}

func memInfo() (total, avail, swapUsed string) {
    f, err := os.Open("/proc/meminfo")
    if err != nil {
        return
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    var t, a, st, sf int64
    for scanner.Scan() {
        var key string
        var val int64
        fmt.Sscanf(scanner.Text(), "%s %d", &key, &val)
        switch key {
        case "MemTotal:":
            t = val
        case "MemAvailable:":
            a = val
        case "SwapTotal:":
            st = val
        case "SwapFree:":
            sf = val
        }
    }
    total = fmt.Sprintf("%.1f GB", float64(t)/1048576)
    avail = fmt.Sprintf("%.1f GB", float64(a)/1048576)
    if st > 0 {
        swapUsed = fmt.Sprintf("%.1f GB", float64(st-sf)/1048576)
    } else {
        swapUsed = "0 GB"
    }
    return
}

func rootDisk() (size, used string) {
    out, err := exec.Command("df", "-h", "/").Output()
    if err != nil {
        return
    }
    lines := strings.Split(strings.TrimSpace(string(out)), "\n")
    if len(lines) < 2 {
        return
    }
    fields := strings.Fields(lines[1])
    if len(fields) >= 5 {
        size = fields[1]
        used = fields[4]
    }
    return
}

func gpuInfo() string {
    out, err := exec.Command("lspci").Output()
    if err != nil {
        return "-"
    }
    for _, line := range strings.Split(string(out), "\n") {
        if strings.Contains(line, "VGA") {
            return strings.TrimSpace(line)
        }
    }
    return "-"
}

func netInfo() [][]string {
    ifaces, _ := net.Interfaces()
    var rows [][]string
    for _, iface := range ifaces {
        addrs, _ := iface.Addrs()
        for _, a := range addrs {
            rows = append(rows, []string{iface.Name, a.String()})
        }
    }
    return rows
}
