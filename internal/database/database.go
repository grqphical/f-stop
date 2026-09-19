package database

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/grqphical/f-stop/internal/auth"
	"github.com/grqphical/f-stop/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

func pgxErrorToDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgErr.Code == "23505" {
			return ErrUniqueConstraint
		}
		return err
	}
	return err
}

type Database struct {
	apiPool    *pgxpool.Pool
	workerPool *pgxpool.Pool
}

func New(workerCount int) (*Database, error) {
	fmt.Println("connecting to Postgres...")
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}
	config.MaxConns = 5
	apiPool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Postgres api pool: %w", err)
	}

	// pgxpool.NewWithConfig is lazy and connects in the background, so an
	// explicit Ping is required to fail fast when Postgres is down.
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = apiPool.Ping(pingCtx)
	cancel()
	if err != nil {
		apiPool.Close()
		return nil, fmt.Errorf("failed to connect to Postgres: %w", err)
	}

	fmt.Println("connected to Postgres!")

	if err := applyMigrations(databaseURL); err != nil {
		apiPool.Close()
		return nil, err
	}

	config, err = pgxpool.ParseConfig(databaseURL)
	if err != nil {
		apiPool.Close()
		return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}
	config.MaxConns = int32(workerCount)
	workerPool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		apiPool.Close()
		return nil, fmt.Errorf("failed to create Postgres worker pool: %w", err)
	}

	pingCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	err = workerPool.Ping(pingCtx)
	cancel()
	if err != nil {
		apiPool.Close()
		workerPool.Close()
		return nil, fmt.Errorf("failed to connect to Postgres: %w", err)
	}

	return &Database{
		apiPool,
		workerPool,
	}, nil
}

func (d *Database) Close() {
	d.apiPool.Close()
	d.workerPool.Close()
}

func (d *Database) Health() error {
	if d.apiPool.Ping(context.Background()) != nil {
		return ErrDBDown
	}

	return nil
}

