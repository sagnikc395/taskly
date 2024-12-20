package common

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	DefaultHeartBeat = 10 * time.Second
)

// get the default DB Connection string -> utility function that will connect with the
// env vars
func GetDBConnection() string {
	var missingEnvvars []string

	checkEnvVar := func(envVar, envVarName string) {
		if envVar == "" {
			missingEnvvars = append(missingEnvvars, envVarName)
		}
	}

	dbUser := os.Getenv("POSTGRES_USER")
	checkEnvVar(dbUser, "POSTGRES_USER")

	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	checkEnvVar(dbPassword, "POSTGRES_PASSWORD")

	dbName := os.Getenv("POSTGRES_DB")
	checkEnvVar(dbName, "POSTGRES_DB")

	dbHost := os.Getenv("POSTGRES_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	//missing events check
	if len(missingEnvvars) > 0 {
		log.Fatalf("Folloing required env vars are missing")
		for _, v := range missingEnvvars {
			log.Fatalf("%s\n", v)
		}
	}

	return fmt.Sprintf("postgres://%s:%s@%s:5432/%s", dbUser, dbPassword, dbHost, dbName)
}

//actural driver to connect to db , with the given heartbeat time
// to the Postgres pool

func ConnectToDB(ctx context.Context, dbConnString string) (*pgxpool.Pool, error) {
	var dbPool *pgxpool.Pool
	var err error
	retryCount := 0
	for retryCount < 10 {
		dbPool, err = pgxpool.Connect(ctx, dbConnString)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to the database. Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
		retryCount++
	}
	if err != nil {
		log.Printf("Ran out of retries to connect to db (5)")
		return nil, err
	}

	log.Printf("Connected to db")
	return dbPool, nil
}
