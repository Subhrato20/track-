package cmd

import (
	"fmt"
	"os"
	"os/exec"
)

func RunSetup() {
	fmt.Println("╭──────────────────────────────────────╮")
	fmt.Println("│  track- Setup                        │")
	fmt.Println("╰──────────────────────────────────────╯")
	fmt.Println()
	fmt.Println("track- scrapes USPS directly — no API key needed!")
	fmt.Println()

	// Check for Chrome/Chromium
	browsers := []struct {
		name string
		path string
	}{
		{"Google Chrome", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"},
		{"Chromium", "/Applications/Chromium.app/Contents/MacOS/Chromium"},
		{"google-chrome", "google-chrome"},
		{"chromium", "chromium"},
	}

	found := false
	for _, b := range browsers {
		if _, err := os.Stat(b.path); err == nil {
			fmt.Printf("  Chrome found: %s\n", b.name)
			found = true
			break
		}
		if _, err := exec.LookPath(b.path); err == nil {
			fmt.Printf("  Chrome found: %s\n", b.name)
			found = true
			break
		}
	}

	if !found {
		fmt.Println("  WARNING: Chrome/Chromium not found.")
		fmt.Println("  Please install Google Chrome to use track-.")
		fmt.Println("  https://www.google.com/chrome/")
		fmt.Println()
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("  You're all set! Run 'track-' to start tracking packages.")
}
