// Command cloudlume is a k9s-style TUI that maps cloud resources found via
// local credentials into one live, organized dashboard.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/steamedeo/cloudlume/internal/config"
	"github.com/steamedeo/cloudlume/internal/provider"
	"github.com/steamedeo/cloudlume/internal/provider/aws"
	"github.com/steamedeo/cloudlume/internal/tui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cloudlume: failed to load config: %v\n", err)
		os.Exit(1)
	}

	provider.Register(aws.New(cfg))

	m := tui.New(provider.All(), cfg.RefreshInterval)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "cloudlume: %v\n", err)
		os.Exit(1)
	}
}
