package ui

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/charmbracelet/bubbles/progress"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type memTick time.Time

type MemModel struct {
    ramPct  float64
    swapPct float64
    watch   bool
    ramBar  progress.Model
    swapBar progress.Model
}

func NewMemModel(watch bool) MemModel {
    prg := progress.New(progress.WithDefaultGradient())
    return MemModel{watch: watch, ramBar: prg, swapBar: prg}
}

func (m MemModel) Init() tea.Cmd {
    if m.watch {
        return tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return memTick(t) })
    }
    return nil
}

func (m MemModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg.(type) {
    case memTick:
        m.readMem()
        return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return memTick(t) })
    case tea.KeyMsg:
        if msg.(tea.KeyMsg).String() == "q" {
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m MemModel) View() string {
    m.readMem() // initial render
    title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Memory Usage — q:quit")
    ramLbl := fmt.Sprintf("RAM %.1f%%", m.ramPct)
    swapLbl := fmt.Sprintf("Swap %.1f%%", m.swapPct)
    return lipgloss.JoinVertical(lipgloss.Left,
        title,
        lipgloss.NewStyle().Render(ramLbl+"\n"+m.ramBar.ViewAs(m.ramPct/100)),
        lipgloss.NewStyle().Render(swapLbl+"\n"+m.swapBar.ViewAs(m.swapPct/100)),
    )
}

func (m *MemModel) readMem() {
    f, err := os.Open("/proc/meminfo")
    if err != nil {
        return
    }
    defer f.Close()
    scanner := bufio.NewScanner(f)
    var memTotal, memFree, swapTotal, swapFree float64
    for scanner.Scan() {
        parts := strings.Fields(scanner.Text())
        if len(parts) < 2 {
            continue
        }
        key := parts[0]
        val, _ := strconv.ParseFloat(parts[1], 64)
        switch {
        case strings.HasPrefix(key, "MemTotal"):
            memTotal = val
        case strings.HasPrefix(key, "MemAvailable"):
            memFree = val
        case strings.HasPrefix(key, "SwapTotal"):
            swapTotal = val
        case strings.HasPrefix(key, "SwapFree"):
            swapFree = val
        }
    }
    if memTotal > 0 {
        m.ramPct = 100 * (memTotal - memFree) / memTotal
    }
    if swapTotal > 0 {
        m.swapPct = 100 * (swapTotal - swapFree) / swapTotal
    }
}
