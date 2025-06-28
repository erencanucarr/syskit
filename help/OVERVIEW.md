# Syskit – Developer Handbook

> Everything you need to understand the internals of the **Syskit** CLI at a glance.

---

## 1. Repository Layout

```
├── cmd/               # Top-level cobra commands (public surface)
│   ├── root.go        # Global flags, i18n bootstrap, command registry
│   ├── cpu.go         # `syskit cpu`
│   ├── mem.go         # `syskit mem`
│   ├── users.go       # `syskit users`
│   ├── ports.go       # `syskit ports`
│   ├── pulse.go       # launches TUI dashboard (wrapper)
│   └── ...
├── internal/
│   └── pulse/         # Bubble Tea dashboard implementation
│       ├── dashboard.go  # linux-only, full feature set
│       ├── pulse.go      # !linux stub so project builds everywhere
│       └── dummy_windows.go
├── utils/             # Reusable helpers (formatting, tables...)
├── lang/              # JSON translation files (en/tr/de/es/example)
├── config/            # Persisted TOML settings
├── help/              # Documentation (you are here)
└── README.md          # User-facing guide
```

---

## 2. Command Lifecycle

1. **Argument Parsing** – `cmd/root.go` configures Cobra and reads global flags.
2. **i18n** – `i18n.Load(langCode)` loads language JSONs into memory.
3. **Execution** – Each file in `cmd/` implements a `*cobra.Command` with its own `RunE`.
4. **Output** – Human table or JSON decided by `utils/format.go`.

> All user-visible strings **must** pass through `i18n.T("key")` so new languages pick them up automatically.

---

## 3. Pulse TUI Architecture (`internal/pulse`)

| Component | File | Details |
|-----------|------|---------|
| **Model** | `dashboard.go` | Holds state (activeTab, cpuLoad, memPerc, diskPerc, netRx/Tx, procs) |
| **Init**  | `Model.Init()` | Triggers a `tea.Tick` every second |
| **Update**| `Model.Update()`| Handles ticks & key events; refreshes metrics |
| **View**  | `Model.View()`  | Renders tabs using lipgloss + bubbles/progress |
| **Metrics**| Helpers | `readLoad`, `readMem`, `readDisk`, `readNet`, `topProcs` |

Build tags ensure:
* `linux`: full dashboard (needs `/proc` + `syscall.Statfs`).
* `!linux`: minimal CPU/MEM stub so `go vet ./...` stays green on Mac/Win.

### Keyboard Shortcuts
* `q` – quit  
* `Tab` / `←` / `→` – switch tabs

---

## 4. Translation Workflow

1. Add a key to **every** file in `lang/` (use `example.json` as template).
2. Reference the key in code: `i18n.T("my_new_key")`.
3. Run `go vet`; missing keys cause a runtime fallback to English.

---

## 5. Adding a New Command

```bash
cd cmd && cp example.go foo.go
```
1. Rename `Use`, `Short`, `RunE` contents.  
2. Inject business logic (parse /proc, run exec, etc.).  
3. Use `utils.RenderTable` or output JSON directly.  
4. Register in `root.go` (`rootCmd.AddCommand(fooCmd)`).

---

## 6. Testing & Linting

* **Static Analysis** – `go vet ./...` (CI fails if non-zero)
* **Unit Tests** – place under `*_test.go` (none critical yet – PRs welcome!)
* **Manual** – `syskit pulse` then play with tabs + `q`.

---

## 7. Build & Release Cheatsheet

```bash
# Build native binary
$ go build -o syskit ./cmd

# Cross-compile for Windows (stub dashboard)
$ GOOS=windows GOARCH=amd64 go build -tags "!linux" -o syskit.exe ./cmd

# Cross-compile for ARM (Raspberry Pi)
$ GOOS=linux GOARCH=arm GOARM=7 go build -tags linux -o syskit_arm ./cmd
```

---

## 8. FAQ

**Q:** Why are there duplicate structs in `pulse.go`?  
**A:** They were legacy stubs; now stripped. Only one Model/Proc in `dashboard.go` matters.

**Q:** Can I add GPU metrics?  
**A:** Yes – extend `dashboard.go`, add a fourth tab; remember to update translations.

---

Happy coding – pull requests are always welcome! 🎉

This document dives into the folder / package structure so contributors can quickly find their way around.

## cmd/* (Cobra Commands)
| File | Responsibility |
|------|----------------|
| `root.go` | Global flags (`--output`, `--lang`) and i18n bootstrapping. Registers every sub-command. |
| `cpu.go`  | Shows number of cores and 1/5/15-minute load averages. |
| `mem.go`  | RAM & swap usage, top memory processes. |
| `ports.go`| Lists listening TCP/UDP ports (netstat replacement). |
| `users.go`| Active login sessions (`who`). |
| `watchdog.go` | Background watchdog that kills runaway processes – optional. |
| `pulse.go`| Thin wrapper that just launches the Bubble Tea dashboard in `internal/pulse`. |

All commands use `utils/table.go` for aligned column output and `i18n` for user-facing strings.

## internal/pulse/
| File | Responsibility |
|------|----------------|
| `dashboard.go` (linux build tag) | Bubble Tea TUI with 3 tabs: CPU/MEM, Disk, Net. Uses progress bars (bubbles/progress) and lipgloss styling. Collects metrics from `/proc`. |
| `pulse.go` (!linux build tag) | Minimal stub so the repo still builds on non-Linux OSes. Provides CPU/MEM listing only. |
| `dummy_windows.go` | Empty implementations of `readDisk` / `readNet` for Windows so vet/build succeed there. |

### Metric Helpers
* `readLoad()` – one-minute load average.
* `readMem()` – MemTotal vs MemAvailable in `/proc/meminfo`.
* `readDisk()` – root filesystem usage via `syscall.Statfs`.
* `readNet()` – cumulative RX/TX from `/proc/net/dev`.
* `topProcs()` – top 10 processes by CPU.

## lang/
Contains translation JSON files (`en.json`, `tr.json`, `de.json`, `es.json`, `example.json`). New keys **must** exist in every file to keep the CLI fully localised.

`i18n.Load(langCode)` reads these at runtime.

## utils/
* `table.go` – simple ANSI column formatter (handles wide characters)
* `format.go` – json / table switch based on `--output` flag.

## config/
TOML-based persisted settings (currently just `Lang`).

## build tags
The dashboard is Linux-specific; other OSes fall back to the stub. This is why there are two build tags:

```
// +build linux   => dashboard.go
// +build !linux  => pulse.go
```

## Adding a new command
1. Create `cmd/yourcmd.go` with `cobra.Command`.
2. Add user-facing strings via `i18n.T("key")`; update all language JSONs.
3. Register the command inside `root.go`.

## Tests
Currently only unit-style integration tests around utils; more welcome!

