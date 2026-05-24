package model

import (
	"crypto/rand"
	"errors"
	"fmt"
	"myapp/datastore/postgres"
)

type User struct {
	ID        int    `json:"id"`
	StudentID string `json:"student_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

const queryAdduser = `INSERT INTO users(student_id, first_name, last_name, phone, email, password)
VALUES($1, $2, $3, $4, $5, $6);`

func (u *User) Signup() error {
	u.StudentID = generateUUID()
	_, err := postgres.Db.Exec(queryAdduser, u.StudentID, u.FirstName, u.LastName, u.Phone, u.Email, u.Password)
	return err
}

const queryGetUser = `
SELECT id, student_id, first_name, last_name, phone, email, password
FROM users
WHERE email=$1;
`

func (u *User) Login() error {
	var dbUser User

	err := postgres.Db.QueryRow(queryGetUser, u.Email).Scan(
		&dbUser.ID,
		&dbUser.StudentID,
		&dbUser.FirstName,
		&dbUser.LastName,
		&dbUser.Phone,
		&dbUser.Email,
		&dbUser.Password,
	)
	if err != nil {
		return errors.New("user not found")
	}

	if dbUser.Password != u.Password {
		return errors.New("invalid password")
	}

	*u = dbUser
	return nil
}

func GetAllUsers() ([]User, error) {
	rows, err := postgres.Db.Query(`
		SELECT id, student_id, first_name, last_name, phone, email
		FROM users ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.StudentID, &u.FirstName, &u.LastName, &u.Phone, &u.Email); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func DeleteUser(id int) error {
	res, err := postgres.Db.Exec(`DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
