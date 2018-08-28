package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx"
)

// PgConnectionConfig struct stores credentials for PG connection
type PgConnectionConfig struct {
	Host     string
	Port     uint
	Username string
	Database string
	Password string
}

// PgConnectionConfig holds PG credentials passed from command line
var PgConfig = PgConnectionConfig{}

// GetConnection connects to database
func GetConnection(cfg *pgx.ConnConfig) *pgx.Conn {
	mainConfig := pgx.ConnConfig{
		TLSConfig: nil,
		Host:      PgConfig.Host,
		Port:      uint16(PgConfig.Port),
		User:      PgConfig.Username,
		Password:  PgConfig.Password,
	}

	if cfg != nil {
		mainConfig = mainConfig.Merge(*cfg)
	}

	// envConfig, _ := pgx.ParseEnvLibpq()
	// mainConfig.Merge(envConfig)

	connStr := fmt.Sprintf("dbname=%s sslmode=disable user=%s password=%s host=%s port=%d",
		mainConfig.Database, mainConfig.User, mainConfig.Password, mainConfig.Host, mainConfig.Port)

	log.Printf("Connecting to database: %s", connStr)

	conn, err := pgx.Connect(mainConfig)

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
