package wezterm

import (
	"encoding/json"
	"net/url"
	"strings"
)

type LivePane struct {
	WindowID int    `json:"window_id"`
	TabID    int    `json:"tab_id"`
	PaneID   int    `json:"pane_id"`
	Title    string `json:"title"`
	CWD      string `json:"cwd"`
}

// CWDPath extracts a filesystem path from the cwd field.
// WezTerm returns URIs like "file://wsl.localhost/Ubuntu/home/..." or "file:///C:/..."
func (p LivePane) CWDPath() string {
	u, err := url.Parse(p.CWD)
	if err != nil {
		return p.CWD
	}

	path := u.Path

	// WSL URI: file://wsl.localhost/Ubuntu/home/... → /home/...
	if strings.Contains(u.Host, "wsl") {
		// Strip the distro name prefix: /Ubuntu/home/... → /home/...
		parts := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 2)
		if len(parts) == 2 {
			path = "/" + parts[1]
		}
	}

	return strings.TrimSuffix(path, "/")
}

func (c *Client) ListPanes() ([]LivePane, error) {
	output, err := c.run("cli", "list", "--format", "json")
	if err != nil {
		return nil, err
	}

	var panes []LivePane
	if err := json.Unmarshal([]byte(output), &panes); err != nil {
		return nil, err
	}

	return panes, nil
}
