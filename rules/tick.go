package rules

import (
	"errors"
	"math/rand"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
)

// GameTick runs the game one tick and updates the state
func GameTick(game *pb.Game) (*pb.GameTick, error) {

	if len(game.Ticks) == 0 {
		return nil, errors.New("invalid game tick, current ticks is empty")
	}

	lastTick := game.Ticks[len(game.Ticks)-1]
	nextTick := &pb.GameTick{
		Turn:   lastTick.Turn + 1,
		Snakes: lastTick.Snakes,
		Food:   lastTick.Food,
	}
	moves := GatherSnakeMoves(time.Duration(game.SnakeTimeout)*time.Millisecond, lastTick)

	// we have all the snake moves now
	// 1. update snake coords
	for update := range moves {
		if update.Err != nil {
			update.Snake.DefaultMove()
		} else {
			update.Snake.Move(update.Move)
		}
	}
	// 2. check for death
	// 	  a - starvation
	//    b - wall collision
	//    c - snake collision
	CheckForDeath(game)
	// 3. game update
	//    a - turn incr
	//    b - reduce health points
	//    c - grow snakes
	//    d - remove eaten food
	//    e - replace eaten food
	for _, s := range nextTick.Snakes {
		s.Health = s.Health - 1
	}

	foodToRemove := []*pb.Point{}
	for _, snake := range nextTick.Snakes { // TODO: This should only be alive snakes
		ate := false
		for _, foodPos := range nextTick.Food {
			if snake.Head().Equals(foodPos) {
				snake.Health = 100
				ate = true
				foodToRemove = append(foodToRemove, foodPos)
			}
		}
		if !ate {
			snake.Body = snake.Body[:len(snake.Body)-1]
		}
	}

	nextTick.Food = updateFood(game.Width, game.Height, lastTick, foodToRemove)
	return nextTick, nil
}

func updateFood(width, height int64, gameTick *pb.GameTick, foodToRemove []*pb.Point) []*pb.Point {
	food := []*pb.Point{}
	for _, foodPos := range gameTick.Food {
		found := false
		for _, r := range foodToRemove {
			if foodPos.Equals(r) {
				found = true
				break
			}
		}

		if !found {
			food = append(food, foodPos)
		}
	}

	for range foodToRemove {
		food = append(food, getUnoccupiedPoint(width, height, gameTick))
	}

	return food
}

func getUnoccupiedPoint(width, height int64, gameTick *pb.GameTick) *pb.Point {
	for {
		x := rand.Int63n(width)
		y := rand.Int63n(height)
		p := &pb.Point{X: x, Y: y}
		for _, f := range gameTick.Food {
			if f.Equals(p) {
				continue
			}
		}

		for _, s := range gameTick.Snakes { // TODO: this should only be alive snakes
			for _, b := range s.Body {
				if b.Equals(p) {
					continue
				}
			}
		}

		return p
	}
}
