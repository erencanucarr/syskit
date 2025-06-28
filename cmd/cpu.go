package cmd

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
	"syskit/internal/i18n"

	"syskit/internal/utils"

	"github.com/spf13/cobra"
)

var (
	watchCPU bool
)

var cpuCmd = &cobra.Command{
	Use:   "cpu",
	Short: "Display CPU information",
	Run: func(cmd *cobra.Command, args []string) {
		if watchCPU {
			for {
				printCPU()
				time.Sleep(2 * time.Second)
			}
		} else {
			printCPU()
		}
	},
}

func init() {
	cpuCmd.Flags().BoolVarP(&watchCPU, "watch", "w", false, "watch mode")
}

func printCPU() {
	cores := runtime.NumCPU()
	loadavg := getLoadAvg()
	headers := []string{i18n.T("cores"), i18n.T("loadavg")}
	rows := [][]string{{fmt.Sprintf("%d", cores), loadavg}}
	utils.Print(headers, rows)
}

func getLoadAvg() string {
	f, err := os.Open("/proc/loadavg")
	if err != nil {
		return "-"
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 3 {
			return strings.Join(parts[:3], " ")
		}
	}
	return "-"
}
