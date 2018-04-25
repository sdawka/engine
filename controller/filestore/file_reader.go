package filestore

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/battlesnakeio/engine/rules"

	"github.com/battlesnakeio/engine/controller/pb"
)

var openFileReader = fsOpenFileReader

type reader interface {
	ReadBytes(delim byte) ([]byte, error)
	Close() error
}

type fsReader struct {
	*os.File
	*bufio.Reader
}

func fsOpenFileReader(id string) (reader, error) {
	f, err := os.OpenFile(getFilePath(id), os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &fsReader{
		File:   f,
		Reader: bufio.NewReader(f),
	}, nil
}

func readLine(r reader, out interface{}) (bool, error) {
	bytes, err := r.ReadBytes('\n')
	eof := err == io.EOF

	if err != nil && !eof {
		return false, err
	}

	if !containsJSONObject(string(bytes)) {
		return !eof, errors.New("invalid json")
	}

	if err = json.Unmarshal(bytes, out); err != nil {
		return !eof, err
	}

	return !eof, nil
}

func readGameInfo(r reader) (gameInfo, bool, error) {
	info := gameInfo{}
	more, err := readLine(r, &info)
	return info, more, err
}

func readFrame(r reader) (frame, bool, error) {
	f := frame{}
	more, err := readLine(r, &f)
	return f, more, err
}

func readArchiveHeader(id string) (gameInfo, error) {
	r, err := openFileReader(id)
	if err != nil {
		return gameInfo{}, err
	}
	defer r.Close()

	info, _, err := readGameInfo(r)
	if err != nil {
		return gameInfo{}, err
	}

	return info, nil
}

func containsJSONObject(s string) bool {
	trimmed := strings.TrimSpace(s)
	return len(trimmed) > 1 && trimmed[0] == '{'
}

func readArchive(id string) (gameArchive, error) {
	r, err := openFileReader(id)
	if err != nil {
		return gameArchive{}, err
	}
	defer r.Close()

	// Skip over game info line
	info, moreLines, err := readGameInfo(r)
	if err != nil {
		return gameArchive{}, err
	}

	// Read the actual frames
	frames := []frame{}
	for moreLines {
		f, more, err := readFrame(r)
		moreLines = more
		if err != nil {
			continue
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

func snakeInfoMap(info gameInfo) map[string]snakeInfo {
	ret := make(map[string]snakeInfo)
	for _, s := range info.Snakes {
		ret[s.ID] = s
	}
	return ret
}

func toGameProto(info gameInfo) *pb.Game {
	game := pb.Game{
		ID:           info.ID,
		Status:       rules.GameStatusStopped,
		Width:        info.Width,
		Height:       info.Height,
		SnakeTimeout: int64(200 * time.Millisecond), // TO DO
		TurnTimeout:  int64(100 * time.Millisecond), // TO DO
		Mode:         string(rules.GameModeMultiPlayer),
	}

	return &game
}

// ReadGameFrames loads all the game ticks stored in given file.
func ReadGameFrames(id string) ([]*pb.GameTick, error) {
	archive, err := readArchive(id)
	if err != nil {
		return nil, err
	}

	infoMap := snakeInfoMap(archive.info)
	ticks := toGameTickProtos(archive.frames, infoMap)
	return ticks, nil
}

// ReadGameInfo reads the header info from the given file.
func ReadGameInfo(id string) (*pb.Game, error) {
	info, err := readArchiveHeader(id)
	if err != nil {
		return nil, err
	}

	return toGameProto(info), nil
}
