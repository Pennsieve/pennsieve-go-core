package core

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

type PostgresAPI interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Close() error
}

// ConnectRDS returns a DB instance.
// The Lambda function leverages IAM roles to gain access to the DB Proxy.
// The function does NOT set the search_path to the organization schema.
// Requires following LAMBDA ENV VARIABLES:
// 		- RDS_PROXY_ENDPOINT
//		- REGION
//		- ENV
func ConnectRDS() (*sql.DB, error) {
	var dbName string = "pennsieve_postgres"
	var dbUser string = fmt.Sprintf("%s_rds_proxy_user", os.Getenv("ENV"))
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
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("PENNSIEVE_DB")
	sslMode := os.Getenv("POSTGRES_SSL_MODE")

	return connect(host, port, user, password, dbName, sslMode)
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

func setOrgSearchPath(db *sql.DB, orgId int) error {

	// Set Search Path to organization
	_, err := db.Exec(fmt.Sprintf("SET search_path = \"%d\";", orgId))
	if err != nil {
		log.Error(fmt.Sprintf("Unable to set search_path to %d.", orgId))
		err := db.Close()
		if err != nil {
			return err
		}
		return err
	}

	return err
}
