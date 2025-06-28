package cmd

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "sort"
    "strconv"
    "strings"

    "github.com/spf13/cobra"
    "syskit/internal/i18n"
    "syskit/internal/utils"
)

var (
    dryRun bool
    force  bool
    auto   bool
)

var syscleanCmd = &cobra.Command{
    Use:   "sysclean",
    Short: "Clean temp files and caches",
    Run: func(cmd *cobra.Command, args []string) {
        runSysclean()
    },
}

type deleteItem struct {
    Path string
    Size int64
}

func runSysclean() {
    showDiskUsage()
    showLargest("/")

    targets := collectTargets()
    previewTargets(targets)

    if dryRun {
        return
    }

    if !force && !auto {
        if !confirm(i18n.T("Proceed with deletion? [y/N]:")) {
            fmt.Println("aborted")
            return
        }
    }

    deleteTargets(targets)
    fmt.Println("\nAfter cleanup:")
    showDiskUsage()
}

func showDiskUsage() {
    total, used := diskUsage("/")
    free := total - used
    headers := []string{"Total", "Used", "Free"}
    rows := [][]string{{human(total), human(used), human(free)}}
    utils.Print(headers, rows)
}

func diskUsage(path string) (total, used uint64) {
    out, err := exec.Command("df", "-k", path).Output()
    if err != nil {
        return
    }
    lines := strings.Split(strings.TrimSpace(string(out)), "\n")
    if len(lines) < 2 {
        return
    }
    fields := strings.Fields(lines[1])
    if len(fields) < 5 {
        return
    }
    t, _ := strconv.ParseUint(fields[1], 10, 64)
    u, _ := strconv.ParseUint(fields[2], 10, 64)
    total = t * 1024
    used = u * 1024
    return
}

func human(b uint64) string {
    const unit = 1024
    if b < unit {
        return fmt.Sprintf("%d B", b)
    }
    div, exp := uint64(unit), 0
    for n := b / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    val := float64(b) / float64(div)
    return fmt.Sprintf("%.1f %ciB", val, "KMGTPE"[exp])
}

func showLargest(root string) {
    cmd := exec.Command("du", "-x", "-m", "--max-depth=1", root)
    out, err := cmd.Output()
    if err != nil {
        return
    }
    type entry struct{ path string; size int }
    var list []entry
    scanner := bufio.NewScanner(strings.NewReader(string(out)))
    for scanner.Scan() {
        fld := strings.Fields(scanner.Text())
        if len(fld) != 2 {
            continue
        }
        sz, _ := strconv.Atoi(fld[0])
        list = append(list, entry{fld[1], sz})
    }
    sort.Slice(list, func(i, j int) bool { return list[i].size > list[j].size })
    fmt.Println("\nTop directories in / (MB):")
    headers := []string{"Directory", "Size(MB)"}
    var rows [][]string
    for i, e := range list {
        if i >= 10 {
            break
        }
        rows = append(rows, []string{e.path, fmt.Sprintf("%d", e.size)})
    }
    utils.Print(headers, rows)
}

func collectTargets() []deleteItem {
    patterns := []string{
        "/tmp/*",
        "/var/tmp/*",
        "/var/cache/apt/archives/*.deb",
        "/var/cache/yum/*",
        "/var/cache/dnf/*",
        "/var/log/*.log.*",
    }
    var items []deleteItem
    for _, pat := range patterns {
        matches, _ := filepath.Glob(pat)
        for _, m := range matches {
            var sz int64
            filepath.Walk(m, func(_ string, info os.FileInfo, err error) error {
                if err == nil {
                    sz += info.Size()
                }
                return nil
            })
            items = append(items, deleteItem{m, sz})
        }
    }
    sort.Slice(items, func(i, j int) bool { return items[i].Size > items[j].Size })
    return items
}

func previewTargets(it []deleteItem) {
    fmt.Println("\nItems to delete:")
    headers := []string{"Path", "Size"}
    var rows [][]string
    var total int64
    for _, d := range it {
        rows = append(rows, []string{d.Path, human(uint64(d.Size))})
        total += d.Size
    }
    utils.Print(headers, rows)
    fmt.Printf("TOTAL: %s\n", human(uint64(total)))
}

func confirm(msg string) bool {
    fmt.Print(msg)
    reader := bufio.NewReader(os.Stdin)
    in, _ := reader.ReadString('\n')
    in = strings.ToLower(strings.TrimSpace(in))
    return in == "y" || in == "yes"
}

func deleteTargets(it []deleteItem) {
    fmt.Println("\nDeleting:")
    for _, d := range it {
        fmt.Printf("%s (%s)\n", d.Path, human(uint64(d.Size)))
        os.RemoveAll(d.Path)
    }
}

func init() {
    syscleanCmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview items to be cleaned")
    syscleanCmd.Flags().BoolVar(&force, "force", false, "force deletion without confirmation")
    syscleanCmd.Flags().BoolVar(&auto, "auto", false, "non-interactive mode for cron")
}

