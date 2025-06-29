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
	procs  []proc
	tbl    table.Model
	sortBy sortField

	// ui helpers
	progress progress.Model
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
	return Model{progress: prg, tbl: tbl, sortBy: sortCPU}
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
		m.refreshTable()
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "f3":
			m.sortBy = sortCPU
		case "f4":
			m.sortBy = sortMEM
		case "f5":
			m.sortBy = sortPID
		}
		var cmd tea.Cmd
		m.tbl, cmd = m.tbl.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}

// View renders UI.
func (m Model) View() string {
	styles := lipgloss.NewStyle().Bold(true)
	var lines []string
	// CPU bars per core
	for i, p := range m.cpuPerc {
		bar := m.progress.ViewAs(float64(p) / 100)
		lines = append(lines, fmt.Sprintf("CPU%-2d %3d%% %s", i, p, bar))
	}
	// Mem, Swap, Disk
	lines = append(lines, fmt.Sprintf("MEM  %3d%% %s", m.memPerc, m.progress.ViewAs(float64(m.memPerc)/100)))
	lines = append(lines, fmt.Sprintf("DISK %3d%% %s", m.diskPerc, m.progress.ViewAs(float64(m.diskPerc)/100)))
	lines = append(lines, fmt.Sprintf("NET  RX %.1f KB/s  TX %.1f KB/s", m.netRx, m.netTx))

	header := styles.Render("Syskit Pulse — q:quit  F3 CPU  F4 MEM  F5 PID")
	panel := lipgloss.JoinVertical(lipgloss.Left, lines...)

	tableView := m.tbl.View()

	return lipgloss.JoinVertical(lipgloss.Left, header, panel, "", tableView)
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
