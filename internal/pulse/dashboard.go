//go:build linux
// +build linux

package pulse

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "runtime"
    "strconv"
    "strings"
    "time"
    "syscall"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/progress"
    "github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

type proc struct {
    pid  string
    name string
    cpu  string
    mem  string
}

type Model struct {
    activeTab int // 0: cpu/mem, 1: disk, 2: net

    cpuLoad float64
    memPerc int
    procs   []proc
    cpuBar  progress.Model
    memBar  progress.Model
    diskBar progress.Model
    // disk
    diskPerc int
    // network (KB/s)
    netRx float64
    netTx float64
    prevRx uint64
    prevTx uint64
}

func Initial() Model {
    prg := progress.New(progress.WithDefaultGradient())
    return Model{cpuBar: prg, memBar: prg, diskBar: prg}
}

func (m Model) Init() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tickMsg:
        m.cpuLoad = readLoad()
        used, total := readMem()
        if total > 0 {
            m.memPerc = int(float64(used) / float64(total) * 100)
        }
        m.procs = topProcs()
        // disk
        m.diskPerc = readDisk()
        // net
        rx, tx := readNet()
        if m.prevRx != 0 {
            dt := 1.0 // seconds between ticks
            m.netRx = float64(rx-m.prevRx)/1024.0/dt
            m.netTx = float64(tx-m.prevTx)/1024.0/dt
        }
        m.prevRx, m.prevTx = rx, tx
        return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
    case tea.KeyMsg:
        switch msg.String() {
        case "q":
            return m, tea.Quit
        case "tab", "right":
            m.activeTab = (m.activeTab + 1) % 3
        case "left":
            m.activeTab = (m.activeTab + 3 - 1) % 3
        }
    }
    return m, nil
}

func (m Model) View() string {
    width := 60
    tabs := []string{"CPU/MEM", "Disk", "Net"}
    var tabLabels []string
    for i, t := range tabs {
        if i == m.activeTab {
            tabLabels = append(tabLabels, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(t))
        } else {
            tabLabels = append(tabLabels, t)
        }
    }
    header := lipgloss.JoinHorizontal(lipgloss.Top, tabLabels...)

    // content below
    cpuBar := m.cpuBar.ViewAs(m.cpuLoad / float64(runtime.NumCPU()))
    memBar := m.memBar.ViewAs(float64(m.memPerc) / 100)

    title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Syskit Pulse — q:quit")
    cpuLbl := fmt.Sprintf("CPU  %.2f  (cores %d)", m.cpuLoad, runtime.NumCPU())
    memLbl := fmt.Sprintf("MEM  %d%%", m.memPerc)

    var rows []string
    rows = append(rows, header)
    rows = append(rows, title)
    if m.activeTab == 0 {
        rows = append(rows, lipgloss.NewStyle().Width(width).Render(cpuLbl+"\n"+cpuBar))
        rows = append(rows, lipgloss.NewStyle().Width(width).Render(memLbl+"\n"+memBar))
    } else if m.activeTab == 1 {
        diskBar := m.diskBar.ViewAs(float64(m.diskPerc) / 100)
        diskLbl := fmt.Sprintf("DISK %d%%", m.diskPerc)
        rows = append(rows, lipgloss.NewStyle().Width(width).Render(diskLbl+"\n"+diskBar))
    } else {
        netLbl := fmt.Sprintf("NET  RX %.1f KB/s  TX %.1f KB/s", m.netRx, m.netTx)
        rows = append(rows, lipgloss.NewStyle().Width(width).Render(netLbl))
    }
    if m.activeTab == 0 {
        rows = append(rows, "\nPID   NAME                 CPU%  MEM%")
    for i, p := range m.procs {
        if i >= 10 {
            break
        }
        rows = append(rows, fmt.Sprintf("%4s %-20s %5s %5s", p.pid, p.name, p.cpu, p.mem))
    }
        }
    return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// helpers
func readLoad() float64 {
    b, _ := os.ReadFile("/proc/loadavg")
    f, _ := strconv.ParseFloat(strings.Fields(string(b))[0], 64)
    return f
}

func readMem() (used, total int64) {
    f, err := os.Open("/proc/meminfo")
    if err != nil {
        return 0, 0
    }
    defer f.Close()
    var avail int64
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        var key string
        var val int64
        fmt.Sscanf(scanner.Text(), "%s %d", &key, &val)
        switch key {
        case "MemTotal:":
            total = val * 1024
        case "MemAvailable:":
            avail = val * 1024
        }
    }
    used = total - avail
    return used, total
}

// readDisk returns percentage of root filesystem used
func readDisk() int {
    var statfs syscall.Statfs_t
    if err := syscall.Statfs("/", &statfs); err != nil {
        return 0
    }
    total := statfs.Blocks * uint64(statfs.Bsize)
    free := statfs.Bfree * uint64(statfs.Bsize)
    used := total - free
    if total == 0 {
        return 0
    }
    return int(float64(used) / float64(total) * 100)
}

// readNet returns cumulative rx, tx bytes across all interfaces
func readNet() (uint64, uint64) {
    f, err := os.Open("/proc/net/dev")
    if err != nil {
        return 0, 0
    }
    defer f.Close()
    var rx, tx uint64
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if strings.Contains(line, ":") {
            parts := strings.Split(line, ":")
            fields := strings.Fields(parts[1])
            if len(fields) >= 8 {
                r, _ := strconv.ParseUint(fields[0], 10, 64)
                t, _ := strconv.ParseUint(fields[8], 10, 64)
                rx += r
                tx += t
            }
        }
    }
    return rx, tx
}

func topProcs() []proc {
    out, _ := exec.Command("ps", "-eo", "pid,comm,%cpu,%mem", "--no-headers", "--sort=-%cpu").Output()
    scanner := bufio.NewScanner(strings.NewReader(string(out)))
    var list []proc
    for scanner.Scan() {
        f := strings.Fields(scanner.Text())
        if len(f) >= 4 {
            list = append(list, proc{pid: f[0], name: f[1], cpu: f[2], mem: f[3]})
        }
    }
    return list
}
