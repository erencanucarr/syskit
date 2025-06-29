//go:build !linux
// +build !linux

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

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
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
	cpuLoad float64
	memPerc int
	procs   []proc
	cpuBar  progress.Model
	memBar  progress.Model
}

func Initial() Model {
	prg := progress.New(progress.WithDefaultGradient())
	return Model{cpuBar: prg, memBar: prg}
}

func (m Model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case tickMsg:
		m.cpuLoad = readLoad()
		used, total := readMem()
		if total > 0 {
			m.memPerc = int(float64(used) / float64(total) * 100)
		}
		m.procs = topProcs()
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
	case tea.KeyMsg:
		k := msg.(tea.KeyMsg)
		if k.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	width := 60
	cpuBar := m.cpuBar.ViewAs(m.cpuLoad / float64(runtime.NumCPU()))
	memBar := m.memBar.ViewAs(float64(m.memPerc) / 100)

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Syskit Pulse — q:quit")
	cpuLbl := fmt.Sprintf("CPU  %.2f  (cores %d)", m.cpuLoad, runtime.NumCPU())
	memLbl := fmt.Sprintf("MEM  %d%%", m.memPerc)

	var rows []string
	rows = append(rows, title)
	rows = append(rows, lipgloss.NewStyle().Width(width).Render(cpuLbl+"\n"+cpuBar))
	rows = append(rows, lipgloss.NewStyle().Width(width).Render(memLbl+"\n"+memBar))
	rows = append(rows, "\nPID   NAME                 CPU%  MEM%")
	for i, p := range m.procs {
		if i >= 10 {
			break
		}
		rows = append(rows, fmt.Sprintf("%4s %-20s %5s %5s", p.pid, p.name, p.cpu, p.mem))
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
