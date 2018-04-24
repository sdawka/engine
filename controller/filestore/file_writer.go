package filestore

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/battlesnakeio/engine/controller/pb"
)

func requireSaveDir() error {
	fmt.Println("requireSaveDir()")
	path := "/home/graeme/.battlesnake/games"
	return os.MkdirAll(path, 0775)
}

func writeLine(f *os.File, data interface{}) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = f.WriteString(string(j) + "\n")
	return err
}

func toPoint(p *pb.Point) point {
	return point{
		X: p.X,
		Y: p.Y,
	}
}

func isDead(s *pb.Snake) bool {
	return s.Death != nil && s.Death.Turn >= 0
}

func toSnakeState(snake *pb.Snake) snakeState {
	points := []point{}
	for _, p := range snake.Body {
		points = append(points, toPoint(p))
	}

	return snakeState{
		ID:     snake.ID,
		Body:   points,
		Health: snake.Health,
		Dead:   isDead(snake),
	}
}

func toFrame(tick *pb.GameTick) frame {
	snakes := []snakeState{}
	for _, s := range tick.Snakes {
		snakes = append(snakes, toSnakeState(s))
	}

	food := []point{}
	for _, f := range tick.Food {
		food = append(food, toPoint(f))
	}

	return frame{
		Turn:   tick.Turn,
		Snakes: snakes,
		Food:   food,
	}
}

func toSnakeInfo(s *pb.Snake) snakeInfo {
	return snakeInfo{
		ID:    s.ID,
		Name:  s.Name,
		Color: s.Color,
		URL:   s.URL,
	}
}

func toGameInfo(game *pb.Game, snakes []*pb.Snake) gameInfo {
	snakeInfos := []snakeInfo{}
	for _, s := range snakes {
		snakeInfos = append(snakeInfos, toSnakeInfo(s))
	}

	return gameInfo{
		ID:     game.ID,
		Width:  game.Width,
		Height: game.Height,
		Snakes: snakeInfos,
	}
}

func writeTick(f *os.File, tick *pb.GameTick) error {
	frame := toFrame(tick)
	return writeLine(f, &frame)
}

func writeGameInfo(f *os.File, game *pb.Game, snakes []*pb.Snake) error {
	info := toGameInfo(game, snakes)
	return writeLine(f, &info)
}

func appendOnlyFileHandle(id string) (*os.File, error) {
	if err := requireSaveDir(); err != nil {
		return nil, err
	}

	path := getFilePath(id)
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
}
