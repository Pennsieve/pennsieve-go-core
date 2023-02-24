package pgdb

import (
	"database/sql"
	"github.com/pennsieve/pennsieve-go-core/pkg/core"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

var testDB *sql.DB
var orgId int

func TestMain(m *testing.M) {
	var err error

	orgId = 2
	testDB, err = core.ConnectENVWithOrg(orgId)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	os.Exit(m.Run())
}
