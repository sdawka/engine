// Package worker provides the actual running of games. It is the core of the
// engine and game logic. It interacts primarily with the controller API writing
// and reading game state.
package worker

import (
	"context"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/battlesnakeio/engine/rules"
	log "github.com/sirupsen/logrus"
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

		gt, err := rules.GameTick(resp.Game)
		if err != nil {
			resp.Game.Status = rules.GameStatusError // TODO: this may need to be changed
			return err
		}

		_, err = w.ControllerClient.AddGameTick(ctx, &pb.AddGameTickRequest{ID: resp.Game.ID, GameTick: gt})
		if err != nil {
			return err
		}

		resp.Game.Ticks = append(resp.Game.Ticks, gt)

		if rules.CheckForGameOver(rules.GameMode(resp.Game.Mode), gt) {
			_, err := w.ControllerClient.EndGame(ctx, &pb.EndGameRequest{ID: resp.Game.ID})
			if err != nil {
				log.WithError(err).WithField("GameID", resp.Game.ID).Error("unable to end game")
			}
			return nil
		}

		turnDelay := time.Duration(resp.Game.TurnTimeout) * time.Millisecond
		remainingDelay := turnDelay - time.Since(start)
		if remainingDelay > 0 {
			time.Sleep(remainingDelay)
		}
	}
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
