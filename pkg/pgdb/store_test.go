package pgdb

import (
	"database/sql"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

var testDB *sql.DB
var orgId int

func TestMain(m *testing.M) {
	var err error

	orgId = 1
	testDB, err = ConnectENVWithOrg(orgId)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	os.Exit(m.Run())
}
