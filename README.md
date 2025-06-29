![syskit](https://github.com/user-attachments/assets/d58250a4-ca28-4260-9032-2263a75d4ffb)

A fully-localized, modular system-management CLI for **Linux** written in Go. Syskit replaces a handful of classic Unix tools with one cohesive, themeable, multi-language interface.

---

## Table of Contents
1. [Key Features](#key-features)
2. [Quick Start](#quick-start)
3. [Internationalisation](#internationalisation)
4. [Command Reference](#command-reference)
5. [Pulse Dashboard](#pulse-dashboard)
6. [Build Tags & Cross-Compilation](#build-tags--cross-compilation)
7. [Configuration](#configuration)
8. [Contributing](#contributing)
9. [License](#license)

---

## Key Features

| Category | Highlights |
|----------|------------|
| **CLI Framework** | [Cobra](https://github.com/spf13/cobra) subtree with intuitive `syskit <noun>` command layout |
| **Live Dashboard** | Real-time CPU / Memory / Disk / Network stats via a Bubble Tea TUI with keyboard navigation |
| **i18n** | 4 ready-to-use languages (🇬🇧 EN, 🇩🇪 DE, 🇪🇸 ES, 🇹🇷 TR) – easy to extend with JSON |
| **Flexible Output** | Auto-formatted table or machine-readable JSON (`--output table|json`) |
| **Linux-native Metrics** | Reads from `/proc`, `ps`, `statfs` for accurate data with zero external deps |
| **Safe Fallbacks** | Stub implementations allow compiling on macOS / Windows even though Pulse extras are Linux-only |

---

## Quick Start

### Prerequisites
* Go ≥ 1.20
* Linux kernel ≥ 3.0 (for Disk / Net metrics)

```bash
# Clone & build
$ git clone https://github.com/erencanucarr/syskit.git && cd syskit
$ go install ./...

# Verify installation
$ syskit --help
```

---

## Internationalisation

Syskit auto-detects the language in this order:
1. `--lang` flag
2. `config.toml` (created on first run)
3. `$LANG` / `$LC_ALL` env vars (first two letters)
4. Fallback: `en`

To run in Turkish:
```bash
syskit --lang tr cpu
```
Adding a new locale is as easy as copying `lang/example.json` into your `xx.json`, translating the values, and rebuilding.

---

## Command Reference

| Command | What it does |
|---------|--------------|
| `syskit users`         | List currently logged-in users (who) |
| `syskit ports`         | Show listening TCP/UDP ports with processes |
| `syskit cpu`           | Core count and load averages |
| `syskit mem`           | RAM & swap stats, top memory hogs |
| `syskit pulse`         | Launch interactive TUI dashboard (q to quit) |
| `syskit watchdog`      | Optional daemon to kill runaway procs |
| `syskit sysclean`      | System clean-up helper (apt/yum caches, logs…) |
| `syskit timeline`      | Boot & shutdown event history |

Run `syskit <command> --help` for per-command flags.

---

## Pulse Dashboard

Keyboard shortcuts:
* `q` – quit
* `Tab` / `←` / `→` – switch between tabs

Tabs:
1. **CPU/MEM** – Bars for aggregate CPU load & memory usage plus live top-10 processes
2. **Disk** – Root filesystem utilisation bar
3. **Net** – RX/TX throughput in KB/s

> Disk & Net tabs are only compiled on Linux (`// +build linux`).

---

## Build Tags & Cross-Compilation

File | Tag | Purpose
---- | ---- | -------
`internal/pulse/dashboard.go` | `linux` | Full Pulse dashboard (syscall.Statfs, /proc/net)
`internal/pulse/pulse.go` | `!linux` | Minimal CPU/MEM stub so non-Linux builds succeed

Cross-compile for Raspberry Pi:
```bash
GOOS=linux GOARCH=arm GOARM=7 go build -tags linux -o syskit_arm ./cmd
```

---

## Configuration

A `~/.config/syskit/config.toml` file is generated automatically, storing persistent defaults such as `Lang` and preferred `OutputFormat`.

---

## Contributing

1. **Fork** then create a branch: `git checkout -b feat/my-feature`  
2. Ensure `go vet ./...` and (if applicable) `go test ./...` pass.  
3. If you introduce new user-facing strings, update **all** translation JSONs.  
4. Open a PR – screenshots/gifs are appreciated!

---

## License

Syskit is released under the MIT License – see `LICENSE` for details.




$ git clone https://github.com/erencanucarr/syskit.git && cd syskit
$ go install ./...
```

The resulting `syskit` binary is placed in your `$GOBIN`.

## Usage

```bash
syskit --help              # global flags
syskit cpu                 # CPU cores & load average
syskit mem -o json         # RAM / swap statistics as JSON
syskit ports               # Active listening ports
syskit pulse               # Interactive dashboard (press q to quit, tab/←/→ to change tab)
```

### Language

```
syskit --lang tr users     # Turkish
LANG=de syskit mem         # German (env fallback)
```

### Build Tags

* `linux` – enables advanced Pulse dashboard (disk/net metrics)
* `!linux` – stub implementation so project still builds on other OSes

The Makefile shows common recipes for cross-compilation.

## Contributing

1. Fork & branch (`feat/xyz`)
2. `go test ./...` & `go vet ./...`
3. Open PR – please add translations for **all** languages if you introduce new strings.

## License

