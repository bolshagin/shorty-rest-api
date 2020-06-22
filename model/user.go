package model

import (
	"database/sql"
	b64 "encoding/base64"
	"errors"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type User struct {
	UserID            int    `json:"id"`
	Email             string `json:"email"`
	Password          string `json:"password,omitempty"`
	EncryptedPassword string `json:"-"`
}

func (u *User) CreateUser(db *sql.DB) (*User, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}

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

	if err := db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&u.UserID); err != nil {
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

func (u *User) Validate() error {
	return validation.ValidateStruct(
		u,
		validation.Field(&u.Email, validation.Required, is.Email),
		validation.Field(&u.Password, validation.Required, validation.Length(6, 100)))
}

func encryptString(email, password string) (string, error) {
	pair := email + ":" + password
	if len(pair) == 0 {
		return "", errors.New("cannot encrypt pair login/password")
	}
	return b64.StdEncoding.EncodeToString([]byte(pair)), nil
}

func (u *User) ClearPassword() {
	u.Password = ""
}
