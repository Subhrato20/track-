package main

import (
	"fmt"
	"os"

	"github.com/Subhrato20/track-/cmd"
	"github.com/Subhrato20/track-/internal/config"
	"github.com/Subhrato20/track-/internal/db"
	"github.com/Subhrato20/track-/internal/tui"
	"github.com/Subhrato20/track-/internal/usps"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "update":
			cmd.RunUpdate()
			return
		case "version":
			fmt.Println("track- v0.1.0")
			return
		case "setup":
			cmd.RunSetup()
			return
		}
	}

	cfg := config.MustLoad()
	database := db.MustOpen()
	defer database.Close()

	client := usps.NewClient(cfg.ClientID, cfg.ClientSecret, cfg.BaseURL)

	m := tui.NewApp(database, client)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
