package pgdb

import (
	"os"
	"testing"
)

//var testQueries *Queries
//var testDB *sql.DB

func TestMain(m *testing.M) {
	//var err error

	//orgId := 2
	//testDB, err := core.ConnectENVWithOrg(orgId)
	//if err != nil {
	//	log.Fatal("cannot connect to db:", err)
	//}
	//
	//testQueries = New(testDB)

	os.Exit(m.Run())
}
