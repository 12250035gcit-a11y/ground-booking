package controller

import (
	"encoding/json"
	"myapp/model"
	"myapp/utils/httpReps"
	"net/http"

	"github.com/gorilla/mux"
)

func AddDetails(w http.ResponseWriter, r *http.Request) {
	var p model.Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	defer r.Body.Close()

	if p.Email == "" {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	if err := p.Add(); err != nil {
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusCreated, map[string]string{"message": "Profile details saved"})
}

func UpdateDetails(w http.ResponseWriter, r *http.Request) {
	email := mux.Vars(r)["email"]
	var p model.Profile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	defer r.Body.Close()

	if err := p.Update(email); err != nil {
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]string{"message": "Profile updated"})
}

func GetSlots(w http.ResponseWriter, r *http.Request) {
	slots, err := model.GetAllSlots()
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if slots == nil {
		slots = []model.Slot{}
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, slots)
}
