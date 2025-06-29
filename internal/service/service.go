package service

import (
    "bufio"
    "os/exec"
    "strings"
)

type Unit struct {
    Name, Load, Active, Sub, Description string
}

// List returns systemd service units (requires Linux with systemd)
func List() ([]Unit, error) {
    // --no-pager avoids less; --all to list even inactive
    cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-legend", "--no-pager")
    out, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    res := []Unit{}
    scanner := bufio.NewScanner(strings.NewReader(string(out)))
    for scanner.Scan() {
        // columns: UNIT LOAD ACTIVE SUB DESCRIPTION
        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            continue
        }
        fields := strings.Fields(line)
        if len(fields) < 5 {
            continue
        }
        res = append(res, Unit{
            Name:        fields[0],
            Load:        fields[1],
            Active:      fields[2],
            Sub:         fields[3],
            Description: strings.Join(fields[4:], " "),
        })
    }
    return res, nil
}

// Control executes systemctl action on a service (start/stop/restart)
func Control(name, action string) error {
    return exec.Command("systemctl", action, name).Run()
}
