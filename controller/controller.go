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
	if err == ErrNotFound {
		return nil, status.Error(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, err
	}
	return &pb.PopResponse{ID: id}, nil
}

// Status retrieves the game state including the last processed game tick.
func (s *Server) Status(ctx context.Context, req *pb.StatusRequest) (*pb.StatusResponse, error) {
	game, err := s.Store.GetGame(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	var lastTick *pb.GameTick
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
	// TODO: Move this to the worker.
	// rules.NotifyGameStart(g)
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

// AddGameTick adds a new game tick to the game, this method must be called
// with a valid lock token passed through the `x-lock-token` header. View the
// pb.ContextWithLockToken(...) function.
func (s *Server) AddGameTick(ctx context.Context, req *pb.AddGameTickRequest) (*pb.AddGameTickResponse, error) {
	token := pb.ContextGetLockToken(ctx)
	err := s.Store.PushGameTick(ctx, req.ID, req.GameTick, token)
	if err != nil {
		return nil, err
	}
	game, err := s.Store.GetGame(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	return &pb.AddGameTickResponse{
		Game: game,
	}, nil
}

// ListGameTicks will list all game ticks given a limit and offset.
func (s *Server) ListGameTicks(ctx context.Context, req *pb.ListGameTicksRequest) (*pb.ListGameTicksResponse, error) {
	ticks, err := s.Store.ListGameTicks(ctx, req.ID, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}
	return &pb.ListGameTicksResponse{
		Ticks: ticks,
		Count: int32(len(ticks)),
	}, nil
}

// EndGame sets the game status to complete.
func (s *Server) EndGame(ctx context.Context, req *pb.EndGameRequest) (*pb.EndGameResponse, error) {
	err := s.Store.SetGameStatus(ctx, req.ID, rules.GameStatusComplete)
	if err != nil {
		return nil, err
	}
	// TODO: Move to the worker.
	// rules.NotifyGameEnd(game)
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
