package rules

import (
	"math/rand"

	"github.com/battlesnakeio/engine/controller/pb"
)

// GameTick runs the game one tick and updates the state
func GameTick(game *pb.Game) {
	game.Turn = game.Turn + 1
	for _, s := range game.Snakes {
		s.Health = s.Health - 1
	}

	foodToRemove := []*pb.Point{}
	for _, snake := range game.Snakes { // TODO: This should only be alive snakes
		ate := false
		for _, foodPos := range game.Food {
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

	game.Food = updateFood(game, foodToRemove)
}

func updateFood(game *pb.Game, foodToRemove []*pb.Point) []*pb.Point {
	food := []*pb.Point{}
	for _, foodPos := range game.Food {
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
		food = append(food, getUnoccupiedPoint(game))
	}

	return food
}

func getUnoccupiedPoint(game *pb.Game) *pb.Point {
	for {
		x := rand.Int63n(game.Width)
		y := rand.Int63n(game.Height)
		p := &pb.Point{X: x, Y: y}
		for _, f := range game.Food {
			if f.Equals(p) {
				continue
			}
		}

		for _, s := range game.Snakes { // TODO: this should only be alive snakes
			for _, b := range s.Body {
				if b.Equals(p) {
					continue
				}
			}
		}

		return p
	}
}
