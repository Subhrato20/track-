package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Subhrato20/track-/internal/config"
	"github.com/Subhrato20/track-/internal/db"
	"github.com/Subhrato20/track-/internal/tracker"
)

func RunUpdate() {
	// Set up logging
	logPath := filepath.Join(config.ConfigDir(), "update.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	log.Printf("Starting update at %s", time.Now().Format(time.RFC3339))

	cfg, err := config.Load()
	if err != nil {
		log.Printf("Error loading config: %v", err)
		os.Exit(1)
	}

	database, err := db.Open()
	if err != nil {
		log.Printf("Error opening database: %v", err)
		os.Exit(1)
	}
	defer database.Close()

	client := tracker.NewClient(cfg.APIKey)

	packages, err := database.ListPackages()
	if err != nil {
		log.Printf("Error listing packages: %v", err)
		os.Exit(1)
	}

	updated := 0
	skipped := 0
	errors := 0

	for _, pkg := range packages {
		if pkg.StatusCategory == "delivered" {
			skipped++
			continue
		}

		resp, err := client.GetTracking(pkg.TrackingNumber)
		if err != nil {
			log.Printf("Error updating %s: %v", pkg.TrackingNumber, err)
			errors++
			continue
		}

		database.UpdatePackageStatus(
			pkg.TrackingNumber, resp.Status, resp.StatusCategory,
			resp.OriginCity, resp.OriginState,
			resp.DestCity, resp.DestState,
			resp.ExpectedDelivery,
		)

		dbEvents := make([]db.TrackingEvent, len(resp.TrackingEvents))
		for i, e := range resp.TrackingEvents {
			dbEvents[i] = db.TrackingEvent{
				TrackingNumber:   pkg.TrackingNumber,
				EventDate:        e.EventDate,
				EventDescription: e.EventDescription,
				City:             e.City,
				State:            e.State,
				Zip:              e.Zip,
				Country:          e.Country,
			}
		}
		database.UpsertEvents(pkg.TrackingNumber, dbEvents)

		updated++
		log.Printf("Updated %s: %s", pkg.TrackingNumber, resp.Status)
	}

	log.Printf("Update complete: %d updated, %d skipped (delivered), %d errors", updated, skipped, errors)
	fmt.Printf("Update complete: %d updated, %d skipped, %d errors\n", updated, skipped, errors)
}
