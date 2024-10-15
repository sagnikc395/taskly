package main

import (
	"flag"
	"log"

	"github.com/sagnikc395/taskly/pkg/common"
	"github.com/sagnikc395/taskly/pkg/scheduler"
)

const (
	schedulerPort = flag.String("scheduler_port", ":0001", "Port on which the Scheduler serves requests")
)

func main() {
	dbConnectionString := common.GetDBConnection()
	schedulerServer := scheduler.NewServer(*&schedulerPort, dbConnectionString)

	err := schedulerServer.Start()
	if err != nil {
		log.Fatalf("Error while starting server: %+v", err)
	}
}
