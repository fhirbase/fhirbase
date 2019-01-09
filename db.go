package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/jackc/pgx"
)

// PgConnectionConfig struct stores credentials for PG connection
type PgConnectionConfig struct {
	SSLMode  string
	Host     string
	Port     uint
	Username string
	Database string
	Password string
}

// PgConnectionConfig holds PG credentials passed from command line
var PgConfig = PgConnectionConfig{}

func GetPgxConnectionConfig(cfg *pgx.ConnConfig) pgx.ConnConfig {
	mainConfig := pgx.ConnConfig{
		TLSConfig: nil,
		Host:      PgConfig.Host,
		Port:      uint16(PgConfig.Port),
		User:      PgConfig.Username,
		Password:  PgConfig.Password,
		Database:  PgConfig.Database,
	}

	// Match libpq default behavior
	if PgConfig.SSLMode == "" {
		PgConfig.SSLMode = "prefer"
	}

	switch PgConfig.SSLMode {
	case "disable":
		mainConfig.UseFallbackTLS = false
		mainConfig.TLSConfig = nil
		mainConfig.FallbackTLSConfig = nil
	case "allow":
		mainConfig.UseFallbackTLS = true
		mainConfig.FallbackTLSConfig = &tls.Config{InsecureSkipVerify: true}
	case "prefer":
		mainConfig.TLSConfig = &tls.Config{InsecureSkipVerify: true}
		mainConfig.UseFallbackTLS = true
		mainConfig.FallbackTLSConfig = nil
	case "require":
		mainConfig.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	case "verify-ca", "verify-full":
		mainConfig.TLSConfig = &tls.Config{
			ServerName: mainConfig.Host,
		}
	default:
		panic("--sslmode param is invalid")
	}

	envConfig, err := pgx.ParseEnvLibpq()

	if err == nil {
		mainConfig = envConfig.Merge(mainConfig)
	}

	if cfg != nil {
		mainConfig = mainConfig.Merge(*cfg)
	}

	return mainConfig
}

// GetConnection connects to database
func GetConnection(cfg *pgx.ConnConfig) *pgx.Conn {
	mainConfig := GetPgxConnectionConfig(cfg)

	connStr := fmt.Sprintf("dbname=%s sslmode=disable user=%s password=%s host=%s port=%d",
		mainConfig.Database, mainConfig.User, mainConfig.Password, mainConfig.Host, mainConfig.Port)

	conn, err := pgx.Connect(mainConfig)

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = conn.Ping(context.Background())

	if err != nil {
		log.Fatalf("Error testing database connection: %v", err)
	}

	fmt.Printf("Connected to database %s\n", connStr)

	return conn
}
