package filestore

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/battlesnakeio/engine/controller/pb"
)

var openFileWriter = appendOnlyFileWriter

type writer interface {
	WriteString(s string) (int, error)
	Close() error
}

func requireSaveDir() error {
	fmt.Println("requireSaveDir()")
	path := "/home/graeme/.battlesnake/games"
	return os.MkdirAll(path, 0775)
}

func writeLine(w writer, data interface{}) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.WriteString(string(j) + "\n")
	return err
}

func toPoint(p *pb.Point) point {
	return point{
		X: p.X,
		Y: p.Y,
	}
}

func toDeath(d *pb.Death) *death {
	if d == nil {
		return nil
	}

	return &death{
		Cause: d.Cause,
		Turn:  d.Turn,
	}
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
		Death:  toDeath(snake.Death),
	}
}

func toFrame(f *pb.GameFrame) frame {
	snakes := []snakeState{}
	for _, s := range f.Snakes {
		snakes = append(snakes, toSnakeState(s))
	}

	food := []point{}
	for _, f := range f.Food {
		food = append(food, toPoint(f))
	}

	return frame{
		Turn:   f.Turn,
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

func writeFrame(w writer, f *pb.GameFrame) error {
	frame := toFrame(f)
	return writeLine(w, &frame)
}

func writeGameInfo(w writer, game *pb.Game, snakes []*pb.Snake) error {
	info := toGameInfo(game, snakes)
	return writeLine(w, &info)
}

func appendOnlyFileWriter(id string, mustCreate bool) (writer, error) {
	if err := requireSaveDir(); err != nil {
		return nil, err
	}

	path := getFilePath(id)
	flags := os.O_APPEND | os.O_WRONLY | os.O_CREATE
	if mustCreate {
		flags |= os.O_EXCL
	}
	return os.OpenFile(path, flags, 0644)
}
