package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Subhrato20/track-/internal/tracker"
)

// RunDebug scrapes a tracking number and prints every field to stdout.
// Usage: track- debug <tracking_number>
func RunDebug(trackingNumber string) {
	fmt.Printf("🔍 Scraping USPS for: %s\n", trackingNumber)
	fmt.Printf("⏱  Started at: %s\n\n", time.Now().Format("15:04:05"))

	client := tracker.NewClient()
	defer client.Close()

	// First dump the raw page HTML so we can see what USPS is actually serving
	fmt.Println("📄 Fetching raw page HTML...")
	html, pageText, err := client.DumpPage(trackingNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ DumpPage error: %v\n", err)
	} else {
		htmlFile := "/tmp/usps_debug.html"
		os.WriteFile(htmlFile, []byte(html), 0644)
		fmt.Printf("   HTML saved to: %s\n", htmlFile)
		fmt.Printf("   Page text (first 500 chars):\n---\n%s\n---\n\n", truncate(pageText, 500))
	}

	// Now run the full tracking extraction
	start := time.Now()
	resp, err := client.GetTracking(trackingNumber)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error after %v: %v\n", elapsed, err)
		os.Exit(1)
	}

	fmt.Printf("✅ Done in %v\n\n", elapsed.Round(time.Millisecond))

	b, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(b))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
