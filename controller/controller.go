// Package controller provides an API available to workers to write games. It
// also provides the internal API for starting games and watching.
package controller

import (
	"context"
	"fmt"
	"net"

	"github.com/battlesnakeio/engine/controller/pb"
	"google.golang.org/grpc"
)

// New will initialize a new Server.
func New(store Store) *Server {
	return &Server{
		Store:   store,
		started: make(chan struct{}),
	}
}

// Server is a grpc server for pb.ControllerServer.
type Server struct {
	Store Store

	started chan struct{}
	port    int
}

// Lock should lock a specific game using the passed in ID. No writes to the
// game should happen as long as the lock is valid. The game being locked does
// not need to exist.
func (s *Server) Lock(ctx context.Context, req *pb.LockRequest) (*pb.LockResponse, error) {
	token := pb.ContextGetLockToken(ctx)
	token, err := s.Store.Lock(ctx, req.ID, token)
	if err != nil {
		return nil, err
	}
	return &pb.LockResponse{Token: token}, nil
}

// Unlock should unlock a game, if already unlocked a valid lock token must be
// present
func (s *Server) Unlock(ctx context.Context, req *pb.UnlockRequest) (*pb.UnlockResponse, error) {
	token := pb.ContextGetLockToken(ctx)
	err := s.Store.Unlock(ctx, req.ID, token)
	if err != nil {
		return nil, err
	}
	return &pb.UnlockResponse{}, nil
}

// Pop should pop a game that is unlocked and unfished from the queue. It can
// be subject to race conditions where it is locked immediately after, this is
// expected.
func (s *Server) Pop(ctx context.Context, _ *pb.PopRequest) (*pb.PopResponse, error) {
	id, err := s.Store.PopGameID(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.PopResponse{ID: id}, nil
}

// Get should fetch the game state.
func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	game, err := s.Store.GetGame(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return &pb.GetResponse{Game: game}, nil
}

// Start inserts a new game to be picked up by a worker.
func (s *Server) Start(ctx context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	err := s.Store.PutGame(ctx, req.Game)
	if err != nil {
		return nil, err
	}
	return &pb.StartResponse{}, nil
}

// Serve will intantiate a grpc server.
func (s *Server) Serve(listen string) error {
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	s.port = lis.Addr().(*net.TCPAddr).Port
	srv := grpc.NewServer()
	pb.RegisterControllerServer(srv, s)
	close(s.started)
	return srv.Serve(lis)
}

// DialAddress will return a localhost address to reach the server. This is
// useful if the server will select it's own port.
func (s *Server) DialAddress() string {
	<-s.started
	return fmt.Sprintf("127.0.0.1:%d", s.port)
}

// Wait will wait until the server has started.
func (s *Server) Wait() { <-s.started }
