package model

import (
	"errors"
	"myapp/datastore/postgres"
)

type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Status    string `json:"status"`
}

const queryAdduser = `INSERT INTO users(first_name, last_name, phone, email, password, status)
VALUES($1, $2, $3, $4, $5, 'pending');`

func (u *User) Signup() error {
	_, err := postgres.Db.Exec(queryAdduser, u.FirstName, u.LastName, u.Phone, u.Email, u.Password)
	return err
}

const queryGetUser = `
SELECT id, first_name, last_name, phone, email, password, status
FROM users
WHERE email=$1;
`

func (u *User) Login() error {
	var dbUser User

	err := postgres.Db.QueryRow(queryGetUser, u.Email).Scan(
		&dbUser.ID,
		&dbUser.FirstName,
		&dbUser.LastName,
		&dbUser.Phone,
		&dbUser.Email,
		&dbUser.Password,
		&dbUser.Status,
	)

	if err != nil {
		return errors.New("user not found")
	}

	if dbUser.Password != u.Password {
		return errors.New("invalid password")
	}

	if dbUser.Status != "approved" {
		return errors.New("account pending approval")
	}

	*u = dbUser
	return nil
}

// GetAllUsers - admin only
func GetAllUsers() ([]User, error) {
	rows, err := postgres.Db.Query(`
		SELECT id, first_name, last_name, phone, email, status
		FROM users ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Phone, &u.Email, &u.Status)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// UpdateUserStatus - admin only
func UpdateUserStatus(id int, status string) error {
	res, err := postgres.Db.Exec(`UPDATE users SET status=$1 WHERE id=$2`, status, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
