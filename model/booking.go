package model

import (
	"database/sql"
	"errors"
	"myapp/datastore/postgres"
	"time"
)

var ErrSlotBooked       = errors.New("slot already booked")
var ErrNotFound         = errors.New("not found")
var ErrInvalidTime      = errors.New("invalid time range")
var ErrDurationExceeded = errors.New("booking duration cannot exceed 1.5 hours")

const insertBookingQuery = `INSERT INTO booking (student_id, match_type, date, starting_time, ending_time, notes, status) VALUES ($1,$2,$3,$4,$5,$6,'pending')`

// Only approved (and cancel_requested) bookings block new submissions
const checkSlotQuery = `SELECT COUNT(*) FROM booking WHERE date = $1 AND starting_time < $3 AND ending_time > $2 AND status IN ('approved','cancel_requested')`
const checkSlotUpdateQuery = `SELECT COUNT(*) FROM booking WHERE date = $1 AND id != $2 AND starting_time < $4 AND ending_time > $3 AND status IN ('approved','cancel_requested')`

const getBookingByIDQuery = `SELECT id, student_id, match_type, date, starting_time, ending_time, notes, status, to_char(created_at, 'YYYY-MM-DD HH24:MI:SS') FROM booking WHERE id=$1`
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
	CreatedAt     string `json:"created_at"`
}

func isValidTime(start, end string) bool {
	s, err1 := time.Parse("15:04", start)
	e, err2 := time.Parse("15:04", end)
	if err1 != nil || err2 != nil {
		return false
	}
	return s.Before(e)
}

func isValidDuration(start, end string) bool {
	s, err1 := time.Parse("15:04", start)
	e, err2 := time.Parse("15:04", end)
	if err1 != nil || err2 != nil {
		return false
	}
	return e.Sub(s) <= 90*time.Minute
}

func (b *Booking) CreateBooking() error {
	if !isValidTime(b.Starting_time, b.Ending_time) {
		return ErrInvalidTime
	}
	if !isValidDuration(b.Starting_time, b.Ending_time) {
		return ErrDurationExceeded
	}

	tx, err := postgres.Db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var count int
	if err = tx.QueryRow(checkSlotQuery, b.Date, b.Starting_time, b.Ending_time).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return ErrSlotBooked
	}

	_, err = tx.Exec(insertBookingQuery, b.StudentID, b.Match_Type, b.Date, b.Starting_time, b.Ending_time, b.Notes)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func GetAllBookings() ([]Booking, error) {
	rows, err := postgres.Db.Query(`
		SELECT id, student_id, match_type, date, starting_time, ending_time, notes, status, to_char(created_at, 'YYYY-MM-DD HH24:MI:SS')
		FROM booking ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.StudentID, &b.Match_Type, &b.Date, &b.Starting_time, &b.Ending_time, &b.Notes, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}

func GetBookingByID(id int) (Booking, error) {
	var b Booking
	err := postgres.Db.QueryRow(getBookingByIDQuery, id).Scan(
		&b.ID, &b.StudentID, &b.Match_Type, &b.Date, &b.Starting_time, &b.Ending_time, &b.Notes, &b.Status, &b.CreatedAt,
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
	if err = tx.QueryRow(checkSlotUpdateQuery, b.Date, id, b.Starting_time, b.Ending_time).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return ErrSlotBooked
	}

	res, err := tx.Exec(updateBookingQuery, b.StudentID, b.Match_Type, b.Date, b.Starting_time, b.Ending_time, b.Notes, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
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

// ApproveBookingAndRejectConflicts approves the booking and auto-rejects any other
// pending bookings that overlap on the same date. Returns count of auto-rejected bookings.
func ApproveBookingAndRejectConflicts(id int) (int, error) {
	tx, err := postgres.Db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var b Booking
	err = tx.QueryRow(`SELECT id, date, starting_time, ending_time FROM booking WHERE id=$1 AND status='pending'`, id).Scan(
		&b.ID, &b.Date, &b.Starting_time, &b.Ending_time,
	)
	if err == sql.ErrNoRows {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}

	// Ensure no already-approved booking conflicts
	var conflictCount int
	if err = tx.QueryRow(
		`SELECT COUNT(*) FROM booking WHERE date=$1 AND id!=$2 AND starting_time < $4 AND ending_time > $3 AND status IN ('approved','cancel_requested')`,
		b.Date, id, b.Starting_time, b.Ending_time,
	).Scan(&conflictCount); err != nil {
		return 0, err
	}
	if conflictCount > 0 {
		return 0, ErrSlotBooked
	}

	if _, err = tx.Exec(`UPDATE booking SET status='approved' WHERE id=$1`, id); err != nil {
		return 0, err
	}

	// Auto-reject all other pending bookings that overlap this slot
	res, err := tx.Exec(
		`UPDATE booking SET status='rejected' WHERE id!=$1 AND date=$2 AND starting_time < $4 AND ending_time > $3 AND status='pending'`,
		id, b.Date, b.Starting_time, b.Ending_time,
	)
	if err != nil {
		return 0, err
	}

	rejected, _ := res.RowsAffected()
	return int(rejected), tx.Commit()
}

// RequestCancelBooking marks an approved booking as cancellation-requested.
func RequestCancelBooking(id int) error {
	res, err := postgres.Db.Exec(`UPDATE booking SET status='cancel_requested' WHERE id=$1 AND status='approved'`, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// GetApprovedBookingsByDate returns only approved bookings for the timeline (legacy).
func GetApprovedBookingsByDate(date string) ([]Booking, error) {
	rows, err := postgres.Db.Query(`
		SELECT id, student_id, match_type, date, starting_time, ending_time, notes, status, to_char(created_at, 'YYYY-MM-DD HH24:MI:SS')
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
		if err := rows.Scan(&b.ID, &b.StudentID, &b.Match_Type, &b.Date, &b.Starting_time, &b.Ending_time, &b.Notes, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}

// GetActiveBookingsByDate returns approved, pending, and cancel_requested bookings
// for the availability HUD so all users can see what's booked or requested.
func GetActiveBookingsByDate(date string) ([]Booking, error) {
	rows, err := postgres.Db.Query(`
		SELECT id, student_id, match_type, date, starting_time, ending_time, notes, status, to_char(created_at, 'YYYY-MM-DD HH24:MI:SS')
		FROM booking WHERE date=$1 AND status IN ('approved','pending','cancel_requested')
		ORDER BY starting_time, created_at
	`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.StudentID, &b.Match_Type, &b.Date, &b.Starting_time, &b.Ending_time, &b.Notes, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, nil
}
