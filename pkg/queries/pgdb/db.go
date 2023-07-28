package pgdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	log "github.com/sirupsen/logrus"
)

// From https://dev.to/techschoolguru/a-clean-way-to-implement-database-transaction-in-golang-2ba

// DBTX Default interface with methods that are available for both DB adn TX sessions.
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// New returns a Queries object backed by a DBTX interface (either DB or TX)
func New(db DBTX) *Queries {
	return &Queries{db: db}
}

// Queries is a struct with a db object that implements the DBTX interface.
// This means that db can either be a direct DB connection or a TX transaction.
type Queries struct {
	db DBTX
}

// WithTx Returns a new Queries object wrapped by a transactions.
func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{
		db: tx,
	}
}

func (q *Queries) ShowSearchPath(loc string) {
	var currentSchema string
	rr := q.db.QueryRowContext(context.Background(), "show search_path")
	rr.Scan(&currentSchema)
	fmt.Printf("%s: Search Path: %v\n", loc, currentSchema)
}

func (q *Queries) WithOrg(orgId int) (*Queries, error) {
	err := setOrgSearchPath(q.db, orgId)
	if err != nil {
		return nil, err
	}

	return &Queries{
		db: q.db,
	}, nil
}

// ConnectRDS returns a DB instance.
// The Lambda function leverages IAM roles to gain access to the DB Proxy.
// The function does NOT set the search_path to the organization schema.
// Requires following LAMBDA ENV VARIABLES:
//   - RDS_PROXY_ENDPOINT
//   - REGION
//   - ENV
//
// If ENV is set to DOCKER, the call is redirected to ConnectENV()
func ConnectRDS() (*sql.DB, error) {
	ENV := os.Getenv("ENV")
	DOCKER_ENV := "DOCKER"
	if ENV == DOCKER_ENV {
		return ConnectENV()
	}
	var dbName string = "pennsieve_postgres"
	var dbUser string = fmt.Sprintf("%s_rds_proxy_user", ENV)
	var dbHost string = os.Getenv("RDS_PROXY_ENDPOINT")
	var dbPort int = 5432
	var dbEndpoint string = fmt.Sprintf("%s:%d", dbHost, dbPort)
	var region string = os.Getenv("REGION")

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error: " + err.Error())
	}

	authenticationToken, err := auth.BuildAuthToken(
		context.TODO(), dbEndpoint, region, dbUser, cfg.Credentials)
	if err != nil {
		panic("failed to create authentication token: " + err.Error())
	}

	return connect(dbHost, strconv.Itoa(dbPort), dbUser, authenticationToken, dbName, "")
}

// ConnectRDSWithOrg returns a DB instance.
// The Lambda function leverages IAM roles to gain access to the DB Proxy.
// The function DOES set the search_path to the organization schema.
func ConnectRDSWithOrg(orgId int) (*sql.DB, error) {
	db, err := ConnectRDS()
	if err != nil {
		return nil, err
	}
	err = setOrgSearchPath(db, orgId)
	return db, err
}

// ConnectENV returns a DB instance. Used for testing, it requires the
// following environment variables to be set
// - POSTGRES_HOST
// - POSTGRES_PORT (will default to 5432 if missing)
// - POSTGRES_USER
// - POSTGRES_PASSWORD
// - PENNSIEVE_DB
// - POSTGRES_SSL_MODE (should be set to "disable" if the server is not https, left blank if it is)
func ConnectENV() (*sql.DB, error) {
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "password")
	dbName := getEnv("PENNSIEVE_DB", "postgres")
	sslMode := getEnv("POSTGRES_SSL_MODE", "disable")

	return connect(host, port, user, password, dbName, sslMode)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func ConnectENVWithOrg(orgId int) (*sql.DB, error) {
	db, err := ConnectENV()
	if err != nil {
		return nil, err
	}
	err = setOrgSearchPath(db, orgId)
	return db, err
}

func connect(dbHost string, dbPort string, dbUser string, authenticationToken string, dbName string, sslMode string) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
		dbHost, dbPort, dbUser, authenticationToken, dbName,
	)
	if sslMode != "" {
		dsn = fmt.Sprintf("%s sslmode=%s", dsn, sslMode)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db, err
}

func setOrgSearchPath(db DBTX, orgId int) error {

	// Set Search Path to organization
	ctx := context.Background()
	_, err := db.ExecContext(ctx, fmt.Sprintf("SET search_path = \"%d\";", orgId))
	if err != nil {
		log.Error(fmt.Sprintf("Unable to set search_path to %d.", orgId))
		return err
	}

	return err
}
