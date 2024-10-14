package main

import "flag"

const (
	serverPort      = flag.String("worker_port", "", "Port where the worker will serve requests.")
	coordinatorPort = flag.String("coordinator", ":8080", "Network address of the Coordinator port")
)

func main() {
	flag.Parse()

	worker := worker.NewServer(*&serverPort, *&coordinatorPort)
	worker.Start()
}
