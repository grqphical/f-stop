package models

import "time"

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Email        string `json:"email"`
}

type PhotoMetadata struct {
	ID        string    `json:"id"`
	OwnerID   int       `json:"ownerID"`
	Filepath  string    `json:"-"`
	Uploaded  time.Time `json:"uploaded"`
	Size      int       `json:"size"`
	MimeType  string    `json:"mimeType"`
	Permalink string    `json:"permalink"`
}
