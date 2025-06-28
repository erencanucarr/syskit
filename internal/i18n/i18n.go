package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type dict map[string]string

var data dict

// Load loads language file from lang/<code>.json. Fallback to en.
func Load(code string) {
	path := filepath.Join("lang", code+".json")
	if !fileExists(path) {
		path = filepath.Join("lang", "en.json")
	}
	f, err := os.Open(path)
	if err != nil {
		data = dict{}
		return
	}
	defer f.Close()
	json.NewDecoder(f).Decode(&data)
}

// T returns translated string for key or key itself
func T(key string) string {
	if val, ok := data[key]; ok {
		return val
	}
	return key
}

func fileExists(p string) bool {
	if _, err := os.Stat(p); err == nil {
		return true
	}
	return false
}
