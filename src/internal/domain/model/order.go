package dmodel

import "time"

type Order struct {
	Id        string    `json:"id"`
	UserId    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	Completed bool      `json:"completed"`
	Cost      int       `json:"cost"`
}
