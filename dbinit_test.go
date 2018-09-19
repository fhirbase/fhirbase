package main

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx"
)

var DbConfig pgx.ConnConfig

func TestsSetup() {
	var err error
	DbConfig, err = pgx.ParseEnvLibpq()

	if err != nil {
		log.Printf("Error parsing PG env variables: %v", err)
	}
}

func TestsTeardown() {

}

func TestMain(m *testing.M) {
	TestsSetup()
	retCode := m.Run()

	TestsTeardown()
	os.Exit(retCode)
}

func TestInitAllSchemas(t *testing.T) {
	pgDbCfg := DbConfig.Merge(pgx.ConnConfig{
		Database: "postgres",
	})

	pgDb := GetConnection(&pgDbCfg)

	for _, version := range AvailableSchemas {
		_, err := pgDb.Exec(fmt.Sprintf("DROP DATABASE %s;", DbConfig.Database))

		if err != nil {
			t.Fatalf("Cannot drop database: %v", err)
		}

		_, err = pgDb.Exec(fmt.Sprintf("CREATE DATABASE %s;", DbConfig.Database))

		if err != nil {
			t.Fatalf("Cannot create database: %v", err)
		}

		db := GetConnection(&DbConfig)
		PerformInit(db, version, func(c int, t int64, d time.Duration) {})
		db.Close()
	}
}
