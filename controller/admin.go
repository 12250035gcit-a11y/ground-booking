package controller

import (
	"encoding/json"
	"myapp/model"
	"myapp/utils/httpReps"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	var admin model.Admin

	err := json.NewDecoder(r.Body).Decode(&admin)
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "Invalid JSON Body")
		return
	}
	defer r.Body.Close()

	err = admin.Login()
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]interface{}{
		"message": "admin login success",
		"admin":   admin,
	})
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := model.GetAllUsers()
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	httpReps.ResponseWithsJSON(w, http.StatusOK, users)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		httpReps.ResponseWithError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := model.DeleteUser(id); err != nil {
		if err == model.ErrNotFound {
			httpReps.ResponseWithError(w, http.StatusNotFound, "user not found")
			return
		}
		httpReps.ResponseWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpReps.ResponseWithsJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}
