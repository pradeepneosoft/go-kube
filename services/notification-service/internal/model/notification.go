package model

import "time"

type Notification struct {
	ID        string
	OrderID   string
	UserID    string
	Message   string
	CreatedAt time.Time
}
