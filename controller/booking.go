package controller

import (
	"encoding/json"
	"myapp/model"
	"myapp/utils/httpReps"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func CreateBooking(w http.ResponseWriter, r *http.Request) {
	var b model.Booking
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	defer r.Body.Close()

	if err := b.CreateBooking(); err != nil {
		switch err {
		case model.ErrInvalidTime:
			httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid time range: end must be after start")
		case model.ErrDurationExceeded:
			httpReps.ResponseWithError(w, http.StatusBadRequest, "Booking duration cannot exceed 1.5 hours")
		case model.ErrSlotBooked:
			httpReps.ResponseWithError(w, http.StatusConflict, "This time slot is already booked or approved")
		default:
			httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusCreated, map[string]string{"message": "Booking submitted and awaiting admin approval"})
}

func GetAllBookings(w http.ResponseWriter, r *http.Request) {
	bookings, err := model.GetAllBookings()
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if bookings == nil {
		bookings = []model.Booking{}
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, bookings)
}

func GetBookingByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	b, err := model.GetBookingByID(id)
	if err == model.ErrNotFound {
		httpReps.ResponseWithError(w, http.StatusNotFound, "Booking not found")
		return
	}
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, b)
}

func UpdateBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	var b model.Booking
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	defer r.Body.Close()

	if err := b.UpdateBooking(id); err != nil {
		switch err {
		case model.ErrSlotBooked:
			httpReps.ResponseWithError(w, http.StatusConflict, "Slot already booked")
		case model.ErrNotFound:
			httpReps.ResponseWithError(w, http.StatusNotFound, "Booking not found")
		default:
			httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]string{"message": "Booking updated"})
}

func DeleteBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	if err := model.DeleteBooking(id); err != nil {
		if err == model.ErrNotFound {
			httpReps.ResponseWithError(w, http.StatusNotFound, "Booking not found")
			return
		}
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]string{"message": "Booking deleted"})
}

// ApproveBooking approves a booking and auto-rejects any conflicting pending bookings.
func ApproveBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	rejected, err := model.ApproveBookingAndRejectConflicts(id)
	if err != nil {
		switch err {
		case model.ErrNotFound:
			httpReps.ResponseWithError(w, http.StatusNotFound, "Booking not found or not in pending state")
		case model.ErrSlotBooked:
			httpReps.ResponseWithError(w, http.StatusConflict, "Another booking is already approved for this slot")
		default:
			httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]any{
		"message":       "Booking approved",
		"auto_rejected": rejected,
	})
}

func RejectBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	if err := model.UpdateBookingStatus(id, "rejected"); err != nil {
		if err == model.ErrNotFound {
			httpReps.ResponseWithError(w, http.StatusNotFound, "Booking not found")
			return
		}
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]string{"message": "Booking rejected"})
}

// RequestCancelBooking lets a user request cancellation of an approved booking.
func RequestCancelBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	if err := model.RequestCancelBooking(id); err != nil {
		if err == model.ErrNotFound {
			httpReps.ResponseWithError(w, http.StatusNotFound, "Booking not found or not in approved state")
			return
		}
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]string{"message": "Cancellation request submitted"})
}

// ApproveCancelBooking confirms a user's cancellation request, freeing the slot.
func ApproveCancelBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	if err := model.UpdateBookingStatus(id, "cancelled"); err != nil {
		if err == model.ErrNotFound {
			httpReps.ResponseWithError(w, http.StatusNotFound, "Booking not found")
			return
		}
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]string{"message": "Cancellation approved — slot is now available"})
}

// DenyCancelBooking rejects a cancellation request and restores the booking to approved.
func DenyCancelBooking(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	if err := model.UpdateBookingStatus(id, "approved"); err != nil {
		if err == model.ErrNotFound {
			httpReps.ResponseWithError(w, http.StatusNotFound, "Booking not found")
			return
		}
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]string{"message": "Cancellation request denied"})
}

func GetBookingsByDate(w http.ResponseWriter, r *http.Request) {
	date := mux.Vars(r)["date"]
	bookings, err := model.GetApprovedBookingsByDate(date)
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if bookings == nil {
		bookings = []model.Booking{}
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, bookings)
}

// GetActiveBookingsByDate returns approved + pending + cancel_requested for the HUD.
func GetActiveBookingsByDate(w http.ResponseWriter, r *http.Request) {
	date := mux.Vars(r)["date"]
	bookings, err := model.GetActiveBookingsByDate(date)
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if bookings == nil {
		bookings = []model.Booking{}
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, bookings)
}
