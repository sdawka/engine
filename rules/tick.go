package rules

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	log "github.com/sirupsen/logrus"
)

// GameTick runs the game one tick and updates the state
func GameTick(game *pb.Game, lastTick *pb.GameTick) (*pb.GameTick, error) {
	if lastTick == nil {
		return nil, fmt.Errorf("rules: invalid state, previous tick is nil")
	}
	nextTick := &pb.GameTick{
		Turn:   lastTick.Turn + 1,
		Snakes: lastTick.Snakes,
		Food:   lastTick.Food,
	}
	duration := time.Duration(game.SnakeTimeout) * time.Millisecond
	log.WithFields(log.Fields{
		"GameID":  game.ID,
		"Turn":    nextTick.Turn,
		"Timeout": duration,
	}).Info("GatherSnakeMoves")
	moves := GatherSnakeMoves(duration, game, lastTick)

	// we have all the snake moves now
	// 1. update snake coords
	updateSnakes(game, nextTick, moves)
	// 2. check for death
	// 	  a - starvation
	//    b - wall collision
	//    c - snake collision
	log.WithFields(log.Fields{
		"GameID": game.ID,
		"Turn":   nextTick.Turn,
	}).Info("check for death")
	deathUpdates := checkForDeath(game.Width, game.Height, nextTick)
	for _, du := range deathUpdates {
		if du.Snake.Death == nil {
			du.Snake.Death = du.Death
		}
	}
	// 3. game update
	//    a - turn incr -- done above when the next tick is created
	//    b - reduce health points
	//    c - grow snakes, and update snake health if they ate
	//    d - shrink snakes that didn't et
	//    e - remove eaten food
	//    f - replace eaten food
	log.WithFields(log.Fields{
		"GameID": game.ID,
		"Turn":   nextTick.Turn,
	}).Info("reduce snake health")
	for _, s := range nextTick.AliveSnakes() {
		s.Health = s.Health - 1
	}

	log.WithFields(log.Fields{
		"GameID": game.ID,
		"Turn":   nextTick.Turn,
	}).Info("handle food")

	foodToRemove := checkForSnakesEating(nextTick)
	nextFood, err := updateFood(game.Width, game.Height, lastTick, foodToRemove)
	if err != nil {
		return nil, err
	}
	nextTick.Food = nextFood
	return nextTick, nil
}

func updateFood(width, height int64, gameTick *pb.GameTick, foodToRemove []*pb.Point) ([]*pb.Point, error) {
	food := []*pb.Point{}
	for _, foodPos := range gameTick.Food {
		found := false
		for _, r := range foodToRemove {
			if foodPos.Equal(r) {
				found = true
				break
			}
		}

		if !found {
			food = append(food, foodPos)
		}
	}

	for range foodToRemove {
		p, err := getUnoccupiedPoint(width, height, gameTick.Food, gameTick.AliveSnakes())
		if err != nil {
			return nil, err
		}
		food = append(food, p)
	}

	return food, nil
}

func getUnoccupiedPoint(width, height int64, food []*pb.Point, snakes []*pb.Snake) (*pb.Point, error) {
	attempts := 0
	for {
		attempts++
		if attempts > 20 {
			return nil, errors.New("unable to find available empty location after 20 attempts")
		}
		x := rand.Int63n(width)
		y := rand.Int63n(height)
		p := &pb.Point{X: x, Y: y}
		for _, f := range food {
			if f.Equal(p) {
				continue
			}
		}

		for _, s := range snakes {
			for _, b := range s.Body {
				if b.Equal(p) {
					continue
				}
			}
		}

		return p, nil
	}
}

func updateSnakes(game *pb.Game, tick *pb.GameTick, moves []*SnakeUpdate) {
	for _, update := range moves {
		if update.Err != nil {
			log.WithFields(log.Fields{
				"GameID":  game.ID,
				"SnakeID": update.Snake.ID,
				"Name":    update.Snake.Name,
				"Turn":    tick.Turn,
			}).Info("Default move")
			update.Snake.DefaultMove()
		} else {
			log.WithFields(log.Fields{
				"GameID":  game.ID,
				"SnakeID": update.Snake.ID,
				"Name":    update.Snake.Name,
				"Turn":    tick.Turn,
				"Move":    update.Move,
			}).Info("Move")
			update.Snake.Move(update.Move)
		}
	}
}

func checkForSnakesEating(tick *pb.GameTick) []*pb.Point {
	foodToRemove := []*pb.Point{}
	for _, snake := range tick.AliveSnakes() {
		ate := false
		for _, foodPos := range tick.Food {
			if snake.Head().Equal(foodPos) {
				snake.Health = 100
				ate = true
				foodToRemove = append(foodToRemove, foodPos)
			}
		}
		if !ate {
			if len(snake.Body) == 0 {
				continue
			}
			snake.Body = snake.Body[:len(snake.Body)-1]
		}
	}
	return foodToRemove
}
