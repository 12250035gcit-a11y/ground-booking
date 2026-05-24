package model

import (
	"database/sql"
	"errors"
	"myapp/datastore/postgres"
	"time"
)

var ErrSlotBooked = errors.New("slot already booked")
var ErrNotFound = errors.New("not found")
var ErrInvalidTime = errors.New("invalid time range")

const insertBookingQuery = `INSERT INTO booking (student_id, match_type, date, starting_time, ending_time, notes, status) VALUES ($1,$2,$3,$4,$5,$6,'pending')`

// Only confirmed/approved bookings block a slot
const checkSlotQuery = `SELECT COUNT(*) FROM booking WHERE date = $1 AND starting_time < $3 AND ending_time > $2 AND status = 'approved'`
const checkSlotUpdateQuery = `SELECT COUNT(*) FROM booking WHERE date = $1 AND id != $2 AND starting_time < $4 AND ending_time > $3 AND status = 'approved'`

const getBookingByIDQuery = `SELECT id, student_id, match_type, date, starting_time, ending_time, notes, status FROM booking WHERE id=$1`
const updateBookingQuery = `UPDATE booking SET student_id=$1, match_type=$2, date=$3, starting_time=$4, ending_time=$5, notes=$6 WHERE id=$7`
const deleteBookingQuery = `DELETE FROM booking WHERE id=$1`

type Booking struct {
	ID            int    `json:"id"`
	StudentID     string `json:"student_id"`
	Match_Type    string `json:"match_type"`
	Date          string `json:"date"`
	Starting_time string `json:"starting_time"`
	Ending_time   string `json:"ending_time"`
	Notes         string `json:"notes"`
	Status        string `json:"status"`
}

func isValidTime(start, end string) bool {
	s, err1 := time.Parse("15:04", start)
	e, err2 := time.Parse("15:04", end)
	if err1 != nil || err2 != nil {
		return false
	}
	return s.Before(e)
}

func (b *Booking) CreateBooking() error {
	if !isValidTime(b.Starting_time, b.Ending_time) {
		return ErrInvalidTime
	}

	_, err := postgres.Db.Exec(insertBookingQuery,
		b.StudentID, b.Match_Type, b.Date, b.Starting_time, b.Ending_time, b.Notes)
	return err
}

func GetAllBookings() ([]Booking, error) {
	rows, err := postgres.Db.Query(`
		SELECT id, student_id, match_type, date, starting_time, ending_time, notes, status
		FROM booking ORDER BY date DESC, starting_time DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		err := rows.Scan(&b.ID, &b.StudentID, &b.Match_Type, &b.Date, &b.Starting_time, &b.Ending_time, &b.Notes, &b.Status)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}

func GetBookingByID(id int) (Booking, error) {
	var b Booking
	err := postgres.Db.QueryRow(getBookingByIDQuery, id).Scan(
		&b.ID, &b.StudentID, &b.Match_Type, &b.Date, &b.Starting_time, &b.Ending_time, &b.Notes, &b.Status,
	)
	if err == sql.ErrNoRows {
		return b, ErrNotFound
	}
	return b, err
}

func (b *Booking) UpdateBooking(id int) error {
	if !isValidTime(b.Starting_time, b.Ending_time) {
		return ErrInvalidTime
	}

	tx, err := postgres.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRow(checkSlotUpdateQuery, b.Date, id, b.Starting_time, b.Ending_time).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrSlotBooked
	}

	res, err := tx.Exec(updateBookingQuery, b.StudentID, b.Match_Type, b.Date, b.Starting_time, b.Ending_time, b.Notes, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return tx.Commit()
}

func DeleteBooking(id int) error {
	res, err := postgres.Db.Exec(deleteBookingQuery, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func UpdateBookingStatus(id int, status string) error {
	res, err := postgres.Db.Exec(`UPDATE booking SET status=$1 WHERE id=$2`, status, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// GetApprovedBookingsByDate returns all approved bookings for a given date (for the timeline)
func GetApprovedBookingsByDate(date string) ([]Booking, error) {
	rows, err := postgres.Db.Query(`
		SELECT id, student_id, match_type, date, starting_time, ending_time, notes, status
		FROM booking WHERE date=$1 AND status='approved'
		ORDER BY starting_time
	`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.StudentID, &b.Match_Type, &b.Date, &b.Starting_time, &b.Ending_time, &b.Notes, &b.Status); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}
