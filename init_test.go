package main

import (
	"fmt"
	"log"
	"os"
	"testing"

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
		pgDb.Exec(fmt.Sprintf("DROP DATABASE %s;", DbConfig.Database))
		pgDb.Exec(fmt.Sprintf("CREATE DATABASE %s;", DbConfig.Database))

		db := GetConnection(&DbConfig)
		PerformInit(db, version)
		db.Close()
	}
}
