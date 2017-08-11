package handler

import (
	"github.com/go-pg/pg"
	"time"
)

type (
	handler struct {
		DB *pg.DB
	}
)

const (
	layout = "2006年 01月02日 15:04:05"
)

func NewHandler(db *pg.DB) *handler {
	return &handler{db}
}

func getTimeLimit(duration string) int {
	switch duration {
	case "day", "latest":
		return int(time.Now().Add(-1 * time.Hour).UnixNano())
	case "week":
		return int(time.Now().Add(-7 * 24 * time.Hour).UnixNano())
	case "month":
		return int(time.Now().Add(-30 * 24 * time.Hour).UnixNano())
	case "year":
		return int(time.Now().Add(-365 * 24 * time.Hour).UnixNano())
	default:
		return 0
	}
}
