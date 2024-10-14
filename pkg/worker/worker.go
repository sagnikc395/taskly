package worker

import (
	"context"
	"net"
	"sync"
	"time"

	ga "github.com/sagnikc395/taskly/pkg/grpcapi"
	"google.golang.org/grpc"
)

const (
	taskProcessTime = 5 * time.Second
	workerPoolSize  = 5
)

// worker server here represents a gRPC server for handling worker tasks
type WorkerServer struct {
	id                       uint32
	serverPort               string
	coordinatorAddress       string
	listener                 net.Listener
	grpcError                *grpc.Server
	coordinatorConnection    *grpc.ClientConn
	cooridnatorServiceClient ga.CoordinatorServiceClient
	heartBeatInterval        time.Duration
	taskqueue                chan *ga.TaskRequest
	ReceivedTasks            map[string]*ga.TaskRequest
	ReceivedTasksMutex       sync.Mutex
	ctx                      context.Context    // root context for all goroutines
	cancel                   context.CancelFunc //function to cancel the context
	wg                       sync.WaitGroup     // wait group to wait for all the goroutines to finish
}
