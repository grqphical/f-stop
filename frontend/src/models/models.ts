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
  uploaded: Date // parsed from ISO 8601 date string (Go time.Time)
  size: number
  mimeType: string
  permalink: string
  exifCoordinates: GPSCoordinates
  exifCameraModel: string | null
  exifTakenAt: Date | null // parsed from ISO 8601 date string
}

export interface JobPayload {
  photoFilepath: string
  photoID: string
}

export interface Job {
  id: number
  status: string
  createdAt: Date // parsed from ISO 8601 date string
  updatedAt: Date // parsed from ISO 8601 date string
}

// Raw wire formats as returned by the API (Go time.Time serialized via encoding/json).
// Use these for response.json() casts, then convert to the parsed interfaces above.
export interface PhotoMetadataDTO extends Omit<PhotoMetadata, 'uploaded' | 'exifTakenAt'> {
  uploaded: string
  exifTakenAt: string | null
}

export interface JobDTO extends Omit<Job, 'createdAt' | 'updatedAt'> {
  createdAt: string
  updatedAt: string
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
