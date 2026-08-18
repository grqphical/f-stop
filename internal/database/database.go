package database

import (
	"context"
	"log"
	"os"

	"github.com/grqphical/f-stop/internal/auth"
	"github.com/grqphical/f-stop/internal/models"
	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
)

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
		return models.User{}, err
	}

	var user models.User
	err = d.conn.QueryRow(context.Background(), "SELECT user_id, username, email FROM Users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Email)

	return user, nil
}
