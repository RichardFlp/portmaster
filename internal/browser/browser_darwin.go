//go:build darwin

package browser

import "os/exec"

func openDefault(url string) error {
	return exec.Command("open", url).Start()
}