func (d *Database) CreateUser(username string, email string, password string) (models.User, error) {
	conn, err := d.apiPool.Acquire(context.Background())
	if err != nil {
		return models.User{}, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	hashedPassword, err := auth.GenerateHashFromPassword(password)
	if err != nil {
		return models.User{}, err
	}

	_, err = conn.Exec(context.Background(), "INSERT INTO Users (username, email, password) VALUES ($1, $2, $3)", username, email, hashedPassword)
	if err != nil {

		return models.User{}, pgxErrorToDatabaseError(err)
	}

	return d.GetUserByEmail(email)
}

func (d *Database) GetUserByEmail(email string) (models.User, error) {
	conn, err := d.apiPool.Acquire(context.Background())
	if err != nil {
		return models.User{}, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()
	var user models.User
	err = conn.QueryRow(context.Background(), "SELECT user_id, username, email, password FROM Users WHERE email = $1", email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)

	return user, err
}

func (d *Database) GetUserByID(id int) (models.User, error) {
	conn, err := d.apiPool.Acquire(context.Background())
	if err != nil {
		return models.User{}, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	var user models.User
	err = conn.QueryRow(context.Background(), "SELECT user_id, username, email, password FROM Users WHERE user_id = $1", id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)

	return user, err
}

// Adds a metadata entry for an uploaded photo and returns it's UUID
func (d *Database) CreatePhotoMetadata(size int64, mimeType string, ownerID int) (string, error) {
	conn, err := d.apiPool.Acquire(context.Background())
	if err != nil {
		return "", pgxErrorToDatabaseError(err)
	}
	defer conn.Release()
	var photo_id string
	err = conn.QueryRow(
		context.Background(),
		"INSERT INTO Photos (photo_id, size, mime_type, owner_id, uploaded_timestamp, filepath) VALUES (uuidv7(), $1, $2, $3, $4, $5) RETURNING photo_id",
		size,
		mimeType,
		ownerID,
		time.Now(),
		"",
	).Scan(&photo_id)

	if err != nil {

		return "", pgxErrorToDatabaseError(err)
	}

	return photo_id, nil
}

func (d *Database) UpdatePhotoMetadataFilePath(uuid string, filepath string) error {
	conn, err := d.apiPool.Acquire(context.Background())
	if err != nil {
		return pgxErrorToDatabaseError(err)
	}
	defer conn.Release()
	_, err = conn.Exec(context.Background(), "UPDATE Photos SET filepath = $1 WHERE photo_id = $2", filepath, uuid)
	return err

}

func (d *Database) SetPhotoThumbnailPath(uuid string, path string) error {
	conn, err := d.apiPool.Acquire(context.Background())
	if err != nil {
		return pgxErrorToDatabaseError(err)
	}
	defer conn.Release()
	_, err = conn.Exec(context.Background(), "UPDATE Photos SET thumbnail_filepath = $1 WHERE photo_id = $2", path, uuid)
	return err
}

func (d *Database) SetPhotoThumbnailJobId(uuid string, jobId int) error {
	conn, err := d.apiPool.Acquire(context.Background())
	if err != nil {
		return pgxErrorToDatabaseError(err)
	}
	defer conn.Release()
	_, err = conn.Exec(context.Background(), "UPDATE Photos SET thumbnail_job_id = $1 WHERE photo_id = $2", jobId, uuid)
	return err
}

func (d *Database) SetPhotoEXIFData(uuid string, latitude *float64, longitude *float64, takenAt *time.Time, cameraModel *string) error {
	conn, err := d.apiPool.Acquire(context.Background())
	if err != nil {
		return pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), "UPDATE Photos SET location = ST_SetSRID(ST_MakePoint($1, $2), 4326), taken_timestamp = $3, camera_model = $4 WHERE photo_id = $5",
		longitude, latitude, takenAt, cameraModel, uuid)
	return pgxErrorToDatabaseError(err)
}

func (d *Database) GetPhotoMetadataFromID(uuid string) (models.PhotoMetadata, error) {
	conn, err := d.apiPool.Acquire(context.Background())
	if err != nil {
		return models.PhotoMetadata{}, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()
	var metadata models.PhotoMetadata
	err = conn.QueryRow(context.Background(), `SELECT 
			photo_id, owner_id, filepath, thumbnail_filepath, thumbnail_job_id, uploaded_timestamp, size, mime_type, ST_Y(location::geometry) AS latitude, ST_X(location::geometry) AS longitude, taken_timestamp, camera_model
		FROM Photos WHERE photo_id = $1`, uuid).
		Scan(&metadata.ID, &metadata.OwnerID, &metadata.Filepath, &metadata.ThumbnailFilepath,
			&metadata.ThumbnailJobId, &metadata.Uploaded, &metadata.Size, &metadata.MimeType,
			&metadata.EXIFCoordinates.Latitude, &metadata.EXIFCoordinates.Longitude,
			&metadata.EXIFTakenAt, &metadata.EXIFCameraModel)

	metadata.Permalink = fmt.Sprintf("/storage/%s", filepath.Base(metadata.Filepath))
	metadata.ThumbnailPermalink = fmt.Sprintf("/storage/thumbnails/%s", filepath.Base(metadata.ThumbnailFilepath))

	return metadata, pgxErrorToDatabaseError(err)
}

func (d *Database) GetUserPhotos(ownerId int) ([]models.PhotoMetadata, error) {
	conn, err := d.apiPool.Acquire(context.Background())
	if err != nil {
		return nil, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()
	var result []models.PhotoMetadata = make([]models.PhotoMetadata, 0)
	rows, err := conn.Query(context.Background(), `SELECT photo_id, owner_id, filepath, thumbnail_filepath, thumbnail_job_id, uploaded_timestamp, size, mime_type, ST_Y(location::geometry) AS lat, ST_X(location::geometry) AS lng, taken_timestamp, camera_model FROM Photos WHERE owner_id = $1`, ownerId)
	if err != nil {
		return nil, pgxErrorToDatabaseError(err)
	}

	for rows.Next() {
		var metadata models.PhotoMetadata
		err = rows.Scan(&metadata.ID, &metadata.OwnerID, &metadata.Filepath, &metadata.ThumbnailFilepath, &metadata.ThumbnailJobId, &metadata.Uploaded, &metadata.Size, &metadata.MimeType, &metadata.EXIFCoordinates.Latitude, &metadata.EXIFCoordinates.Longitude, &metadata.EXIFTakenAt, &metadata.EXIFCameraModel)
		if err != nil {
			return nil, pgxErrorToDatabaseError(err)
		}

		metadata.Permalink = fmt.Sprintf("/storage/%s", filepath.Base(metadata.Filepath))
		metadata.ThumbnailPermalink = fmt.Sprintf("/storage/thumbnails/%s", filepath.Base(metadata.ThumbnailFilepath))

		result = append(result, metadata)
	}

	return result, nil
}

func (d *Database) DeletePhoto(uuid string) error {
	conn, err := d.apiPool.Acquire(context.Background())
	if err != nil {
		return pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	metadata, err := d.GetPhotoMetadataFromID(uuid)
	if err != nil {
		return err
	}

	os.Remove(metadata.Filepath)
	os.Remove(metadata.ThumbnailFilepath)

	_, err = conn.Exec(context.Background(), "DELETE FROM Photos WHERE photo_id = $1", uuid)

	return pgxErrorToDatabaseError(err)
}

func (d *Database) EnqueueJob(payload models.JobPayload) (int, error) {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return -1, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	var jobId int
	err = conn.QueryRow(context.Background(), "INSERT INTO Jobs (payload) VALUES ($1) RETURNING id", payload).Scan(&jobId)
	if err != nil {
		return -1, pgxErrorToDatabaseError(err)
	}
	return jobId, nil
}

func (d *Database) DequeueJob(batchSize int, maxRetries int) (models.Job, error) {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return models.Job{}, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	var job models.Job
	err = conn.QueryRow(context.Background(), `WITH next_job AS (
    SELECT id
    FROM jobs
    WHERE
        retry_count < $1
        AND (
            status = 'pending'
            OR (status = 'in_progress' AND visible_at <= now())
        )
    ORDER BY created_at
    LIMIT $2
    FOR UPDATE SKIP LOCKED
)
UPDATE jobs
SET status = 'in_progress',
    updated_at = now(),
    visible_at = now() + interval '60 seconds',
    retry_count = retry_count + 1
FROM next_job
WHERE jobs.id = next_job.id
RETURNING jobs.*;`, maxRetries, batchSize).
		Scan(&job.ID, &job.Status, &job.Payload, &job.VisibleAt, &job.RetryCount, &job.CreatedAt, &job.UpdatedAt)

	return job, pgxErrorToDatabaseError(err)
}
func (d *Database) AcknowledgeSuccess(job_id int) error {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return pgxErrorToDatabaseError(err)
	}
	defer conn.Release()
	_, err = conn.Exec(context.Background(), `UPDATE jobs
SET status = 'done',
    updated_at = now()
WHERE id = $1;`, job_id)

	return pgxErrorToDatabaseError(err)
}
func (d *Database) AcknowledgeFailure(job_id int) error {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return pgxErrorToDatabaseError(err)
	}
	defer conn.Release()
	_, err = conn.Exec(context.Background(), `UPDATE jobs
SET status = 'failed',
    updated_at = now()
WHERE id = $1;`, job_id)

	return pgxErrorToDatabaseError(err)
}

func (d *Database) GetJob(job_id int) (models.Job, error) {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return models.Job{}, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	var job models.Job
	err = conn.QueryRow(context.Background(), "SELECT * FROM Jobs WHERE id = $1", job_id).
		Scan(
			&job.ID,
			&job.Status,
			&job.Payload,
			&job.VisibleAt,
			&job.RetryCount,
			&job.CreatedAt,
			&job.UpdatedAt,
		)

	return job, pgxErrorToDatabaseError(err)
}

func (d *Database) CreateTag(name string, ownerId int) (int, error) {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return -1, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	var tagId int
	err = conn.QueryRow(context.Background(), "INSERT INTO Tags (owner_id, name) VALUES ($1, $2) RETURNING id", ownerId, name).Scan(&tagId)
	if err != nil {
		return -1, pgxErrorToDatabaseError(err)
	}

	return tagId, nil
}

func (d *Database) GetTagByName(name string, ownerId int) (models.Tag, error) {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return models.Tag{}, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	var tag models.Tag
	err = conn.QueryRow(context.Background(), "SELECT * FROM Tags WHERE name = $1 AND owner_id = $2", name, ownerId).
		Scan(&tag.ID, &tag.OwnerID, &tag.Name)

	return tag, pgxErrorToDatabaseError(err)
}

func (d *Database) GetTagByID(id int) (models.Tag, error) {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return models.Tag{}, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	var tag models.Tag
	err = conn.QueryRow(context.Background(), "SELECT * FROM Tags WHERE id = $1", id).
		Scan(&tag.ID, &tag.OwnerID, &tag.Name)

	return tag, pgxErrorToDatabaseError(err)
}

func (d *Database) GetUserTags(ownerId int) ([]models.Tag, error) {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return nil, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	var tags []models.Tag = make([]models.Tag, 0)

	rows, err := conn.Query(context.Background(), "SELECT * FROM Tags WHERE owner_id = $1", ownerId)
	if err != nil {
		return nil, pgxErrorToDatabaseError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var tag models.Tag
		err = rows.Scan(&tag.ID, &tag.OwnerID, &tag.Name)
		if err != nil {
			return nil, pgxErrorToDatabaseError(err)
		}

		tags = append(tags, tag)
	}

	return tags, pgxErrorToDatabaseError(rows.Err())
}

func (d *Database) RenameTag(tagId int, newName string) error {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), "UPDATE Tags SET name = $1 WHERE id = $2", newName, tagId)
	return pgxErrorToDatabaseError(err)
}

func (d *Database) DeleteTag(tagId int) error {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), "DELETE FROM Tags WHERE id = $1", tagId)
	return pgxErrorToDatabaseError(err)
}

