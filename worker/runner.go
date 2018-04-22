package worker

import (
	"context"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/battlesnakeio/engine/rules"
	log "github.com/sirupsen/logrus"
)

// Runner will run an invidual game to completion. It takes a game id and a
// connection to the controller as arguments.
func Runner(ctx context.Context, client pb.ControllerClient, id string) error {
	resp, err := client.Status(ctx, &pb.StatusRequest{ID: id})
	if err != nil {
		return err
	}
	lastTick := resp.LastTick

	for {
		if lastTick != nil && lastTick.Turn == 0 {
			rules.NotifyGameStart(resp.Game, lastTick)
		}
		nextTick, err := rules.GameTick(resp.Game, lastTick)
		if err != nil {
			// This is a GameTick error, we can assume that this is a fatal
			// error and no more game processing can take place at this point.
			log.WithError(err).
				WithField("game", id).
				Error("ending game due to fatal error")
			if _, endErr := client.EndGame(ctx, &pb.EndGameRequest{ID: resp.Game.ID}); endErr != nil {
				log.WithError(endErr).
					WithField("game", id).
					Error("failed to end game after fatal error")
			}
			return err
		}

		log.WithField("game", id).
			WithField("turn", nextTick.Turn).
			Info("adding game tick")
		_, err = client.AddGameTick(ctx, &pb.AddGameTickRequest{
			ID:       resp.Game.ID,
			GameTick: nextTick,
		})
		if err != nil {
			// This is likely a lock error, not to worry here, we can exit.
			return err
		}

		if rules.CheckForGameOver(rules.GameMode(resp.Game.Mode), nextTick) {
			log.WithField("game", id).
				WithField("turn", nextTick.Turn).
				Info("ending game")
			rules.NotifyGameEnd(resp.Game, nextTick)
			_, err := client.EndGame(ctx, &pb.EndGameRequest{ID: resp.Game.ID})
			if err != nil {
				return err
			}
			return nil
		}

		lastTick = nextTick
	}
}
