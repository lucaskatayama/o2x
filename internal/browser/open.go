package browser

import (
	"os/exec"
	"runtime"
)

func Open(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Run()
	case "windows":
		return exec.Command("cmd", "/c", "start", url).Run()
	default:
		return exec.Command("xdg-open", url).Run()
	}
}
