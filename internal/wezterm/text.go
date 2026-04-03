package wezterm

import (
	"strconv"
	"strings"
)

// SendText sends text to a pane. Does NOT append Enter.
func (c *Client) SendText(paneID int, text string) error {
	_, err := c.run("cli", "send-text", "--pane-id", strconv.Itoa(paneID), "--no-paste", text)
	return err
}

// SendEnter sends a carriage return (Enter key) to a pane.
func (c *Client) SendEnter(paneID int) error {
	_, err := c.run("cli", "send-text", "--pane-id", strconv.Itoa(paneID), "--no-paste", "\r")
	return err
}

// GetText retrieves text from a pane.
// If startLine is non-zero, reads from that line (negative = scrollback).
func (c *Client) GetText(paneID int, startLine int) (string, error) {
	args := []string{"cli", "get-text", "--pane-id", strconv.Itoa(paneID)}
	if startLine != 0 {
		args = append(args, "--start-line", strconv.Itoa(startLine))
	}
	return c.run(args...)
}

func LastLines(text string, count int) string {
	if count < 1 {
		return ""
	}

	hasTrailingNewline := strings.HasSuffix(text, "\n")
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return ""
	}
	if len(lines) > count {
		lines = lines[len(lines)-count:]
	}

	result := strings.Join(lines, "\n")
	if hasTrailingNewline {
		result += "\n"
	}

	return result
}
