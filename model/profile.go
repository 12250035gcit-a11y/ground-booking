package model

import (
	"myapp/datastore/postgres"
)

type Profile struct {
	ID         int    `json:"id"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Phone      int64  `json:"phone"`
}

func (p *Profile) Add() error {
	_, err := postgres.Db.Exec(
		`INSERT INTO profile (email, department, phone) VALUES ($1, $2, $3)
		 ON CONFLICT (email) DO UPDATE SET department=$2, phone=$3`,
		p.Email, p.Department, p.Phone,
	)
	return err
}

func GetProfileByEmail(email string) (Profile, error) {
	var p Profile
	err := postgres.Db.QueryRow(
		`SELECT id, email, department, phone FROM profile WHERE email=$1`, email,
	).Scan(&p.ID, &p.Email, &p.Department, &p.Phone)
	return p, err
}

func (p *Profile) Update(email string) error {
	_, err := postgres.Db.Exec(
		`UPDATE profile SET department=$1, phone=$2 WHERE email=$3`,
		p.Department, p.Phone, email,
	)
	return err
}
