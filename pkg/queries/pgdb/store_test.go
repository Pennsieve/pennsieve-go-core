package pgdb

import (
	"database/sql"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

var testDB map[int]*sql.DB

func TestMain(m *testing.M) {
	var err error

	testDB = make(map[int]*sql.DB)

	db1, err := ConnectENVWithOrg(1)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testDB[1] = db1

	// Add stub dataset for testing against other datasets within same org.
	addDataset(db1)

	db2, err := ConnectENVWithOrg(3)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testDB[3] = db2

	os.Exit(m.Run())
}

func addDataset(db *sql.DB) {
	_, err := db.Exec("INSERT INTO datasets (id, name, node_id, state,status_id) VALUES (2,'Test Dataset 2', 'N:Dataset:00000000-6803-4a67-bf20-83076774a5c7', 'READY', 1)")
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to add dataset for test: %v", err))
	}

}

// TestStore is the main Test Suite function for Packages.
func TestStore(t *testing.T) {
	for scenario, fn := range map[string]func(
		tt *testing.T, store *SQLStore, orgId int,
	){
		"Test name expansion":                        testNameExpansion,
		"Test inserting packages with mixed parents": testAddingMixedParentPackages,
	} {
		t.Run(scenario, func(t *testing.T) {
			orgId := 1
			store := NewSQLStore(testDB[orgId])
			fn(t, store, orgId)
		})
	}
}
