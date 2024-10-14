package main

import "flag"

const (
	coordinatorPort = flag.String("coordinator_port", ":8080", "Port on which the Coordinator serves requests")
)

func main() {
	flag.Parse()
	dbConnString := common.GetDBConnection()
	coordinator := coordinator.NewServer(*&coordinatorPort, dbConnString)
	coordinator.Start()
}
