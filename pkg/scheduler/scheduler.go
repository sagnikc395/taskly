package scheduler

import (
	"context"
	"net/http"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
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
