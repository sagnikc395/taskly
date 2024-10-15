package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	ScheduledAt pgtype.Timestamp
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

// POST handler to add new tasks
func (s *SchedulerServer) handleScheduleTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "POST method only", http.StatusMethodNotAllowed)
		return
	}

	//JSON decode
	var commandReq CommandRequest

	if err := json.NewDecoder(r.Body).Decode(&commandReq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//log
	log.Printf("Received Schedule request: %+v", commandReq)

	// Parse the scheduled_at time
	scheduledTime, err := time.Parse(time.RFC3339, commandReq.ScheduledAt)
	if err != nil {
		http.Error(w, "Invalid date format. Use ISO 8601 format.", http.StatusBadRequest)
		return
	}

	//  scheduled time to Unix timestamp
	unixTimestamp := time.Unix(scheduledTime.Unix(), 0)

	taskId, err := s.insertTaskIntoDB(context.Background(),
		Task{
			Command:     commandReq.Command,
			ScheduledAt: pgtype.Timestamp{Time: unixTimestamp},
		})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to submit task. Error : %s", err.Error()), http.StatusInternalServerError)
		return
	}

	//respond with the parsed data
	response := struct {
		Command     string `json:"command"`
		ScheduledAt int64  `json:"scheduled_at"`
		TaskID      string `json:"task_id"`
	}{
		Command:     commandReq.Command,
		ScheduledAt: scheduledTime.Unix(),
		TaskID:      taskId,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (s *SchedulerServer) handleGetTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	//task id from the query parameters
	taskId := r.URL.Query().Get("task_id")

	//check if the task id is empty

	if taskId == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	//query the db to get the task status

	var task Task
	err := s.dbPool.QueryRow(context.Background(), "SELECT * FROM tasks where ID = $1", taskId).Scan(
		&task.Id,
		&task.Command,
		&task.ScheduledAt,
		&task.PickedAt,
		&task.StartedAt,
		&task.CompletedAt,
		&task.FailedAt,
	)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get task status. Error : %s", err.Error()), http.StatusInternalServerError)
		return
	}

	//prepare response JSON
	response := struct {
		TaskID      string `json:"task_id"`
		Command     string `json:"command"`
		ScheduledAt string `json:"scheduled_at,omitempty"`
		PickedAt    string `json:"picked_at,omitempty"`
		StartedAt   string `json:"started_at,omitempty"`
		CompletedAt string `json:"completed_at,omitempty"`
		FailedAt    string `json:"failed_at,omitempty"`
	}{
		TaskID:      task.Id,
		Command:     task.Command,
		ScheduledAt: "",
		PickedAt:    "",
		StartedAt:   "",
		CompletedAt: "",
		FailedAt:    "",
	}

	//set the scheduled_at time if non-null
	if task.ScheduledAt.Status == 2 {
		response.ScheduledAt = task.ScheduledAt.Time.String()
	}

	// set the picked_at time if not_null
	if task.PickedAt.Status == 2 {
		response.PickedAt = task.PickedAt.Time.String()
	}

	//set the started_at time if non-null
	if task.CompletedAt.Status == 2 {
		response.FailedAt = task.FailedAt.Time.String()
	}

	//convert the response struct to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		return
	}

	//set the header and send response
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)

}

// insertTaskIntoDB inserts a new task into the tasks into the tasks table and returns the autogenerated UUID
func (s *SchedulerServer) insertTaskIntoDB(ctx context.Context, task Task) (string, error) {
	//SQL statement with returning clause
	sqlStatement := "INSERT INTO tasks (command, scheduled_at) VALUES ($1, $2) RETURNING id"

	var insertedId string

	// execute the query and scan the returned id into the insertedId variable
	err := s.dbPool.QueryRow(ctx, sqlStatement, task.Command, task.ScheduledAt.Time).Scan(&insertedId)
	if err != nil {
		return "", err
	}

	//return UUID
	return insertedId, nil
}

func (s *SchedulerServer) awaitShutDown() error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	return s.Stop()
}

// stop gracefully server, and shut down and close the db conn pool
func (s *SchedulerServer) Stop() error {
	s.dbPool.Close()
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(ctx)
	}

	log.Println("🛑 Scheduler server and db pool stopped.")
	return nil
}
