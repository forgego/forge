//go:build unix

package generator

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// processUmask determines the file creation mask without mutating the process umask.
// On Linux, it reads /proc/self/status where available. As a fallback (e.g. other Unix platforms),
// it probes the umask by creating a temporary file with mode 0o666 and reading its mode.
func processUmask() int {
	if mask, ok := readProcUmask(); ok {
		return mask
	}
	return probeUmask()
}

func readProcUmask() (int, bool) {
	f, err := os.Open("/proc/self/status")
	if err != nil {
		return 0, false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Umask:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				val, err := strconv.ParseInt(fields[1], 8, 0)
				if err == nil {
					return int(val), true
				}
			}
		}
	}
	return 0, false
}

func probeUmask() int {
	f, err := os.CreateTemp("", ".umask-probe-*")
	if err != nil {
		return 0o022
	}
	path := f.Name()
	_ = f.Close()
	_ = os.Remove(path)

	probe, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o666)
	if err != nil {
		return 0o022
	}
	defer func() {
		_ = probe.Close()
		_ = os.Remove(path)
	}()

	fi, err := probe.Stat()
	if err != nil {
		return 0o022
	}
	return int(0o666 &^ fi.Mode().Perm())
}
