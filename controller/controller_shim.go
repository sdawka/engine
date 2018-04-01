package controller

import (
	"context"

	"github.com/battlesnakeio/engine/controller/pb"
	"google.golang.org/grpc"
)

// ServerShim provides a shim between grpc client and server, so the server methods can be called in process as if just a client.
type ServerShim struct {
	server *Server
}

// NewInMemory creates a server shim client
func NewInMemory(server *Server) pb.ControllerClient {
	return &ServerShim{
		server: server,
	}
}

// Lock should lock a specific game using the passed in ID. No writes to the
// game should happen as long as the lock is valid. The game being locked does
// not need to exist.
func (s *ServerShim) Lock(ctx context.Context, req *pb.LockRequest, opts ...grpc.CallOption) (*pb.LockResponse, error) {
	return s.server.Lock(ctx, req)
}

// Unlock should unlock a game, if already unlocked a valid lock token must be
// present
func (s *ServerShim) Unlock(ctx context.Context, req *pb.UnlockRequest, opts ...grpc.CallOption) (*pb.UnlockResponse, error) {
	return s.server.Unlock(ctx, req)
}

// Pop should pop a game that is unlocked and unfished from the queue. It can
// be subject to race conditions where it is locked immediately after, this is
// expected.
func (s *ServerShim) Pop(ctx context.Context, req *pb.PopRequest, opts ...grpc.CallOption) (*pb.PopResponse, error) {
	return s.server.Pop(ctx, req)
}

// Status retreives the status of a game
func (s *ServerShim) Status(ctx context.Context, req *pb.StatusRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error) {
	return s.server.Status(ctx, req)
}

// Start starts an existing game, ready to be picked up by a worker.
func (s *ServerShim) Start(ctx context.Context, req *pb.StartRequest, opts ...grpc.CallOption) (*pb.StartResponse, error) {
	return s.server.Start(ctx, req)
}

// Create creates a new game
func (s *ServerShim) Create(ctx context.Context, req *pb.CreateRequest, opts ...grpc.CallOption) (*pb.CreateResponse, error) {
	return s.server.Create(ctx, req)
}
