// Package controller provides an API available to workers to write games. It
// also provides the internal API for starting games and watching.
package controller

import (
	"context"
	"fmt"
	"net"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/battlesnakeio/engine/rules"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// Pop should pop a game that is unlocked and unfinished from the queue, lock
// the game and return it to the worker to begin processing. This call will
// be polled by the workers.
func (s *Server) Pop(ctx context.Context, _ *pb.PopRequest) (*pb.PopResponse, error) {
	id, err := s.Store.PopGameID(ctx)
	if err != nil {
		return nil, err
	}

	token, err := s.Store.Lock(ctx, id, "")
	if err != nil {
		return nil, err
	}

	return &pb.PopResponse{ID: id, Token: token}, nil
}

// Status retrieves the game state including the last processed game tick.
func (s *Server) Status(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	game, err := s.Store.GetGame(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	var lastTick *pb.GameFrame
	ticks, err := s.Store.ListGameTicks(ctx, req.ID, 1, -1)
	if err != nil {
		return nil, err
	}
	if len(ticks) > 0 {
		lastTick = ticks[0]
	}
	return &pb.StatusResponse{Game: game, LastTick: lastTick}, nil
}

// Start starts the game running, and will make it ready to be picked up by a
// worker.
func (s *Server) Start(ctx context.Context, req *pb.StartRequest) (*pb.StartResponse, error) {
	err := s.Store.SetGameStatus(ctx, req.ID, rules.GameStatusRunning)
	if err != nil {
		return nil, err
	}
	return &pb.StartResponse{}, nil
}

// Create creates a new game, but doesn't start running frames.
func (s *Server) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	game, ticks, err := rules.CreateInitialGame(req)
	if err != nil {
		return nil, err
	}
	err = s.Store.CreateGame(ctx, game, ticks)
	if err != nil {
		return nil, err
	}
	return &pb.CreateResponse{
		ID: game.ID,
	}, nil
}

// AddGameFrame adds a new game frame to the game. A lock must be held for this
// call to succeed.
func (s *Server) AddGameFrame(ctx context.Context, req *pb.AddGameFrameRequest) (*pb.AddGameFrameResponse, error) {
	token := pb.ContextGetLockToken(ctx)

	if req.GameFrame == nil {
		return nil, status.Error(codes.InvalidArgument, "controller: game tick must not be nil")
	}

	// Lock the game again, if this fails, the lock is not valid.
	_, err := s.Store.Lock(ctx, req.ID, token)
	if err != nil {
		return nil, err
	}

	// TODO: Need to check that game tick follows the sequence from the previous
	// tick here.

	err = s.Store.PushGameTick(ctx, req.ID, req.GameFrame)
	if err != nil {
		return nil, err
	}
	game, err := s.Store.GetGame(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return &pb.AddGameFrameResponse{
		Game: game,
	}, nil
}

// ListGameFrames will list all game ticks given a limit and offset.
func (s *Server) ListGameFrames(ctx context.Context, req *pb.ListGameFramesRequest) (*pb.ListGameFramesResponse, error) {
	if req.Limit == 0 {
		req.Limit = 50
	}
	if req.Limit > 50 {
		req.Limit = 50
	}
	ticks, err := s.Store.ListGameTicks(ctx, req.ID, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}
	return &pb.ListGameFramesResponse{
		Ticks: ticks,
		Count: int32(len(ticks)),
	}, nil
}

// EndGame sets the game status to complete. A lock must be held for this call
// to succeed.
func (s *Server) EndGame(ctx context.Context, req *pb.EndGameRequest) (*pb.EndGameResponse, error) {
	token := pb.ContextGetLockToken(ctx)

	// Lock the game again, if this fails, the lock is not valid. We only need
	// the lock for the next part where we set the game status.
	newToken, err := s.Store.Lock(ctx, req.ID, token)
	if err != nil {
		return nil, err
	}
	token = newToken

	err = s.Store.SetGameStatus(ctx, req.ID, rules.GameStatusComplete)
	if err != nil {
		return nil, err
	}

	err = s.Store.Unlock(ctx, req.ID, token)
	if err != nil {
		return nil, err
	}

	return &pb.EndGameResponse{}, nil
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
	s.Wait()
	return fmt.Sprintf("127.0.0.1:%d", s.port)
}

// Wait will wait until the server has started.
func (s *Server) Wait() { <-s.started }
