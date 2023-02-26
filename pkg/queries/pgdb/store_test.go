package pgdb

import (
	"database/sql"
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

	db2, err := ConnectENVWithOrg(3)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testDB[3] = db2

	os.Exit(m.Run())
}
