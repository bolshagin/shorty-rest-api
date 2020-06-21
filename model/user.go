package model

import (
	"database/sql"
	b64 "encoding/base64"
	"errors"
	"fmt"
)

type User struct {
	ID int `json:"id"`
	Email string `json:"email"`
	Password string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
}

func (u *User) CreateUser(db *sql.DB) (*User, error) {
	if err := u.BeforeCreate(); err != nil {
		return nil, err
	}

	query := fmt.Sprintf("INSERT INTO Users (email, password, encrypted_password) VALUES ('%s', '%s', '%s')",
		u.Email,
		u.Password,
		u.EncryptedPassword)

	_, err := db.Exec(query)
	if err != nil {
		return nil, err
	}

	if err := db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&u.ID); err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) BeforeCreate() error {
	if len(u.Password) > 0 {
		enc, err := encryptString(u.Email, u.Password)
		if err != nil {
			return err
		}
		u.EncryptedPassword = enc
	}
	return nil
}

func encryptString(email, password string) (string, error) {
	pair := email + ":" + password
	if len(pair) == 0 {
		return "", errors.New("cannot encrypt pair login/password")
	}
	return b64.StdEncoding.EncodeToString([]byte(pair)), nil
}

func (u *User) ClearPassword () {
	u.Password = ""
}