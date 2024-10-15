package scheduler

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sagnikc395/taskly/pkg/common"
)

// CommandRequest is the structure of the request body
type CommandRequest struct {
	Command     string `json:"command"`
	ScheduledAt string `json:"scheduled_at"`
}

//Task that will be scheduled

type Task struct {
	Id          string
	Command     string
	ScheduledAT pgtype.Timestamp
	PickedAt    pgtype.Timestamp
	StartedAt   pgtype.Timestamp
	CompletedAt pgtype.Timestamp
	FailedAt    pgtype.Timestamp
}

// SchedulerServer -> an HTTP Server that will make the tasks discoverable
type SchedulerServer struct {
	serverPort         string
	dbConnectionString string
	dbPool             *pgxpool.Pool
	ctx                context.Context
	cancel             context.CancelFunc
	httpServer         *http.Server
}

// constructor for SchedulerServer
func NewServer(port string, dbConnString string) *SchedulerServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &SchedulerServer{
		serverPort:         port,
		dbConnectionString: dbConnString,
		ctx:                ctx,
		cancel:             cancel,
	}
}

//init and start the schedulerserver

func (s *SchedulerServer) Start() error {
	var err error
	s.dbPool, err = common.ConnectToDB(s.ctx, s.dbConnectionString)
	if err != nil {
		return err
	}

	http.HandleFunc("/schedule", s.handleScheduleTask)
	// add the new route handler
	http.HandleFunc("/status/", s.handleGetTaskStatus)

	s.httpServer = &http.Server{
		Addr: s.serverPort,
	}

	//start a new coroutine to start the server async
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error : %s\n", err)
		}
	}()

	return s.awaitShutDown()
}
