package db

import "time"

type Package struct {
	ID               int
	TrackingNumber   string
	Nickname         string
	Status           string
	StatusCategory   string
	OriginCity       string
	OriginState      string
	DestCity         string
	DestState        string
	ExpectedDelivery string
	LastUpdated      string
	CreatedAt        string
}

func (db *DB) InsertPackage(trackingNumber, nickname string) (*Package, error) {
	now := time.Now().Format(time.RFC3339)
	result, err := db.conn.Exec(
		`INSERT INTO packages (tracking_number, nickname, last_updated, created_at)
		 VALUES (?, ?, ?, ?)`,
		trackingNumber, nickname, now, now,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &Package{
		ID:             int(id),
		TrackingNumber: trackingNumber,
		Nickname:       nickname,
		Status:         "Unknown",
		StatusCategory: "unknown",
		LastUpdated:    now,
		CreatedAt:      now,
	}, nil
}

func (db *DB) ListPackages() ([]Package, error) {
	rows, err := db.conn.Query(
		`SELECT id, tracking_number, nickname, status, status_category,
		        origin_city, origin_state, dest_city, dest_state,
		        expected_delivery, last_updated, created_at
		 FROM packages ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var packages []Package
	for rows.Next() {
		var p Package
		err := rows.Scan(
			&p.ID, &p.TrackingNumber, &p.Nickname, &p.Status, &p.StatusCategory,
			&p.OriginCity, &p.OriginState, &p.DestCity, &p.DestState,
			&p.ExpectedDelivery, &p.LastUpdated, &p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		packages = append(packages, p)
	}
	return packages, rows.Err()
}

func (db *DB) GetPackage(trackingNumber string) (*Package, error) {
	var p Package
	err := db.conn.QueryRow(
		`SELECT id, tracking_number, nickname, status, status_category,
		        origin_city, origin_state, dest_city, dest_state,
		        expected_delivery, last_updated, created_at
		 FROM packages WHERE tracking_number = ?`,
		trackingNumber,
	).Scan(
		&p.ID, &p.TrackingNumber, &p.Nickname, &p.Status, &p.StatusCategory,
		&p.OriginCity, &p.OriginState, &p.DestCity, &p.DestState,
		&p.ExpectedDelivery, &p.LastUpdated, &p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (db *DB) UpdatePackageStatus(trackingNumber, status, statusCategory, originCity, originState, destCity, destState, expectedDelivery string) error {
	now := time.Now().Format(time.RFC3339)
	_, err := db.conn.Exec(
		`UPDATE packages SET status = ?, status_category = ?,
		        origin_city = ?, origin_state = ?, dest_city = ?, dest_state = ?,
		        expected_delivery = ?, last_updated = ?
		 WHERE tracking_number = ?`,
		status, statusCategory, originCity, originState, destCity, destState,
		expectedDelivery, now, trackingNumber,
	)
	return err
}

func (db *DB) DeletePackage(trackingNumber string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM tracking_events WHERE tracking_number = ?", trackingNumber); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM packages WHERE tracking_number = ?", trackingNumber); err != nil {
		return err
	}

	return tx.Commit()
}
