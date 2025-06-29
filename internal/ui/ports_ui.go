package ui

import (
    "os/exec"
    "strings"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type portsTick time.Time

type portRow struct {
    proto string
    local string
    peer  string
}

type PortsModel struct {
    filter string
    watch  bool
    rows   []portRow
}

func NewPortsModel(filter string, watch bool) PortsModel {
    return PortsModel{filter: filter, watch: watch}
}

func (m PortsModel) Init() tea.Cmd {
    m.refresh()
    if m.watch {
        return tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return portsTick(t) })
    }
    return nil
}

func (m PortsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg.(type) {
    case portsTick:
        m.refresh()
        return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return portsTick(t) })
    case tea.KeyMsg:
        if msg.(tea.KeyMsg).String() == "q" {
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m PortsModel) View() string {
    header := lipgloss.NewStyle().Bold(true).Render("Proto  Local Address  Peer")
    var lines []string
    lines = append(lines, header)
    for _, r := range m.rows {
        lines = append(lines, r.proto+"  "+r.local+"  "+r.peer)
    }
    if len(m.rows)==0 {
        lines = append(lines, "<no matches>")
    }
    hint := lipgloss.NewStyle().Faint(true).Render("q: quit")
    lines = append(lines, "", hint)
    return strings.Join(lines, "\n")
}

func (m *PortsModel) refresh() {
    cmd := exec.Command("ss", "-tuln")
    out, err := cmd.Output()
    if err != nil {
        m.rows = nil
        return
    }
    lines := strings.Split(strings.TrimSpace(string(out)), "\n")
    var rows []portRow
    for _, l := range lines[1:] {
        if m.filter != "" && !strings.Contains(l, m.filter) {
            continue
        }
        f := strings.Fields(l)
        if len(f) < 5 {
            continue
        }
        rows = append(rows, portRow{proto: f[0], local: f[3], peer: f[4]})
    }
    m.rows = rows
}
