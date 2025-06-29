//go:build linux
// +build linux

package pulse

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"sort"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tickMsg time.Time

// proc represents a process row.
type proc struct {
	pid  string
	user string
	name string
	cpu  string
	mem  string
}

type sortField int

const (
	sortCPU sortField = iota
	sortMEM
	sortPID
)

type CustomAction struct {
	Name      string
	Process   string  // process name or "*" for any
	CPUThresh float64 // threshold, e.g. 90.0
	Action    string  // shell command, e.g. "kill -9 {pid}"
	Triggered bool    // to avoid repeat
}

type Model struct {
	// metrics
	cpuPerc  []int // per-core percent (len = NumCPU)
	memPerc  int
	swapPerc int
	diskPerc int
	netRx    float64 // KB/s
	netTx    float64

	// prev values for delta calculations
	prevIdle  []uint64
	prevTotal []uint64
	prevRx    uint64
	prevTx    uint64

	// process table
	procs      []proc
	tbl        table.Model
	sortBy     sortField
	filter     string
	filterMode bool
	selected   int

	// ui helpers
	progress progress.Model

	// kill process feedback
	killMsg string

	// custom actions
	customActions []CustomAction
	actionMsg     string

	// help menu
	helpMode bool
}

// Initial returns initialised model.
func Initial() Model {
	prg := progress.New(progress.WithDefaultGradient())
	columns := []table.Column{
		{Title: "PID", Width: 6},
		{Title: "USER", Width: 8},
		{Title: "CPU%", Width: 5},
		{Title: "MEM%", Width: 5},
		{Title: "CMD", Width: 20},
	}
	tbl := table.New(table.WithColumns(columns), table.WithFocused(true))

	// Örnek custom action: "python" adlı süreç %90 CPU'ya ulaşırsa öldür
	actions := []CustomAction{
		{
			Name:      "Kill High CPU Python",
			Process:   "python",
			CPUThresh: 90.0,
			Action:    "kill -9 {pid}",
			Triggered: false,
		},
	}

	return Model{
		progress:      prg,
		tbl:           tbl,
		sortBy:        sortCPU,
		customActions: actions,
	}
}

// Init starts ticker.
func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// Update handles events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.refreshMetrics()
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
	case tea.KeyMsg:
		if m.helpMode {
			if msg.String() == "h" || msg.String() == "?" || msg.String() == "esc" {
				m.helpMode = false
			}
			return m, nil
		}
		switch {
		case m.filterMode:
			switch msg.Type {
			case tea.KeyEsc, tea.KeyCtrlC:
				m.filterMode = false
				m.filter = ""
			case tea.KeyEnter:
				m.filterMode = false
			case tea.KeyBackspace:
				if len(m.filter) > 0 {
					m.filter = m.filter[:len(m.filter)-1]
				}
			default:
				if msg.String() != "" && len(msg.String()) == 1 {
					m.filter += msg.String()
				}
			}
			return m, nil
		default:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "f3":
				m.sortBy = sortCPU
			case "f4":
				m.sortBy = sortMEM
			case "f5":
				m.sortBy = sortPID
			case "/":
				m.filterMode = true
				m.filter = ""
			case "up", "k":
				if m.selected > 0 {
					m.selected--
				}
			case "down", "j":
				if m.selected < len(m.sortProcs())-1 {
					m.selected++
				}
			case "K":
				// Kill selected process
				procs := m.sortProcs()
				if m.selected < len(procs) {
					pid := procs[m.selected].pid
					err := exec.Command("kill", "-9", pid).Run()
					if err != nil {
						m.killMsg = fmt.Sprintf("Failed to kill process: %s", err)
					} else {
						m.killMsg = fmt.Sprintf("Process %s killed", pid)
					}
				}
			case "h", "?":
				m.helpMode = true
			}
		}
	}
	return m, nil
}

