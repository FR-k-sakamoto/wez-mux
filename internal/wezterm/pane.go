package wezterm

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SplitPaneOptions struct {
	Direction string
	Percent   int
	CWD       string
	TopLevel  bool
}

// CurrentPaneID returns the current pane ID.
// 1. WEZTERM_PANE env var
// 2. Match pane whose cwd contains the current working directory
// 3. Error with hint to use --pane-id
func CurrentPaneID(client *Client) (int, error) {
	// Try env var first
	if value := strings.TrimSpace(os.Getenv("WEZTERM_PANE")); value != "" {
		id, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("parse WEZTERM_PANE %q: %w", value, err)
		}
		return id, nil
	}

	// Fall back: match by cwd
	panes, err := client.ListPanes()
	if err != nil {
		return 0, fmt.Errorf("could not list panes: %w", err)
	}
	if len(panes) == 0 {
		return 0, fmt.Errorf("no panes found")
	}

	cwd, _ := os.Getwd()
	if cwd != "" {
		for _, p := range panes {
			if p.CWDPath() == cwd {
				return p.PaneID, nil
			}
		}
	}

	// Could not auto-detect — show available panes for manual selection
	hint := "could not detect current pane. Use --pane-id to specify.\nAvailable panes:\n"
	for _, p := range panes {
		hint += fmt.Sprintf("  pane_id=%d  tab_id=%d  cwd=%s\n", p.PaneID, p.TabID, p.CWDPath())
	}
	return 0, fmt.Errorf(hint)
}

func (c *Client) SplitPane(parentPaneID int, opts SplitPaneOptions) (int, error) {
	args := []string{
		"cli",
		"split-pane",
		"--pane-id", strconv.Itoa(parentPaneID),
		"--percent", strconv.Itoa(opts.Percent),
	}

	switch opts.Direction {
	case "right":
		args = append(args, "--right")
	case "bottom":
		args = append(args, "--bottom")
	default:
		return 0, fmt.Errorf("unsupported split direction %q", opts.Direction)
	}

	if opts.TopLevel {
		args = append(args, "--top-level")
	}
	if strings.TrimSpace(opts.CWD) != "" {
		args = append(args, "--cwd", opts.CWD)
	}

	output, err := c.run(args...)
	if err != nil {
		return 0, err
	}

	id, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return 0, fmt.Errorf("parse split-pane output %q: %w", strings.TrimSpace(output), err)
	}

	return id, nil
}

// SpawnTab creates a new tab in the specified window and returns its pane ID.
func (c *Client) SpawnTab(windowID int, cwd string) (int, error) {
	args := []string{
		"cli",
		"spawn",
		"--window-id", strconv.Itoa(windowID),
	}
	if strings.TrimSpace(cwd) != "" {
		args = append(args, "--cwd", cwd)
	}

	output, err := c.run(args...)
	if err != nil {
		return 0, err
	}

	id, err := strconv.Atoi(strings.TrimSpace(output))
	if err != nil {
		return 0, fmt.Errorf("parse spawn output %q: %w", strings.TrimSpace(output), err)
	}

	return id, nil
}

// WindowIDForPane returns the window_id that contains the given pane.
func WindowIDForPane(panes []LivePane, paneID int) (int, error) {
	for _, p := range panes {
		if p.PaneID == paneID {
			return p.WindowID, nil
		}
	}
	return 0, fmt.Errorf("pane %d not found in pane list", paneID)
}

func (c *Client) KillPane(paneID int) error {
	_, err := c.run("cli", "kill-pane", "--pane-id", strconv.Itoa(paneID))
	return err
}
