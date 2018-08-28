package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx"
	"log"
)

// PgConnectionConfig struct stores credentials for PG connection
type PgConnectionConfig struct {
	Host     string
	Port     uint
	Username string
	Database string
	Password string
}

// PgConnectionConfig holds PG credentials
var PgConfig = PgConnectionConfig{}

// GetConnection connects to database
func GetConnection() *pgx.Conn {
	connStr := fmt.Sprintf("dbname=%s sslmode=disable user=%s password=%s host=%s port=%d",
		PgConfig.Database, PgConfig.Username, PgConfig.Password, PgConfig.Host, PgConfig.Port)

	log.Printf("Connecting to database: %s", connStr)

	config, _ := pgx.ParseConnectionString(connStr)

	conn, err := pgx.Connect(config)

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = conn.Ping(context.Background())

	if err != nil {
		log.Fatalf("Error testing database connection: %v", err)
	}

	log.Printf("Connected to database")

	return conn
}
