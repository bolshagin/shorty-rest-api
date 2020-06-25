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
	UserID        int    `json:"userid"`
	Email         string `json:"email"`
	Password      string `json:"password,omitempty"`
	EncryptedPair string `json:"-"`
	Links         []Link `json:"links,omitempty"`
}

func (u *User) CreateUser(db *sql.DB) (*User, error) {
	if err := u.Validate(); err != nil {
		return nil, err
	}

	if err := u.BeforeCreate(); err != nil {
		return nil, err
	}

	b, err := u.UserExistsByEmail(u.Email, db)
	if err != nil {
		return nil, err
	}

	if b {
		return nil, errors.New(fmt.Sprintf("user with email '%s' already exists", u.Email))
	}

	query := fmt.Sprintf(
		"INSERT INTO Users (email, password, encrypted_pair) VALUES ('%s', '%s', '%s')",
		u.Email,
		u.Password,
		u.EncryptedPair)

	_, err = db.Exec(query)
	if err != nil {
		return nil, err
	}

	if err := db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&u.UserID); err != nil {
		return nil, err
	}

	return u, nil
}

func (u *User) UserExistsByEmail(email string, db *sql.DB) (bool, error) {
	var count int
	if err := db.QueryRow("SELECT COUNT(1) FROM users WHERE email = ?",
		email).Scan(&count); err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func (u *User) Find(id int, db *sql.DB) (*User, error) {
	if err := db.QueryRow("SELECT userid, email, password, encrypted_pair FROM users WHERE userid = ?",
		id).Scan(&u.UserID, &u.Email, &u.Password, &u.EncryptedPair); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}

		return nil, err
	}
	return u, nil
}

func (u *User) FindAllLinks(id int, db *sql.DB) (*User, error) {
	rows, err := db.Query(
		"SELECT l.long_url, l.short_url FROM users u JOIN links l ON u.userid = l.userid WHERE u.userid = ?",
		id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.LongURL, &l.ShortURL); err != nil {
			return nil, err
		}
		u.Links = append(u.Links, l)
	}

	return u, nil
}

func (u *User) BeforeCreate() error {
	if len(u.Password) > 0 {
		enc, err := encryptString(u.Email, u.Password)
		if err != nil {
			return err
		}
		u.EncryptedPair = enc
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
