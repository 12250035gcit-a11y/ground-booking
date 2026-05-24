package model

import (
	"errors"
	"myapp/datastore/postgres"
)

type Admin struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *Admin) Login() error {
	var dbAdmin Admin

	err := postgres.Db.QueryRow(
		`SELECT id, email, password FROM admins WHERE email=$1`, a.Email,
	).Scan(&dbAdmin.ID, &dbAdmin.Email, &dbAdmin.Password)

	if err != nil {
		return errors.New("admin not found")
	}

	if dbAdmin.Password != a.Password {
		return errors.New("invalid password")
	}

	*a = dbAdmin
	return nil
}
