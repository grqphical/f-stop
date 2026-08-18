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
	_ "github.com/joho/godotenv/autoload"
)

func pgxErrorToDatabaseError(err error) error {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if pgErr.Code == "23505" {
			return ErrUniqueConstraint
		} else {
			return err
		}
	}
	return nil
}

type Database struct {
	conn *pgx.Conn
}

func New() *Database {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to Postgres: %v\n", err)
	}

	applyMigrations(os.Getenv("DATABASE_URL"))

	return &Database{
		conn,
	}
}

func (d *Database) Close() {
	d.conn.Close(context.Background())
}

func (d *Database) CreateUser(username string, email string, password string) (models.User, error) {
	hashedPassword, err := auth.GenerateHashFromPassword(password)
	if err != nil {
		return models.User{}, err
	}

	_, err = d.conn.Exec(context.Background(), "INSERT INTO Users (username, email, password) VALUES ($1, $2, $3)", username, email, hashedPassword)
	if err != nil {

		return models.User{}, pgxErrorToDatabaseError(err)
	}

	return d.GetUserByEmail(email)
}

func (d *Database) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	err := d.conn.QueryRow(context.Background(), "SELECT user_id, username, email, password FROM Users WHERE email = $1", email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)

	return user, err
}

func (d *Database) GetUserByID(id int) (models.User, error) {
	var user models.User
	err := d.conn.QueryRow(context.Background(), "SELECT user_id, username, email, password FROM Users WHERE user_id = $1", id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)

	return user, err
}

// Adds a metadata entry for an uploaded photo and returns it's UUID
func (d *Database) CreatePhotoMetadata(size int64, mimeType string, ownerID int) (string, error) {
	var photo_id string
	err := d.conn.QueryRow(
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
	_, err := d.conn.Exec(context.Background(), "UPDATE Photos SET filepath = $1 WHERE photo_id = $2", filepath, uuid)
	return err

}

func (d *Database) GetPhotoMetadataFromID(uuid string) (models.PhotoMetadata, error) {
	var metadata models.PhotoMetadata
	err := d.conn.QueryRow(context.Background(), "SELECT * FROM Photos WHERE photo_id = $1", uuid).
		Scan(&metadata.ID, &metadata.OwnerID, &metadata.Filepath, &metadata.Uploaded, &metadata.Size, &metadata.MimeType)

	metadata.Permalink = fmt.Sprintf("/storage/%s", filepath.Base(metadata.Filepath))

	return metadata, err
}

func (d *Database) DeletePhoto(uuid string) error {
	_, err := d.conn.Exec(context.Background(), "DELETE FROM Photos WHERE photo_id = $1", uuid)
	return err
}
