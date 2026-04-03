package wezterm

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Client struct {
	Bin string
}

func NewClient(bin string) *Client {
	if strings.TrimSpace(bin) == "" {
		bin = detectBinary()
	}

	return &Client{Bin: bin}
}

// detectBinary finds the wezterm CLI binary.
// On WSL, "wezterm" may not exist but "wezterm.exe" does.
func detectBinary() string {
	if _, err := exec.LookPath("wezterm"); err == nil {
		return "wezterm"
	}
	if _, err := exec.LookPath("wezterm.exe"); err == nil {
		return "wezterm.exe"
	}
	return "wezterm"
}

func (c *Client) run(args ...string) (string, error) {
	cmd := exec.Command(c.Bin, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s %v: %s", c.Bin, args, msg)
	}

	return stdout.String(), nil
}
