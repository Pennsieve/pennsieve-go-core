package pgdb

import (
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/nodeId"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

var testDB map[int]*sql.DB

var publishingTeamId int64
var publishingTeamName string
var publishingTeamNodeId string

func logFatalError(message string, err error) {
	log.Fatal(fmt.Sprintf("%s (error: %+v)", message, err))
}

func TestMain(m *testing.M) {
	var err error

	testDB = make(map[int]*sql.DB)

	db0, err := ConnectENV()
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testDB[0] = db0
	addUsers(db0)
	addUsersToOrganizations(db0)
	addOrganization(db0)
	addPublishingTeam(db0)
	addTeamToOrganization(db0, 1, publishingTeamId, "publishers")
	addUserToTeam(db0, 1001, publishingTeamId)

	db1, err := ConnectENVWithOrg(1)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testDB[1] = db1

	// Add stub dataset for testing against other datasets within same org.
	addDataset(db1)

	db2, err := ConnectENVWithOrg(2)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testDB[2] = db2
	addDatasetStatus(db2)
	addDataUseAgreements(db2)

	db3, err := ConnectENVWithOrg(3)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	testDB[3] = db3
	addDatasetStatus(db3)
	addDataUseAgreements(db3)
	addContributors(db3)

	os.Exit(m.Run())
}

func addOrganization(db *sql.DB) {
	statement := "INSERT INTO pennsieve.organizations (id, node_id, name, slug, encryption_key_id) VALUES ($1, $2, $3, $4, $5)"
	_, err := db.Exec(statement, 42, "N:organization:2b809c6f-9941-47a2-9593-9540fbe77ff1", "Ultimate", "ultimate", "NO_ENCRYPTION_KEY")
	if err != nil {
		log.Fatal(fmt.Sprintf("unable to add organization"))
	}
}

func addUsers(db *sql.DB) {
	type Users struct {
		userId         int64
		nodeId         string
		emailAddress   string
		firstName      string
		lastName       string
		preferredOrgId int64
		cognitoId      string
		isSuperAdmin   string
	}

	users := []Users{
		{userId: 1001, nodeId: "N:user:1", emailAddress: "user1@pennsieve.org", firstName: "one", lastName: "user", preferredOrgId: 1, cognitoId: "11111111-1111-1111-1111-111111111111", isSuperAdmin: "f"},
		{userId: 1002, nodeId: "N:user:2", emailAddress: "user2@pennsieve.org", firstName: "two", lastName: "user", preferredOrgId: 2, cognitoId: "22222222-2222-2222-2222-222222222222", isSuperAdmin: "f"},
		{userId: 1003, nodeId: "N:user:3", emailAddress: "user3@pennsieve.org", firstName: "three", lastName: "user", preferredOrgId: 3, cognitoId: "33333333-3333-3333-3333-333333333333", isSuperAdmin: "f"},
		{userId: 1004, nodeId: "N:user:4", emailAddress: "user4@pennsieve.org", firstName: "four", lastName: "user", preferredOrgId: 3, cognitoId: "44444444-4444-4444-4444-444444444444", isSuperAdmin: "f"},
	}

	statement := "INSERT INTO pennsieve.users (id, node_id, email, first_name, last_name, preferred_org_id, cognito_id, is_super_admin)" +
		"VALUES($1, $2, $3, $4, $5, $6, $7, $8);"

	for _, user := range users {
		_, err := db.Exec(statement, user.userId, user.nodeId, user.emailAddress, user.firstName, user.lastName, user.preferredOrgId, user.cognitoId, user.isSuperAdmin)
		if err != nil {
			log.Fatal(fmt.Sprintf("unable to add user with userId: %d", user.userId))
		}
	}
}

func addUsersToOrganizations(db *sql.DB) {
	type OrgUserPermission struct {
		organizationId int64
		userId         int64
		permissionBit  pgdb.DbPermission
	}

	memberships := []OrgUserPermission{
		{organizationId: 1, userId: 1001, permissionBit: pgdb.Delete},
		{organizationId: 2, userId: 1002, permissionBit: pgdb.Delete},
		{organizationId: 3, userId: 1003, permissionBit: pgdb.Delete},
		{organizationId: 3, userId: 1004, permissionBit: pgdb.Delete},
	}

	statement := "INSERT INTO pennsieve.organization_user (organization_id, user_id, permission_bit) VALUES ($1, $2, $3)"

	for _, membership := range memberships {
		_, err := db.Exec(statement, membership.organizationId, membership.userId, membership.permissionBit)
		if err != nil {
			log.Fatal(fmt.Sprintf("unable to add organization membership org: %d user : %d perm: %d",
				membership.organizationId, membership.userId, membership.permissionBit))
		}
	}
}

func addPublishingTeam(db *sql.DB) {
	publishingTeamId = 999
	publishingTeamName = "Publishers"
	publishingTeamNodeId = "N:team:10534606-9c55-409a-99bd-503f3e873c69"
	statement := "INSERT INTO pennsieve.teams (id, name, node_id) VALUES ($1, $2, $3);"

	_, err := db.Exec(statement, publishingTeamId, publishingTeamName, publishingTeamNodeId)
	if err != nil {
		logFatalError("failed to add publishing team", err)
	}
}

func addTeamToOrganization(db *sql.DB, orgId int64, teamId int64, teamType string) {
	statement := "INSERT INTO pennsieve.organization_team (organization_id, team_id, permission_bit, system_team_type) " +
		"VALUES ($1, $2, $3, NULLIF($4, ''))"

	_, err := db.Exec(statement, orgId, teamId, 16, teamType)
	if err != nil {
		logFatalError("failed to add team to organization", err)
	}
}

func addUserToTeam(db *sql.DB, userId int64, teamId int64) {
	statement := "INSERT INTO pennsieve.team_user (team_id, user_id, permission_bit) VALUES($1, $2, $3)"

	_, err := db.Exec(statement, teamId, userId, 8)
	if err != nil {
		logFatalError("failed to add user to team", err)
	}
}

func addContributors(db *sql.DB) {
	type Contrib struct {
		userId    int64
		firstName string
		lastName  string
		email     string
		orcid     string
	}

	contribList := []Contrib{
		{
			userId:    1004,
			firstName: "four",
			lastName:  "user",
			email:     "user4@pennsieve.org",
			orcid:     "0000-0000-0000-4444",
		},
		{
			userId:    0,
			firstName: "external",
			lastName:  "user",
			email:     "user@external.org",
			orcid:     "0000-0000-0000-1234",
		},
	}

	var err error
	for _, contrib := range contribList {
		if contrib.userId > 0 {
			_, err = db.Exec("INSERT INTO contributors (user_id, first_name, last_name, email, orcid) VALUES($1, $2, $3, $4, $5)",
				contrib.userId, contrib.firstName, contrib.lastName, contrib.email, contrib.orcid)
		} else {
			_, err = db.Exec("INSERT INTO contributors (first_name, last_name, email, orcid) VALUES($1, $2, $3, $4)",
				contrib.firstName, contrib.lastName, contrib.email, contrib.orcid)
		}
		if err != nil {
			log.Fatal(fmt.Sprintf("unable to add contributor: %+v (error: %+v)", contrib, err))
		}
	}
}

func addDatasetStatus(db *sql.DB) {
	var err error
	_, err = db.Exec("INSERT INTO dataset_status (id, name, display_name, original_name, color) VALUES (1001, 'Initial_Status', 'Initial Status', 'Initial_Status', '#000000')")
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to add dataset_status (1) for test: %v", err))
	}
	_, err = db.Exec("INSERT INTO dataset_status (id, name, display_name, original_name, color) VALUES (1002, 'Second_Status', 'Second Status', 'Second_Status', '#000000')")
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to add dataset_status (2) for test: %v", err))
	}
}

