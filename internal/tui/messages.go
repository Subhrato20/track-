package tui

import (
	"github.com/Subhrato20/track-/internal/db"
)

type PackagesLoadedMsg struct {
	Packages []db.Package
}

type PackageAddedMsg struct {
	Package db.Package
}

type PackageDeletedMsg struct {
	TrackingNumber string
}

type TrackingUpdatedMsg struct {
	TrackingNumber string
	Package        db.Package
	Events         []db.TrackingEvent
}

type TrackingErrorMsg struct {
	TrackingNumber string
	Err            error
}

type AllRefreshedMsg struct{}

type EventsLoadedMsg struct {
	Events []db.TrackingEvent
}

type ErrorMsg struct {
	Err error
}
