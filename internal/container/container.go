package container

import (
    "bufio"
    "fmt"
    "os/exec"
    "strings"
)

type Item struct {
    ID, Name, Image, Status string
    Ports []string
}

// dockerCmd returns an *exec.Cmd using docker or podman binary.
// If neither engine exists in PATH, it returns a nil cmd and an error.
func dockerCmd(args ...string) (*exec.Cmd, error) {
    if path, err := exec.LookPath("docker"); err == nil {
        return exec.Command(path, args...), nil
    }
    if path, err := exec.LookPath("podman"); err == nil {
        return exec.Command(path, args...), nil
    }
    return nil, fmt.Errorf("docker or podman not found in PATH")
}

func List() ([]Item, error) {
    format := "{{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"
    cmd, err := dockerCmd("ps", "--format", format)
    if err != nil {
        return nil, err
    }
    out, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    items := []Item{}
    scanner := bufio.NewScanner(strings.NewReader(string(out)))
    for scanner.Scan() {
        parts := strings.Split(scanner.Text(), "\t")
        if len(parts) < 5 {
            continue
        }
        ports := strings.Split(parts[4], ",")
        if parts[4] == "" {
            ports = []string{}
        }
        items = append(items, Item{
            ID:     parts[0],
            Name:   parts[1],
            Image:  parts[2],
            Status: parts[3],
            Ports:  ports,
        })
    }
    return items, nil
}

func Prune() error {
    cmd, err := dockerCmd("container", "prune", "-f")
    if err != nil {
        return err
    }
    return cmd.Run()
}
