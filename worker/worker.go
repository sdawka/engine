// Package worker provides the actual running of games. It is the core of the
// engine and game logic. It interacts primarily with the controller API writing
// and reading game state.
package worker

import (
	"context"
	"log"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/battlesnakeio/engine/rules"
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
	resp, err := w.ControllerClient.Status(ctx, &pb.StatusRequest{ID: id})
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		start := time.Now()
		moves := rules.GatherSnakeMoves(resp.Game)

		// we have all the snake moves now
		// 1. update snake coords
		for update := range moves {
			if update.Err != nil {
				update.Snake.DefaultMove()
			}
			update.Snake.Move(update.Move)
		}
		// 2. check for death
		// 	  a - starvation
		//    b - wall collision
		//    c - snake collision
		rules.CheckForDeath(resp.Game)
		// 3. game update
		//    a - turn incr
		//    b - reduce health points
		//    c - grow snakes
		//    d - remove eaten food
		//    e - replace eaten food
		rules.GameTick(resp.Game)

		_, err = w.ControllerClient.Update(ctx, &pb.UpdateRequest{Game: resp.Game})
		if err != nil {
			return err
		}

		turnDelay := time.Duration(resp.Game.TurnTimeout) * time.Millisecond
		remainingDelay := turnDelay - time.Since(start)
		if remainingDelay > 0 {
			time.Sleep(remainingDelay)
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
