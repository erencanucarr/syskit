package config

import (
    "io/ioutil"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v2"
)

// Config represents user configuration file ~/.syskit/config.yaml
// Thresholds are percentages (0-100)
// SMTP credentials used for alert emails
//
// Example YAML:
// smtp:
//   host: smtp.example.com
//   port: 587
//   username: user
//   password: pass
//   to: admin@example.com
// thresholds:
//   cpu: 90
//   ram: 90
//   disk: 90
//

type Config struct {
    Lang string `yaml:"lang"`
    SMTP struct {
        Host     string `yaml:"host"`
        Port     int    `yaml:"port"`
        Username string `yaml:"username"`
        Password string `yaml:"password"`
        To       string `yaml:"to"`
    } `yaml:"smtp"`
    Thresholds struct {
        CPU  int `yaml:"cpu"`
        RAM  int `yaml:"ram"`
        Disk int `yaml:"disk"`
    } `yaml:"thresholds"`
}

var cfg *Config

// Save writes current cfg to disk.
func Save() error {
    if cfg == nil {
        return nil
    }
    path := filePath()
    os.MkdirAll(filepath.Dir(path), 0o755)
    data, err := yaml.Marshal(cfg)
    if err != nil {
        return err
    }
    return ioutil.WriteFile(path, data, 0o644)
}

// Load reads config from file; if not present returns defaults.
func Load() *Config {
    if cfg != nil {
        return cfg
    }
    cfg = &Config{}
    cfg.Thresholds.CPU = 90
    cfg.Thresholds.RAM = 90
    cfg.Thresholds.Disk = 90

    path := filePath()
    data, err := ioutil.ReadFile(path)
    if err != nil {
        return cfg // defaults
    }
    yaml.Unmarshal(data, cfg)
    return cfg
}

func filePath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".syskit", "config.yaml")
}
