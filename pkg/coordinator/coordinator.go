package coordinator

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sagnikc395/taskly/pkg/common"
	"google.golang.org/grpc"
)

const (
	shutdownTimeout  = 5 * time.Second
	defaultMaxMisses = 1
	scanInterval     = 10 * time.Second
)

//coordinator struct

type CoordinatorServer struct {
	serverPort          string
	listener            net.Listener
	grpcServer          *grpc.Server
	WorkerPool          map[uint32]*workerInfo
	WorkerPoolMutex     sync.Mutex
	WorkerPoolKeys      []uint32
	WorkerPoolKeysMutex sync.RWMutex
	maxHeartbeatMisses  uint8
	heartbeatInterval   time.Duration
	roundRobinIndex     uint32
	dbConnectionString  string
	dbPool              *pgxpool.Pool
	ctx                 context.Context    // The root context for all goroutines
	cancel              context.CancelFunc // Function to cancel the context
	wg                  sync.WaitGroup     // WaitGroup to wait for all goroutines to finish
}

type workerInfo struct {
	heartbeatMisses     uint8
	address             string
	grpcConnection      *grpc.ClientConn
	workerServiceClient pb.workerServiceClient
}

// NewServer init and return a new Server instance
func NewServer(port string, dbConnString string) *CoordinatorServer {
	ctx, cancel := context.WithCancel(context.Background())

	return &CoordinatorServer{
		WorkerPool:         make(map[uint32]*workerInfo),
		maxHeartbeatMisses: defaultMaxMisses,
		heartbeatInterval:  common.DefaultHeartBeat,
		serverPort:         port,
		ctx:                ctx,
		cancel:             cancel,
	}
}

// initaite the servers operations
func (s *CoordinatorServer) Start() error {
	var err error
	go s.manageWorkerPool()

	if err := s.startGRPCServer(); err != nil {
		return fmt.Errorf("gRPC server start failed: %v", err)
	}
	s.dbPool, err = common.ConnectToDB(s.ctx, s.dbConnectionString)
	if err != nil {
		return err
	}
	go s.scanDatabase()
	return s.awaitShutDown()
}
