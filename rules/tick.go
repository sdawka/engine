package rules

import (
	"math/rand"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
)

// GameTick runs the game one tick and updates the state
func GameTick(game *pb.Game) *pb.GameTick {

	lastMove := game.Ticks[len(game.Ticks)-1]
	nextMove := &pb.GameTick{}
	moves := GatherSnakeMoves(time.Duration(game.SnakeTimeout)*time.Millisecond, lastMove)

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
	CheckForDeath(game)
	// 3. game update
	//    a - turn incr
	//    b - reduce health points
	//    c - grow snakes
	//    d - remove eaten food
	//    e - replace eaten food
	for _, s := range lastMove.Snakes {
		s.Health = s.Health - 1
	}

	foodToRemove := []*pb.Point{}
	for _, snake := range lastMove.Snakes { // TODO: This should only be alive snakes
		ate := false
		for _, foodPos := range lastMove.Food {
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

	lastMove.Food = updateFood(game.Width, game.Height, lastMove, foodToRemove)
	return nextMove
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
