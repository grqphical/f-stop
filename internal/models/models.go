package models

import "time"

type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Email        string `json:"email"`
}

type GPSCoordinates struct {
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

type PhotoMetadata struct {
	ID                string    `json:"id"`
	OwnerID           int       `json:"ownerID"`
	Filepath          string    `json:"-"`
	ThumbnailFilepath string    `json:"-"`
	ThumbnailJobId    *int      `json:"thumbnailJobId"`
	Uploaded          time.Time `json:"uploaded"`
	Size              int       `json:"size"`
	MimeType          string    `json:"mimeType"`
	Permalink         string    `json:"permalink"`

	EXIFCoordinates GPSCoordinates `json:"exifCoordinates"`
	EXIFCameraModel *string        `json:"exifCameraModel"`
	EXIFTakenAt     *time.Time     `json:"exifTakenAt"`
}

type JobPayload struct {
	Filepath string `json:"photoFilepath"`
	PhotoID  string `json:"photoID"`
}

type Job struct {
	ID         int        `json:"id"`
	Status     string     `json:"status"`
	Payload    JobPayload `json:"-"`
	VisibleAt  time.Time  `json:"-"`
	RetryCount int        `json:"-"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type Tag struct {
	ID      int    `json:"id"`
	OwnerID int    `json:"ownerID"`
	Name    string `json:"name"`
}

type TagAssignment struct {
	PhotoID string `json:"photoId"`
	TagID   string `json:"tagID"`
}
