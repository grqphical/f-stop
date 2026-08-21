package database

import (
	"context"
	"errors"
	"fmt"
	"log"
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

func New(workerCount int) *Database {
	config, _ := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	config.MaxConns = 5
	apiPool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("failed to connect to Postgres: %v\n", err)
	}

	applyMigrations(os.Getenv("DATABASE_URL"))

	config, _ = pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	config.MaxConns = int32(workerCount)
	workerPool, err := pgxpool.NewWithConfig(context.Background(), config)

	return &Database{
		apiPool,
		workerPool,
	}
}

func (d *Database) Close() {
	d.apiPool.Close()
	d.workerPool.Close()
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

	return metadata, pgxErrorToDatabaseError(err)
}

func (d *Database) GetUserPhotos(ownerId int) ([]models.PhotoMetadata, error) {
	conn, err := d.apiPool.Acquire(context.Background())
	if err != nil {
		return nil, pgxErrorToDatabaseError(err)
	}
	defer conn.Release()
	var result []models.PhotoMetadata = make([]models.PhotoMetadata, 0)
	rows, err := conn.Query(context.Background(), "SELECT * FROM Photos WHERE owner_id = $1", ownerId)
	if err != nil {
		return nil, pgxErrorToDatabaseError(err)
	}

	for rows.Next() {
		var metadata models.PhotoMetadata
		err = rows.Scan(&metadata.ID, &metadata.OwnerID, &metadata.Filepath, &metadata.ThumbnailFilepath, &metadata.ThumbnailJobId, &metadata.Uploaded, &metadata.Size, &metadata.MimeType)
		if err != nil {
			return nil, pgxErrorToDatabaseError(err)
		}

		metadata.Permalink = fmt.Sprintf("/storage/%s", filepath.Base(metadata.Filepath))

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
