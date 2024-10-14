package common

import (
	"fmt"
	"log"
	"os"
	"time"
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
