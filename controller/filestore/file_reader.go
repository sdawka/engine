package filestore

import (
	"bufio"
	"encoding/json"
	"io"
	"os"

	"github.com/battlesnakeio/engine/rules"

	"github.com/battlesnakeio/engine/controller/pb"
)

func readLine(r *bufio.Reader, out interface{}) (bool, error) {
	bytes, err := r.ReadBytes('\n')
	eof := err == io.EOF

	if err != nil && !eof {
		return false, err
	}

	if err = json.Unmarshal(bytes, out); err != nil {
		return false, err
	}

	return !eof, nil
}

func readGameInfo(r *bufio.Reader) (gameInfo, bool, error) {
	info := gameInfo{}
	more, err := readLine(r, &info)
	return info, more, err
}

func readFrame(r *bufio.Reader) (frame, bool, error) {
	f := frame{}
	more, err := readLine(r, &f)
	return f, more, err
}

func readArchive(id string) (gameArchive, error) {
	f, err := os.OpenFile(getFilePath(id), os.O_RDONLY, 0644)
	if err != nil {
		return gameArchive{}, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	info, moreLines, err := readGameInfo(reader)
	if err != nil {
		return gameArchive{}, nil
	}

	frames := []frame{}
	for moreLines {
		f, more, err := readFrame(reader)
		moreLines = more
		if err != nil {
			return gameArchive{}, nil
		}

		frames = append(frames, f)
	}

	return gameArchive{
		info:   info,
		frames: frames,
	}, nil
}

func toSnakeProto(s snakeState, info snakeInfo) *pb.Snake {
	body := []*pb.Point{}
	for _, p := range s.Body {
		body = append(body, toPointProto(p))
	}

	return &pb.Snake{
		ID:     s.ID,
		Name:   info.Name,
		URL:    info.URL,
		Body:   body,
		Health: s.Health,
		Death:  toDeathProto(s.Death),
		Color:  info.Color,
	}
}

func toPointProto(p point) *pb.Point {
	return &pb.Point{
		X: p.X,
		Y: p.Y,
	}
}

func toDeathProto(d *death) *pb.Death {
	if d == nil {
		return nil
	}

	return &pb.Death{
		Cause: d.Cause,
		Turn:  d.Turn,
	}
}

func toTickProto(f frame, infoMap map[string]snakeInfo) *pb.GameTick {
	food := []*pb.Point{}
	for _, p := range f.Food {
		food = append(food, toPointProto(p))
	}

	snakes := []*pb.Snake{}
	for _, s := range f.Snakes {
		snakes = append(snakes, toSnakeProto(s, infoMap[s.ID]))
	}

	return &pb.GameTick{
		Turn:   f.Turn,
		Food:   food,
		Snakes: snakes,
	}
}

func toGameTickProtos(frames []frame, infoMap map[string]snakeInfo) []*pb.GameTick {
	ticks := []*pb.GameTick{}
	for _, f := range frames {
		ticks = append(ticks, toTickProto(f, infoMap))
	}
	return ticks
}

func snakeInfoMap(archive gameArchive) map[string]snakeInfo {
	ret := make(map[string]snakeInfo)
	for _, s := range archive.info.Snakes {
		ret[s.ID] = s
	}
	return ret
}

func toGameProtos(archive gameArchive) (*pb.Game, []*pb.GameTick) {
	infoMap := snakeInfoMap(archive)

	game := pb.Game{
		ID:           archive.info.ID,
		Status:       rules.GameStatusStopped,
		Width:        archive.info.Width,
		Height:       archive.info.Height,
		SnakeTimeout: 200, // TO DO
		TurnTimeout:  100, // TO DO
		Mode:         string(rules.GameModeMultiPlayer),
	}

	ticks := toGameTickProtos(archive.frames, infoMap)

	return &game, ticks
}

// ReadGame loads the game stored in a file with the given id.
func ReadGame(id string) (*pb.Game, []*pb.GameTick, error) {
	archive, err := readArchive(id)
	if err != nil {
		return nil, nil, err
	}

	game, ticks := toGameProtos(archive)
	return game, ticks, nil
}