// View renders UI.
func (m Model) View() string {
	// Styles
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Background(lipgloss.Color("236")).Padding(0, 1)
	panelStyle := lipgloss.NewStyle().Width(40).Padding(0, 1)
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57")).Bold(true)
	killMsgStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("160")).Bold(true)

	// Header
	header := headerStyle.Render(" Syskit Pulse — q:quit  ↑↓:Navigate  K:Kill  /:Search  F3 CPU  F4 MEM  F5 PID ")

	// Panels
	var cpuPanels []string
	for i, p := range m.cpuPerc {
		bar := m.progress.ViewAs(float64(p) / 100)
		cpuPanels = append(cpuPanels, fmt.Sprintf("CPU%-2d %3d%% %s", i, p, bar))
	}
	cpuPanel := panelStyle.Foreground(lipgloss.Color("39")).Render(lipgloss.JoinVertical(lipgloss.Left, cpuPanels...))
	memPanel := panelStyle.Foreground(lipgloss.Color("45")).Render(fmt.Sprintf("MEM  %3d%% %s", m.memPerc, m.progress.ViewAs(float64(m.memPerc)/100)))
	diskPanel := panelStyle.Foreground(lipgloss.Color("99")).Render(fmt.Sprintf("DISK %3d%% %s", m.diskPerc, m.progress.ViewAs(float64(m.diskPerc)/100)))
	netPanel := panelStyle.Foreground(lipgloss.Color("81")).Render(fmt.Sprintf("NET  RX %.1f KB/s  TX %.1f KB/s", m.netRx, m.netTx))

	// All metrics in a single vertical box, equal width, no iç içe border
	metricsBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Margin(1, 0).
		Width(44).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				cpuPanel,
				memPanel,
				diskPanel,
				netPanel,
			),
		)

	// Filter bar
	filterBar := ""
	if m.filterMode {
		filterBar = panelStyle.Foreground(lipgloss.Color("220")).Render(fmt.Sprintf("Search: %s", m.filter))
	}

	// Table with selection and filter
	rows := []table.Row{}
	procs := m.sortProcs()
	if m.filter != "" {
		var filtered []proc
		for _, p := range procs {
			if strings.Contains(strings.ToLower(p.name), strings.ToLower(m.filter)) ||
				strings.Contains(strings.ToLower(p.user), strings.ToLower(m.filter)) ||
				strings.Contains(strings.ToLower(p.pid), strings.ToLower(m.filter)) {
				filtered = append(filtered, p)
			}
		}
		procs = filtered
	}
	for i, p := range procs {
		row := table.Row{p.pid, p.user, p.cpu, p.mem, p.name}
		if i == m.selected {
			for j := range row {
				row[j] = selectedStyle.Render(row[j])
			}
		}
		rows = append(rows, row)
	}
	m.tbl.SetRows(rows)
	if m.selected >= len(rows) && len(rows) > 0 {
		m.selected = len(rows) - 1
	}
	tableView := m.tbl.View()

	// Kill feedback
	killMsg := ""
	if m.killMsg != "" {
		killMsg = killMsgStyle.Render(m.killMsg)
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, metricsBox, filterBar, tableView, killMsg)
}

// --- helpers ---
func (m *Model) refreshMetrics() {
	// CPU
	m.cpuPerc, m.prevIdle, m.prevTotal = readCPUPerc(m.prevIdle, m.prevTotal)

	// MEM
	used, total := readMem()
	if total > 0 {
		m.memPerc = int(float64(used) / float64(total) * 100)
	}

	// DISK root
	m.diskPerc = readDisk()

	// NET
	rx, tx := readNet()
	if m.prevRx != 0 {
		m.netRx = float64(rx-m.prevRx) / 1024
		m.netTx = float64(tx-m.prevTx) / 1024
	}
	m.prevRx, m.prevTx = rx, tx

	// processes
	m.procs = topProcs()
}

func (m *Model) refreshTable() {
	rows := []table.Row{}
	for _, p := range m.sortProcs() {
		rows = append(rows, table.Row{p.pid, p.user, p.cpu, p.mem, p.name})
	}
	m.tbl.SetRows(rows)
	if len(rows) > 0 && m.tbl.Cursor() >= len(rows) {
		m.tbl.SetCursor(len(rows) - 1)
	}
}

func (m Model) sortProcs() []proc {
	out := make([]proc, len(m.procs))
	copy(out, m.procs)
	switch m.sortBy {
	case sortCPU:
		sort.Slice(out, func(i, j int) bool { return out[i].cpu > out[j].cpu })
	case sortMEM:
		sort.Slice(out, func(i, j int) bool { return out[i].mem > out[j].mem })
	case sortPID:
		sort.Slice(out, func(i, j int) bool { return out[i].pid < out[j].pid })
	}
	if len(out) > 30 {
		out = out[:30]
	}
	return out
}

// readCPUPerc returns percentage per core and new idle/total arrays.
func readCPUPerc(prevIdle, prevTotal []uint64) (percs []int, newIdle, newTotal []uint64) {
	f, _ := os.Open("/proc/stat")
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		if !strings.HasPrefix(txt, "cpu") || txt[:4] == "cpu " {
			continue
		}
		fields := strings.Fields(txt)
		if len(fields) < 5 {
			continue
		}
		var idle, total uint64
		var vals []uint64
		for _, s := range fields[1:] {
			v, _ := strconv.ParseUint(s, 10, 64)
			vals = append(vals, v)
		}
		for i, v := range vals {
			total += v
			if i == 3 { // idle
				idle = v
			}
		}
		newIdle = append(newIdle, idle)
		newTotal = append(newTotal, total)
	}
	percs = make([]int, len(newIdle))
	for i := range newIdle {
		if len(prevIdle) != len(newIdle) {
			percs[i] = 0
			continue
		}
		idleTicks := newIdle[i] - prevIdle[i]
		totalTicks := newTotal[i] - prevTotal[i]
		if totalTicks == 0 {
			percs[i] = 0
		} else {
			percs[i] = int(100 - (float64(idleTicks)/float64(totalTicks))*100)
		}
	}
	return percs, newIdle, newTotal
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
	return
}

func readDisk() int {
	var s syscall.Statfs_t
	if err := syscall.Statfs("/", &s); err != nil {
		return 0
	}
	total := s.Blocks * uint64(s.Bsize)
	free := s.Bfree * uint64(s.Bsize)
	if total == 0 {
		return 0
	}
	used := total - free
	return int(float64(used) / float64(total) * 100)
}

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
	out, _ := exec.Command("ps", "-eo", "pid,user,comm,%cpu,%mem", "--no-headers").Output()
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var list []proc
	for scanner.Scan() {
		f := strings.Fields(scanner.Text())
		if len(f) >= 5 {
			list = append(list, proc{pid: f[0], user: f[1], name: f[2], cpu: f[3], mem: f[4]})
		}
	}
	return list
}
