package ui

import (
    "os"
    "strconv"
    "strings"
    "time"
    "runtime"

    "github.com/charmbracelet/bubbles/progress"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type cpuTick time.Time

type CpuModel struct {
    load   float64
    watch  bool
    bar    progress.Model
}

func NewCpuModel(watch bool) CpuModel {
    prg := progress.New(progress.WithDefaultGradient())
    return CpuModel{watch: watch, bar: prg}
}

func (m CpuModel) Init() tea.Cmd {
    if m.watch {
        return tea.Tick(time.Second, func(t time.Time) tea.Msg { return cpuTick(t) })
    }
    return nil
}

func (m CpuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg.(type) {
    case cpuTick:
        m.load = readLoad()
        return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return cpuTick(t) })
    case tea.KeyMsg:
        if msg.(tea.KeyMsg).String() == "q" {
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m CpuModel) View() string {
    m.load = readLoad() // initial or each rendering when not watch
    cores := runtime.NumCPU()
    bar := m.bar.ViewAs(m.load / float64(cores))
    title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("CPU Usage — q:quit")
    info := lipgloss.NewStyle().Render(
        "Load: " + strconv.FormatFloat(m.load, 'f', 2, 64) + " | Cores: " + strconv.Itoa(cores))
    return lipgloss.JoinVertical(lipgloss.Left, title, info, bar)
}

func readLoad() float64 {
    b, _ := os.ReadFile("/proc/loadavg")
    f, _ := strconv.ParseFloat(strings.Fields(string(b))[0], 64)
    return f
}
