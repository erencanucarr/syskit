//go:build windows
// +build windows

package pulse

func readDisk() int {
    return 0
}

func readNet() (uint64, uint64) {
    return 0, 0
}
