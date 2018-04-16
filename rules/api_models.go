package rules

import (
	"github.com/battlesnakeio/engine/controller/pb"
)

// MoveResponse the message format of the move response from a Snake API call
type MoveResponse struct {
	Move string
}

// SnakeRequest the message send for all snake api calls
type SnakeRequest struct {
	Game  Game  `json:"game"`
	Turn  int64 `json:"turn"`
	Board Board `json:"board"`
	You   Snake `json:"you"`
}

// Game represents the current game state
type Game struct {
	ID string `json:"id"`
}

// Board provides information about the game board
type Board struct {
	Height int64    `json:"height"`
	Width  int64    `json:"width"`
	Food   []Coords `json:"food"`
	Snakes []Snake  `json:"snakes"`
}

// Snake represents information about a snake in the game
type Snake struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Health int64    `json:"health"`
	Body   []Coords `json:"body"`
}

// Coords represents a point on the board
type Coords struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

func buildSnakeRequest(game *pb.Game, tick *pb.GameTick, snakeID string) SnakeRequest {
	var you *pb.Snake
	for _, s := range tick.Snakes {
		if s.ID == snakeID {
			you = s
			break
		}
	}
	return SnakeRequest{
		Game: Game{ID: game.ID},
		Turn: tick.Turn,
		Board: Board{
			Height: game.Height,
			Width:  game.Width,
			Food:   convertPoints(tick.Food),
			Snakes: convertSnakes(tick.Snakes),
		},
		You: convertSnake(you),
	}
}

func convertPoints(points []*pb.Point) []Coords {
	coords := []Coords{}

	for _, p := range points {
		coords = append(coords, Coords{X: p.X, Y: p.Y})
	}

	return coords
}

func convertSnakes(pbSnakes []*pb.Snake) []Snake {
	snakes := []Snake{}

	for _, s := range pbSnakes {
		snakes = append(snakes, convertSnake(s))
	}

	return snakes
}

func convertSnake(snake *pb.Snake) Snake {
	return Snake{
		ID:     snake.ID,
		Name:   snake.Name,
		Health: snake.Health,
		Body:   convertPoints(snake.Body),
	}
}
