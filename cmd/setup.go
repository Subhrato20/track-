package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Subhrato20/track-/internal/config"
)

func RunSetup() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("╭──────────────────────────────────────╮")
	fmt.Println("│  track- Setup                        │")
	fmt.Println("╰──────────────────────────────────────╯")
	fmt.Println()
	fmt.Println("You need a Ship24 API key to use track-.")
	fmt.Println("Sign up free at https://www.ship24.com and copy your API key.")
	fmt.Println()

	fmt.Print("API Key: ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		fmt.Println("\nError: API key is required.")
		os.Exit(1)
	}

	cfg := &config.Config{
		APIKey: apiKey,
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Config saved! Run 'track-' to start tracking packages.")
}
