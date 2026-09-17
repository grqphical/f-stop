// TypeScript port of internal/models/models.go
// Go's time.Time is serialized as ISO 8601 string via encoding/json.
// Nullable Go pointers (*T) are represented as T | null.

export interface User {
  id: number
  username: string
  email: string
}

export interface GPSCoordinates {
  latitude: number | null
  longitude: number | null
}

export interface PhotoMetadata {
  id: string
  ownerID: number
  thumbnailPermalink: string
  thumbnailJobId: number | null
  uploaded: string // ISO 8601 date string (Go time.Time)
  size: number
  mimeType: string
  permalink: string
  exifCoordinates: GPSCoordinates
  exifCameraModel: string | null
  exifTakenAt: string | null // ISO 8601 date string
}

export interface JobPayload {
  photoFilepath: string
  photoID: string
}

export interface Job {
  id: number
  status: string
  createdAt: string // ISO 8601 date string
  updatedAt: string // ISO 8601 date string
}

export interface Tag {
  id: number
  ownerID: number
  name: string
}

export interface TagAssignment {
  photoId: string
  tagID: string
}

// Matches Go type TagAssigmentPayload (note original Go typo preserved)
export interface TagAssigmentPayload {
  tags: number[]
}

// Alias with correct spelling for convenience
export type TagAssignmentPayload = TagAssigmentPayload
