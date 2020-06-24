package model

import (
	"database/sql"
	"time"
)

type Click struct {
	LinkID int       `json:"linkid"`
	Date   time.Time `json."time"`
}

func (c *Click) CreateClick(id int, db *sql.DB) (*Click, error) {
	loc, _ := time.LoadLocation("UTC")
	t := time.Now().In(loc)

	_, err := db.Query("INSERT INTO clicks (linkid, click_time) VALUES (?, ?)", id, t)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
