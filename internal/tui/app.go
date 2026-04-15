package tui

import (
	"github.com/Subhrato20/track-/internal/db"
	"github.com/Subhrato20/track-/internal/usps"
	tea "github.com/charmbracelet/bubbletea"
)

type viewState int

const (
	viewList viewState = iota
	viewDetail
	viewAdd
	viewDelete
)

type AppModel struct {
	currentView   viewState
	list          listModel
	detail        detailModel
	add           addModel
	deleteConfirm deleteModel
	database      *db.DB
	uspsClient    *usps.Client
	width         int
	height        int
}

func NewApp(database *db.DB, uspsClient *usps.Client) AppModel {
	return AppModel{
		currentView: viewList,
		list:        newListModel(),
		detail:      newDetailModel(),
		add:         newAddModel(),
		deleteConfirm: newDeleteModel(),
		database:    database,
		uspsClient:  uspsClient,
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.loadPackages()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.width = msg.Width
		m.list.height = msg.Height
		m.detail.SetSize(msg.Width, msg.Height)
		m.add.width = msg.Width
		return m, nil

	case tea.KeyMsg:
		// Global quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		switch m.currentView {
		case viewList:
			return m.updateList(msg)
		case viewDetail:
			return m.updateDetail(msg)
		case viewAdd:
			return m.updateAdd(msg)
		case viewDelete:
			return m.updateDelete(msg)
		}
	}

	// Route non-key messages
	switch m.currentView {
	case viewList:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case viewDetail:
		var cmd tea.Cmd
		m.detail, cmd = m.detail.Update(msg)
		return m, cmd
	case viewAdd:
		var cmd tea.Cmd
		m.add, cmd = m.add.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m AppModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "a":
		m.currentView = viewAdd
		m.add.Reset()
		return m, m.add.trackingInput.Focus()

	case "d":
		pkg := m.list.SelectedPackage()
		if pkg != nil {
			m.deleteConfirm.SetPackage(*pkg)
			m.currentView = viewDelete
		}
		return m, nil

	case "enter":
		pkg := m.list.SelectedPackage()
		if pkg != nil {
			m.detail.SetPackage(*pkg)
			m.detail.SetSize(m.width, m.height)
			m.currentView = viewDetail
			return m, m.loadEvents(pkg.TrackingNumber)
		}
		return m, nil

	case "r":
		m.list.refreshing = true
		return m, tea.Batch(m.list.spinner.Tick, m.refreshAll())
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m AppModel) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.currentView = viewList
		return m, nil

	case "r":
		return m, m.refreshPackage(m.detail.pkg.TrackingNumber)
	}

	var cmd tea.Cmd
	m.detail, cmd = m.detail.Update(msg)
	return m, cmd
}

func (m AppModel) updateAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.currentView = viewList
		return m, nil

	case "enter":
		if err := m.add.Validate(); err != nil {
			m.add.err = err
			return m, nil
		}
		m.add.loading = true
		m.add.err = nil
		return m, tea.Batch(m.add.spinner.Tick, m.addPackage(m.add.TrackingNumber(), m.add.Nickname()))
	}

	var cmd tea.Cmd
	m.add, cmd = m.add.Update(msg)
	return m, cmd
}

func (m AppModel) updateDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		trackingNumber := m.deleteConfirm.pkg.TrackingNumber
		m.currentView = viewList
		return m, m.deletePackage(trackingNumber)

	case "n", "N", "esc":
		m.currentView = viewList
		return m, nil
	}

	return m, nil
}

func (m AppModel) View() string {
	switch m.currentView {
	case viewDetail:
		return m.detail.View()
	case viewAdd:
		return m.add.View()
	case viewDelete:
		return m.deleteConfirm.View()
	default:
		return m.list.View()
	}
}

// --- Commands ---

func (m AppModel) loadPackages() tea.Cmd {
	return func() tea.Msg {
		packages, err := m.database.ListPackages()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return PackagesLoadedMsg{Packages: packages}
	}
}

