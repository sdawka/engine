// Package worker provides the actual running of games. It is the core of the
// engine and game logic. It interacts primarily with the controller API writing
// and reading game state.
package worker

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Worker is the worker interface. It wraps a main Perform functions which is
// where all of the game logic should live.
type Worker struct {
	ControllerClient  pb.ControllerClient
	PollInterval      time.Duration
	HeartbeatInterval time.Duration
}

// perform does the actual work of running a game.
func (w *Worker) perform(ctx context.Context, id string, workerID int) error {
	for i := 0; i < 2; i++ {
		// Stubbed out work emulation.
		log.Printf("[%d] performing work on %s", workerID, id)
		select {
		case <-ctx.Done():
			log.Println("perform closed")
			return ctx.Err()
		case <-time.After(time.Duration(rand.Intn(10)) * time.Second):
		}
	}
	return nil
}

// Run will run the worker in a loop.
func (w *Worker) Run(ctx context.Context, workerID int) {
	for {
		// We are now holding the lock.
		if err := w.run(ctx, workerID); err != nil {
			s, ok := status.FromError(err)
			if !ok || s.Code() != codes.NotFound {
				log.Printf("[%d] run failed: %v", workerID, err)
			}

			select {
			case <-time.After(w.PollInterval):
			case <-ctx.Done():
				return
			}
		}
	}
}

func (w *Worker) run(ctx context.Context, workerID int) error {
	// Pop an item of work.
	var id string
	{
		res, err := w.ControllerClient.Pop(ctx, &pb.PopRequest{})
		if err != nil {
			return err
		}
		id = res.ID
	}

	// Attempt to get the lock initially.
	var token string
	{
		res, err := w.ControllerClient.Lock(ctx, &pb.LockRequest{ID: id})
		if err != nil {
			return err
		}
		token = res.Token
	}

	log.Printf("[%d] acquired lock %s token=%s", workerID, id, token)

	// Get a context with the lock token.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx = pb.ContextWithLockToken(ctx, token)

	defer func() {
		log.Printf("[%d] unlocking %s", workerID, id)
		_, err := w.ControllerClient.Unlock(ctx, &pb.UnlockRequest{ID: id})
		if err != nil {
			log.Printf("[%d] unlock %s failed: %v", workerID, id, err)
		}
	}()

	// Hold the lock, heartbeating every HeartbeatInterval.
	go func() {
		t := time.NewTicker(w.HeartbeatInterval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				_, err := w.ControllerClient.Lock(ctx, &pb.LockRequest{ID: id})
				if err != nil {
					log.Printf("[%d] lock expired during heartbeat %v", workerID, err)
					cancel()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Perform the actual work, this should respect context and Done() rules.
	// Perform should be able to write to storage using the context and have
	// a valid lock for the key.
	return w.perform(ctx, id, workerID)
}
