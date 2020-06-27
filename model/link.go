package model

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/bolshagin/shorty-rest-api/tools"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"net/http"
)

type Link struct {
	LinkID   int    `json:"linkid,omitempty"`
	LongURL  string `json:"long_url,omitempty"`
	ShortURL string `json:"short_url,omitempty"`
	UserID   int    `json:"userid,omitempty"`
	Clicks   int    `json:"n_clicks,omitempty"`
}

var (
	errLinkNotFound  = errors.New("link not found")
	errLinkNotExists = errors.New("link not exists")
)

func (l *Link) CreateLink(db *sql.DB, r *http.Request) (*Link, error) {
	if err := l.Validate(); err != nil {
		return nil, err
	}

	_, err := db.Exec("INSERT INTO links (long_url, userid) VALUES (?, ?)", l.LongURL, l.UserID)
	if err != nil {
		return nil, err
	}

	if err := db.QueryRow("SELECT LAST_INSERT_ID()").Scan(&l.LinkID); err != nil {
		return nil, err
	}

	l.ShortURL = "http://" + r.Host + "/" + tools.Encode(l.LinkID)
	db.QueryRow("UPDATE links SET short_url = ? WHERE linkid = ?", l.ShortURL, l.LinkID)

	return l, nil
}

func (l *Link) Find(id int, db *sql.DB) (*Link, error) {
	if err := db.QueryRow("SELECT linkid, long_url, short_url, userid FROM links WHERE linkid = ?",
		id).Scan(&l.LinkID, &l.LongURL, &l.ShortURL, &l.UserID); err != nil {
		if err == sql.ErrNoRows {
			return nil, errLinkNotFound
		}
		return nil, err
	}
	return l, nil
}

func (l *Link) FindLinkClicks(id int, short string, db *sql.DB) (*Link, error) {
	query := fmt.Sprintf(
		"SELECT l.linkid, MAX(l.long_url), MAX(short_url), MAX(l.userid), COUNT(c.linkid) "+
			"FROM links l LEFT JOIN clicks c ON l.linkid = c.linkid "+
			"WHERE l.userid = %v AND l.short_url = '%v'"+
			"GROUP BY l.linkid", id, short)

	if err := db.QueryRow(query).Scan(&l.LinkID, &l.LongURL, &l.ShortURL, &l.UserID, &l.Clicks); err != nil {
		if err == sql.ErrNoRows {
			return nil, errLinkNotFound
		}
		return nil, err
	}
	return l, nil
}

func FindLinksTop(db *sql.DB) (*[]Link, error) {
	var links []Link
	rows, err := db.Query(
		"SELECT l.long_url, COUNT(l.linkid) FROM links l JOIN clicks c ON l.linkid = c.linkid " +
			"GROUP BY l.long_url " +
			"ORDER BY COUNT(l.linkid) DESC " +
			"LIMIT 20")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.LongURL, &l.Clicks); err != nil {
			return nil, err
		}
		links = append(links, l)
	}

	return &links, nil
}

func (l *Link) DeleteByUserIDAndShort(id int, short string, db *sql.DB) error {
	b, err := l.ExistsByUserIDAndShort(id, short, db)
	if err != nil {
		return err
	}

	if !b {
		return errLinkNotExists
	}

	_, err = db.Query(
		"DELETE l, c FROM links l LEFT JOIN clicks c on l.linkid = c.linkid WHERE userid = ? AND short_url = ?",
		id,
		short,
	)
	return err
}

func (l *Link) ExistsByUserIDAndShort(id int, short string, db *sql.DB) (bool, error) {
	var count int
	if err := db.QueryRow(
		"SELECT COUNT(1) FROM links WHERE userid = ? AND short_url = ?",
		id,
		short).Scan(&count); err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func (l *Link) Validate() error {
	return validation.ValidateStruct(
		l,
		validation.Field(&l.LongURL, validation.Required, is.URL),
	)
}

func (l *Link) ClearUserID() {
	l.UserID = 0
}
