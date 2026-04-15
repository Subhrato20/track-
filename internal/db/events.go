package db

type TrackingEvent struct {
	ID               int
	TrackingNumber   string
	EventDate        string
	EventDescription string
	City             string
	State            string
	Zip              string
	Country          string
}

func (db *DB) UpsertEvents(trackingNumber string, events []TrackingEvent) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT OR IGNORE INTO tracking_events
		 (tracking_number, event_date, event_description, city, state, zip, country)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		if _, err := stmt.Exec(
			trackingNumber, e.EventDate, e.EventDescription,
			e.City, e.State, e.Zip, e.Country,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetEvents(trackingNumber string) ([]TrackingEvent, error) {
	rows, err := db.conn.Query(
		`SELECT id, tracking_number, event_date, event_description, city, state, zip, country
		 FROM tracking_events
		 WHERE tracking_number = ?
		 ORDER BY event_date DESC`,
		trackingNumber,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []TrackingEvent
	for rows.Next() {
		var e TrackingEvent
		err := rows.Scan(&e.ID, &e.TrackingNumber, &e.EventDate, &e.EventDescription,
			&e.City, &e.State, &e.Zip, &e.Country)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
