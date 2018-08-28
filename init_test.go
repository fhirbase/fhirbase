package main

import (
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

func TestFoo(t *testing.T) {
	db := GetConnection(&DbConfig)

	PerformInit(db, "3.3.0")
}
