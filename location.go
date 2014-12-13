package main

import (
	"time"
)

type Location struct {
	Id        int64     `json:"-"`
	Nickname  string    `json:"nickname"`
	Latitude  float32   `json:"latitude"`
	Longitude float32   `json:"longitude"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
