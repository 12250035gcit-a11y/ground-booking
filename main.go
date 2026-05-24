package main

import (
	"log"
	"myapp/datastore/postgres"
	"myapp/routs"
)

func main() {
	postgres.Connect()
	log.Println("Server starting on :8080")
	routs.Router()
}
