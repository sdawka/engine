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

func toTickProto(f frame) *pb.GameTick {

}

func toGameTickProtos(frames []frame) []*pb.GameTick {
	ticks := []*pb.GameTick{}
	for _, f := range frames {
		ticks = append(ticks, toTickProto(f))
	}
	return ticks
}

func toGameProtos(archive gameArchive) (*pb.Game, []*pb.GameTick) {
	game := pb.Game{
		ID:           archive.info.ID,
		Status:       rules.GameStatusStopped,
		Width:        archive.info.Width,
		Height:       archive.info.Height,
		SnakeTimeout: 200, // TO DO
		TurnTimeout:  100, // TO DO
		Mode:         string(rules.GameModeMultiPlayer),
	}

	ticks := toGameTickProtos(archive.frames)

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
