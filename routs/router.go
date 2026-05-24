package routs

import (
	"log"
	"myapp/controller"
	"net/http"

	"github.com/gorilla/mux"
)

func Router() {
	router := mux.NewRouter()

	// User auth
	router.HandleFunc("/user/add", controller.Adduser).Methods("POST")
	router.HandleFunc("/user/login", controller.LoginUser).Methods("POST")

	// Admin auth
	router.HandleFunc("/admin/login", controller.AdminLogin).Methods("POST")

	// Admin user management
	router.HandleFunc("/admin/users", controller.GetAllUsers).Methods("GET")
	router.HandleFunc("/admin/users/{id}", controller.DeleteUser).Methods("DELETE")

	// Profile details
	router.HandleFunc("/add/details", controller.AddDetails).Methods("POST")
	router.HandleFunc("/update/{email}", controller.UpdateDetails).Methods("PUT")

	// Slots
	router.HandleFunc("/api/slots", controller.GetSlots).Methods("GET")

	// Bookings
	router.HandleFunc("/booking", controller.CreateBooking).Methods("POST")
	router.HandleFunc("/bookings", controller.GetAllBookings).Methods("GET")
	router.HandleFunc("/bookings/date/{date}", controller.GetBookingsByDate).Methods("GET")
	router.HandleFunc("/booking/{id}", controller.GetBookingByID).Methods("GET")
	router.HandleFunc("/booking/{id}", controller.UpdateBooking).Methods("PUT")
	router.HandleFunc("/booking/{id}", controller.DeleteBooking).Methods("DELETE")
	router.HandleFunc("/booking/{id}/approve", controller.ApproveBooking).Methods("PUT")
	router.HandleFunc("/booking/{id}/reject", controller.RejectBooking).Methods("PUT")

	// Static files
	fs := http.FileServer(http.Dir("./views"))
	router.PathPrefix("/").Handler(fs)

	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
