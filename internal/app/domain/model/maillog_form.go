package model

import "time"

type IndexMailLogRequest struct {
	DataTableRequest
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Keyword     string `json:"keyword"`
	SearchField int    `json:"searchField"`
}
type IndexMailLogResponse struct {
	ID        int64     `json:"id"`
	LogTime   time.Time `json:"logTime"`
	SessionID string    `json:"sessionID"`
	Process   string    `json:"process"`
	Message   string    `json:"message"`
}
