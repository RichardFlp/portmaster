//go:build linux

package browser

import "os/exec"

func openDefault(url string) error {
	return exec.Command("xdg-open", url).Start()
}
