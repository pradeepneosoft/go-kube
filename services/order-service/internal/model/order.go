package model

import "time"

type Order struct {
	ID          string
	UserID      string
	ProductName string
	Quantity    int
	Status      string
	CreatedAt   time.Time
}