func (m AppModel) loadEvents(trackingNumber string) tea.Cmd {
	return func() tea.Msg {
		events, err := m.database.GetEvents(trackingNumber)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return EventsLoadedMsg{Events: events}
	}
}

func (m AppModel) addPackage(trackingNumber, nickname string) tea.Cmd {
	return func() tea.Msg {
		pkg, err := m.database.InsertPackage(trackingNumber, nickname)
		if err != nil {
			return TrackingErrorMsg{TrackingNumber: trackingNumber, Err: err}
		}

		// Try to fetch tracking info immediately
		if m.uspsClient != nil {
			resp, err := m.uspsClient.GetTracking(trackingNumber)
			if err == nil {
				m.database.UpdatePackageStatus(
					trackingNumber, resp.Status, resp.StatusCategory,
					resp.OriginCity, resp.OriginState,
					resp.DestCity, resp.DestState,
					resp.ExpectedDelivery,
				)

				dbEvents := make([]db.TrackingEvent, len(resp.TrackingEvents))
				for i, e := range resp.TrackingEvents {
					dbEvents[i] = db.TrackingEvent{
						TrackingNumber:   trackingNumber,
						EventDate:        e.EventDate,
						EventDescription: e.EventDescription,
						City:             e.City,
						State:            e.State,
						Zip:              e.Zip,
						Country:          e.Country,
					}
				}
				m.database.UpsertEvents(trackingNumber, dbEvents)

				// Re-read the updated package
				updated, err := m.database.GetPackage(trackingNumber)
				if err == nil {
					pkg = updated
				}
			}
		}

		return PackageAddedMsg{Package: *pkg}
	}
}

func (m AppModel) deletePackage(trackingNumber string) tea.Cmd {
	return func() tea.Msg {
		m.database.DeletePackage(trackingNumber)
		return PackageDeletedMsg{TrackingNumber: trackingNumber}
	}
}

func (m AppModel) refreshPackage(trackingNumber string) tea.Cmd {
	return func() tea.Msg {
		if m.uspsClient == nil {
			return TrackingErrorMsg{TrackingNumber: trackingNumber, Err: nil}
		}

		resp, err := m.uspsClient.GetTracking(trackingNumber)
		if err != nil {
			return TrackingErrorMsg{TrackingNumber: trackingNumber, Err: err}
		}

		m.database.UpdatePackageStatus(
			trackingNumber, resp.Status, resp.StatusCategory,
			resp.OriginCity, resp.OriginState,
			resp.DestCity, resp.DestState,
			resp.ExpectedDelivery,
		)

		dbEvents := make([]db.TrackingEvent, len(resp.TrackingEvents))
		for i, e := range resp.TrackingEvents {
			dbEvents[i] = db.TrackingEvent{
				TrackingNumber:   trackingNumber,
				EventDate:        e.EventDate,
				EventDescription: e.EventDescription,
				City:             e.City,
				State:            e.State,
				Zip:              e.Zip,
				Country:          e.Country,
			}
		}
		m.database.UpsertEvents(trackingNumber, dbEvents)

		pkg, _ := m.database.GetPackage(trackingNumber)
		events, _ := m.database.GetEvents(trackingNumber)

		return TrackingUpdatedMsg{
			TrackingNumber: trackingNumber,
			Package:        *pkg,
			Events:         events,
		}
	}
}

func (m AppModel) refreshAll() tea.Cmd {
	return func() tea.Msg {
		packages, _ := m.database.ListPackages()

		for _, pkg := range packages {
			if pkg.StatusCategory == "delivered" {
				continue
			}

			if m.uspsClient == nil {
				continue
			}

			resp, err := m.uspsClient.GetTracking(pkg.TrackingNumber)
			if err != nil {
				continue
			}

			m.database.UpdatePackageStatus(
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
			m.database.UpsertEvents(pkg.TrackingNumber, dbEvents)
		}

		return AllRefreshedMsg{}
	}
}
