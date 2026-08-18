package database

import (
	"context"
	"log"
	"os"

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

	return &Database{
		conn,
	}
}

func (d *Database) Close() {
	d.conn.Close(context.Background())
}
