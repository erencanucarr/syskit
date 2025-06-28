package cmd

import (
	"bufio"
	"fmt"

	"syskit/internal/i18n"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"syskit/internal/utils"

	"github.com/spf13/cobra"
)

var sortBy string

var memCmd = &cobra.Command{
	Use:   "mem",
	Short: "Display memory usage",
	Run: func(cmd *cobra.Command, args []string) {
		printMem()
	},
}

func init() {
	memCmd.Flags().StringVarP(&sortBy, "sort", "s", "RSS", "sort by column when listing top processes (RSS|PSS)")
}

func printMem() {
	ram, swap := getMemInfo()
	procRows := topMemoryProcs()

	headers := []string{i18n.T("ram_percent"), i18n.T("swap_percent")}
	rows := [][]string{{ram, swap}}
	utils.Print(headers, rows)

	if len(procRows) > 0 {
		fmt.Println("\n" + i18n.T("Top processes"))
		utils.Print([]string{i18n.T("pid"), i18n.T("command"), i18n.T("rss")}, procRows)
	}
}

func getMemInfo() (string, string) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return "-", "-"
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
		if strings.HasPrefix(key, "MemTotal") {
			memTotal = val
		} else if strings.HasPrefix(key, "MemAvailable") {
			memFree = val
		} else if strings.HasPrefix(key, "SwapTotal") {
			swapTotal = val
		} else if strings.HasPrefix(key, "SwapFree") {
			swapFree = val
		}
	}
	ramPercent := 100 * (memTotal - memFree) / memTotal
	swapPercent := 0.0
	if swapTotal > 0 {
		swapPercent = 100 * (swapTotal - swapFree) / swapTotal
	}
	return fmt.Sprintf("%.1f", ramPercent), fmt.Sprintf("%.1f", swapPercent)
}

type procMem struct {
	pid  string
	cmd  string
	mem  int
}

func topMemoryProcs() [][]string {
	out, err := exec.Command("ps", "-eo", "pid,comm,rss", "--no-headers").Output()
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	procs := []procMem{}
	for _, l := range lines {
		parts := strings.Fields(l)
		if len(parts) < 3 {
			continue
		}
		rss, _ := strconv.Atoi(parts[2])
		procs = append(procs, procMem{pid: parts[0], cmd: parts[1], mem: rss})
	}
	// sort descending by mem
	sort.Slice(procs, func(i, j int) bool { return procs[i].mem > procs[j].mem })
	rows := [][]string{}
	for i, p := range procs {
		if i >= 5 {
			break
		}
		rows = append(rows, []string{p.pid, p.cmd, fmt.Sprintf("%d", p.mem)})
	}
	return rows
}
