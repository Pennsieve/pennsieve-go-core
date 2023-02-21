package core

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/rds/auth"
	log "github.com/sirupsen/logrus"
	"os"
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

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		dbHost, dbPort, dbUser, authenticationToken, dbName,
	)

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

// ConnectRDSWithOrg returns a DB instance.
// The Lambda function leverages IAM roles to gain access to the DB Proxy.
// The function DOES set the search_path to the organization schema.
func ConnectRDSWithOrg(orgId int) (*sql.DB, error) {
	db, err := ConnectRDS()

	// Set Search Path to organization
	_, err = db.Exec(fmt.Sprintf("SET search_path = \"%d\";", orgId))
	if err != nil {
		log.Error(fmt.Sprintf("Unable to set search_path to %d.", orgId))
		err := db.Close()
		if err != nil {
			return nil, err
		}
		return nil, err
	}

	return db, err
}
