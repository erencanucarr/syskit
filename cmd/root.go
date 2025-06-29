package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"syskit/internal/config"
	"syskit/internal/i18n"
	"syskit/internal/utils"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	outputFormat string
	langCode     string
)

// rootCmd is the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "syskit",
	Short: "Modular Linux system management CLI",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg := config.Load()
		// determine language: flag > config > env
		if langCode == "" {
			if cfg.Lang != "" {
				langCode = cfg.Lang
			}
			langEnv := os.Getenv("LANG")
			if idx := strings.Index(langEnv, "."); idx != -1 {
				langEnv = langEnv[:idx]
			}
			if len(langEnv) >= 2 {
				langCode = langEnv[:2]
			} else {
				langCode = "en"
			}
		}
		i18n.Load(langCode)
		if langCode != cfg.Lang {
			cfg.Lang = langCode
			_ = config.Save()
		}
		syscleanCmd.Long = i18n.T("sysclean_help")
		utils.SetFormat(outputFormat)
	},
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// try plugin fallback
		if tryPlugin(os.Args[1:]) {
			return
		}
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// custom colorful help output

	// Bubble Tea tabanlı TUI yardım ekranı
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(cmd.CommandPath() + " - " + cmd.Short)
		long := ""
		if cmd.Long != "" {
			long = cmd.Long
		}
		fmt.Println(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Render(header))
		if long != "" {
			fmt.Println(lipgloss.NewStyle().Padding(0, 2).Render(long))
		}

		if len(cmd.Commands()) > 0 {
			fmt.Println(lipgloss.NewStyle().Bold(true).Render("Commands"))
			for _, c := range cmd.Commands() {
				if !c.IsAvailableCommand() || c.Hidden {
					continue
				}
				fmt.Printf("  %s  %s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Render(c.Name()), c.Short)
			}
		}

		if cmd.HasAvailableLocalFlags() {
			fmt.Println(lipgloss.NewStyle().Bold(true).Render("Flags"))
			fmt.Print(cmd.LocalFlags().FlagUsagesWrapped(90))
		}
		if cmd.HasAvailableInheritedFlags() {
			fmt.Println(lipgloss.NewStyle().Bold(true).Render("Global Flags"))
			fmt.Print(cmd.InheritedFlags().FlagUsagesWrapped(90))
		}
	})
	cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table", "output format: table|json|yaml")
	rootCmd.PersistentFlags().StringVar(&langCode, "lang", "", "language code (en, tr, de, es)")

	rootCmd.AddCommand(langCmd)
	rootCmd.AddCommand(cpuCmd)
	rootCmd.AddCommand(memCmd)
	rootCmd.AddCommand(portsCmd)
	rootCmd.AddCommand(usersCmd)
	rootCmd.AddCommand(syscleanCmd)
	rootCmd.AddCommand(pluginCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(scheduleCmd)
	rootCmd.AddCommand(processWatchCmd)
	rootCmd.AddCommand(timelineCmd)
	rootCmd.AddCommand(watchdogCmd)
	rootCmd.AddCommand(pulseCmd)
	rootCmd.AddCommand(servicesCmd)
	rootCmd.AddCommand(containersCmd)
}

// tryPlugin executes plugin binary if present under pluginDir.
func tryPlugin(args []string) bool {
	if len(args) == 0 {
		return false
	}
	bin := filepath.Join(PluginDir(), args[0])
	if stat, err := os.Stat(bin); err == nil && !stat.IsDir() {
		cmd := exec.Command(bin, args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Println("plugin error:", err)
		}
		return true
	}
	return false
}
