package models

type ListenRequest struct {
	UserID    string 	`json: "user_id" binding: "required"`
	TrackID   string    `json: "track_id" binding: "required"`
	Title     string    `json: "title"  binding: "required"`
}