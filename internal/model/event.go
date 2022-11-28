package model

type Event struct {
	Uuid         string `json:"uuid" db:"uuid"`
	CalenderUuid string `json:"calender_uuid" db:"calender_uuid"`
	Title        string `json:"title" db:"title"`
	StartTime    int64  `json:"start_time" db:"start_time"`
	EndTime      int64  `json:"end_time" db:"end_time"`
}
