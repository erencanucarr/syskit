package ui

import (
    "os/exec"
    "strings"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type usersTick time.Time

type userRow struct {
    name string
    tty  string
}

type UsersModel struct {
    watch bool
    rows  []userRow
}

func NewUsersModel(watch bool) UsersModel {
    return UsersModel{watch: watch}
}

func (m UsersModel) Init() tea.Cmd {
    m.refresh()
    if m.watch {
        return tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return usersTick(t) })
    }
    return nil
}

func (m UsersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg.(type) {
    case usersTick:
        m.refresh()
        return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg { return usersTick(t) })
    case tea.KeyMsg:
        if msg.(tea.KeyMsg).String() == "q" {
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m UsersModel) View() string {
    header := lipgloss.NewStyle().Bold(true).Render("User  TTY")
    var lines []string
    lines = append(lines, header)
    for _, r := range m.rows {
        lines = append(lines, r.name+"  "+r.tty)
    }
    if len(m.rows) == 0 {
        lines = append(lines, "<no users>")
    }
    lines = append(lines, "", lipgloss.NewStyle().Faint(true).Render("q: quit"))
    return strings.Join(lines, "\n")
}

func (m *UsersModel) refresh() {
    out, err := exec.Command("who").Output()
    if err != nil {
        m.rows = nil
        return
    }
    lines := strings.Split(strings.TrimSpace(string(out)), "\n")
    var rows []userRow
    for _, l := range lines {
        f := strings.Fields(l)
        if len(f) >= 2 {
            rows = append(rows, userRow{name: f[0], tty: f[1]})
        }
    }
    m.rows = rows
}
