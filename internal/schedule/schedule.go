package schedule

import (
    "bytes"
    "fmt"
    "os/exec"
    "strings"
)

// Entry represents a syskit managed cron entry
type Entry struct {
    Name    string
    Spec    string
    Command string
}

// Read returns current crontab lines (user)
func Read() ([]string, error) {
    out, err := exec.Command("crontab", "-l").CombinedOutput()
    if err != nil {
        // if exit status 1 and message "no crontab" treat as empty
        if strings.Contains(string(out), "no crontab") {
            return []string{}, nil
        }
        // other error
    }
    lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
    if len(lines) == 1 && lines[0] == "" {
        return []string{}, nil
    }
    return lines, nil
}

// Write replaces the user's crontab with given lines
func Write(lines []string) error {
    inp := bytes.NewBufferString(strings.Join(lines, "\n") + "\n")
    cmd := exec.Command("crontab", "-")
    cmd.Stdin = inp
    return cmd.Run()
}

// ParseEntries extracts syskit-managed entries
func ParseEntries(lines []string) []Entry {
    entries := []Entry{}
    for i := 0; i < len(lines)-1; i++ {
        line := strings.TrimSpace(lines[i])
        if strings.HasPrefix(line, "# syskit:") {
            name := strings.TrimPrefix(line, "# syskit:")
            if i+1 < len(lines) {
                next := strings.TrimSpace(lines[i+1])
                fields := strings.Fields(next)
                if len(fields) >= 6 {
                    spec := strings.Join(fields[:5], " ")
                    cmd := strings.Join(fields[5:], " ")
                    entries = append(entries, Entry{Name: name, Spec: spec, Command: cmd})
                }
            }
        }
    }
    return entries
}

// RemoveEntry removes entry with given name
func RemoveEntry(lines []string, name string) ([]string, bool) {
    var out []string
    removed := false
    for i := 0; i < len(lines); i++ {
        if strings.TrimSpace(lines[i]) == "# syskit:"+name {
            // skip this and next line
            i++
            removed = true
            continue
        }
        out = append(out, lines[i])
    }
    return out, removed
}

// AddEntry appends new entry lines
func AddEntry(lines []string, e Entry) []string {
    lines = append(lines, fmt.Sprintf("# syskit:%s", e.Name))
    lines = append(lines, fmt.Sprintf("%s %s", e.Spec, e.Command))
    return lines
}