func (d *Database) AssignPhotoTags(tagIds []int, photoId string) error {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	tx, err := conn.Begin(context.Background())
	if err != nil {
		return pgxErrorToDatabaseError(err)
	}
	defer tx.Rollback(context.Background())

	for _, tagId := range tagIds {
		_, err = tx.Exec(context.Background(), "INSERT INTO TagAssignments (photo_id, tag_id) VALUES ($1, $2)", photoId, tagId)
		if err != nil {
			return pgxErrorToDatabaseError(err)
		}
	}

	return pgxErrorToDatabaseError(tx.Commit(context.Background()))
}

func (d *Database) RemovePhotoTag(tagId int, photoId string) error {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), "DELETE FROM TagAssignments WHERE photo_id = $1 AND tag_id = $2", photoId, tagId)
	return pgxErrorToDatabaseError(err)
}

func (d *Database) GetPhotoTags(photoId string) ([]models.Tag, error) {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return nil, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), `SELECT t.id, t.owner_id, t.name
FROM Tags t
JOIN TagAssignments pt ON pt.tag_id = t.id
WHERE pt.photo_id = $1;`, photoId)
	if err != nil {
		return nil, pgxErrorToDatabaseError(err)
	}
	defer rows.Close()

	tags := make([]models.Tag, 0)
	for rows.Next() {
		var tag models.Tag
		err = rows.Scan(&tag.ID, &tag.OwnerID, &tag.Name)
		if err != nil {
			return nil, pgxErrorToDatabaseError(err)
		}

		tags = append(tags, tag)
	}

	return tags, pgxErrorToDatabaseError(rows.Err())

}

