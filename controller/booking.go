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
		if err == model.ErrInvalidTime {
			httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid time range: end must be after start")
			return
		}
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
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
		if err == model.ErrSlotBooked {
			httpReps.ResponseWithError(w, http.StatusConflict, "Slot already booked")
			return
		}
		if err == model.ErrNotFound {
			httpReps.ResponseWithError(w, http.StatusNotFound, "Booking not found")
			return
		}
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
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

func ApproveBooking(w http.ResponseWriter, r *http.Request) {
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
	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]string{"message": "Booking approved"})
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
