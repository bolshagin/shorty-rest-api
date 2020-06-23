package model

import (
	"database/sql"
	"github.com/bolshagin/shorty-rest-api/tools"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
)

type Link struct {
	LinkID   int    `json:"linkid"`
	LongURL  string `json:"long_url"`
	ShortURL string `json:"short_url"`
}

func (l *Link) CreateLink(db *sql.DB) (*Link, error) {
	if err := l.Validate(); err != nil {
		return nil, err
	}

	_, err := db.Exec("INSERT INTO links (long_url) VALUES (?)", l.LongURL)
	if err != nil {
		return nil, err
	}

	if err := db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&l.LinkID); err != nil {
		return nil, err
	}

	l.ShortURL = tools.Encode(l.LinkID)
	db.QueryRow("UPDATE links SET short_url = ? WHERE linkid = ?", l.ShortURL, l.LinkID)

	return l, nil
}

func (l *Link) Validate() error {
	return validation.ValidateStruct(
		l,
		validation.Field(&l.LongURL, validation.Required, is.URL),
	)
}