func (d *Database) GetTagPhotos(tagId int) ([]models.PhotoMetadata, error) {
	conn, err := d.workerPool.Acquire(context.Background())
	if err != nil {
		return nil, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), `SELECT
			p.photo_id, p.owner_id, p.filepath, p.thumbnail_filepath, p.thumbnail_job_id, p.uploaded_timestamp, p.size, p.mime_type,
			ST_Y(p.location::geometry) AS latitude, ST_X(p.location::geometry) AS longitude, p.taken_timestamp, p.camera_model
		FROM Photos p
		JOIN TagAssignments ta ON ta.photo_id = p.photo_id
		WHERE ta.tag_id = $1;`, tagId)
	if err != nil {
		return nil, pgxErrorToDatabaseError(err)
	}
	defer rows.Close()

	photos := make([]models.PhotoMetadata, 0)
	for rows.Next() {
		var metadata models.PhotoMetadata
		err = rows.Scan(&metadata.ID, &metadata.OwnerID, &metadata.Filepath, &metadata.ThumbnailFilepath,
			&metadata.ThumbnailJobId, &metadata.Uploaded, &metadata.Size, &metadata.MimeType,
			&metadata.EXIFCoordinates.Latitude, &metadata.EXIFCoordinates.Longitude,
			&metadata.EXIFTakenAt, &metadata.EXIFCameraModel)
		if err != nil {
			return nil, pgxErrorToDatabaseError(err)
		}

		metadata.Permalink = fmt.Sprintf("/storage/%s", filepath.Base(metadata.Filepath))
		metadata.ThumbnailPermalink = fmt.Sprintf("/storage/thumbnails/%s", filepath.Base(metadata.ThumbnailFilepath))

		photos = append(photos, metadata)
	}

	return photos, pgxErrorToDatabaseError(rows.Err())
}
