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
	fmt.Println("You need USPS API credentials to use track-.")
	fmt.Println("Register at https://developers.usps.com to get them.")
	fmt.Println()

	fmt.Print("Consumer Key (client_id): ")
	clientID, _ := reader.ReadString('\n')
	clientID = strings.TrimSpace(clientID)

	fmt.Print("Consumer Secret (client_secret): ")
	clientSecret, _ := reader.ReadString('\n')
	clientSecret = strings.TrimSpace(clientSecret)

	if clientID == "" || clientSecret == "" {
		fmt.Println("\nError: both fields are required.")
		os.Exit(1)
	}

	cfg := &config.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("Config saved! Run 'track-' to start tracking packages.")
}