func addDataUseAgreements(db *sql.DB) {
	var err error
	_, err = db.Exec("INSERT INTO data_use_agreements (id, name, body, description, is_default) VALUES (1001, 'Data Use Agreement General', 'Use the data any way you like.', 'for general, unrestricted use', true)")
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to add data_use_agreements (1) for test: %v", err))
	}
	_, err = db.Exec("INSERT INTO data_use_agreements (id, name, body, description, is_default) VALUES (1002, 'Data Use Agreement Restricted', 'Use the data as permitted.', 'requires authorization', false)")
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to add data_use_agreements (2) for test: %v", err))
	}
}

func addDataset(db *sql.DB) {
	_, err := db.Exec("INSERT INTO datasets (id, name, node_id, state,status_id) VALUES (2,'Test Dataset 2', 'N:Dataset:00000000-6803-4a67-bf20-83076774a5c7', 'READY', 1)")
	if err != nil {
		log.Fatal(fmt.Sprintf("Unable to add dataset for test: %v", err))
	}

}

func addTestDataset(db *sql.DB, datasetName string) {
	datasetNodeId := nodeId.NodeId(nodeId.DataSetCode)
	datasetState := "READY"
	datasetStatusId := 1
	statement := fmt.Sprintf("INSERT INTO datasets (name, node_id, state, status_id) VALUES ('%s', '%s', '%s', %d);",
		datasetName, datasetNodeId, datasetState, datasetStatusId)
	_, err := db.Exec(statement)
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
