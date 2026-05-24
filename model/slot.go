package model

import "myapp/datastore/postgres"

type Slot struct {
	ID        int    `json:"id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Status    string `json:"status"`
}

func GetAllSlots() ([]Slot, error) {
	rows, err := postgres.Db.Query(`SELECT id, start_time, end_time, status FROM slots ORDER BY start_time`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []Slot
	for rows.Next() {
		var s Slot
		if err := rows.Scan(&s.ID, &s.StartTime, &s.EndTime, &s.Status); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, nil
}
