package model

type Calendar struct {
	Uuid      string   `json:"uuid" db:"uuid"`
	UserToken string   `json:"user_token" db:"user_token"`
	Date      int64    `json:"date" db:"date"`
	Events    []*Event `json:"events" db:"-"`
}
