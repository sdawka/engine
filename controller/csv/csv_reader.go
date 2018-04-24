package csv

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/battlesnakeio/engine/controller/pb"
)

var errWrongSnakeCount = errors.New("wrong snake count")

// readMetadata extracts the metadata from the json on the first line of the
// file. The line is expected to start with a '#' so that it is ignored by
// standard CSV tools.
func readMetadata(reader *bufio.Reader) (*gameArchive, error) {
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	if len(line) < 2 {
		return nil, errors.New("invalid metadata line")
	}

	// Chop off the '#' character before the json
	metaJSON := line[1:]

	meta := &gameArchive{}
	err = json.Unmarshal(metaJSON, meta)
	return meta, err
}

func parseLine(line string, snakeCount int) ([]string, error) {
	trimmed := strings.TrimSpace(line)
	moves := strings.Split(trimmed, ",")

	if len(moves) != snakeCount {
		return nil, errWrongSnakeCount
	}

	return moves, nil
}

func readTurns(reader *bufio.Reader, snakeCount int) ([][]string, error) {
	turns := [][]string{}
	eof := false

	for !eof {
		line, err := reader.ReadString('\n')
		eof = err == io.EOF
		if err != nil && !eof {
			return nil, err
		}

		moves, err := parseLine(line, snakeCount)

		// Wrong count on last line could mean that the process was interrupted
		// in the middle of writing a turn. Don't consider that an error, just
		// resume from the previous turn.
		if err == errWrongSnakeCount && eof {
			break
		} else if err != nil {
			return nil, err
		}

		turns = append(turns, moves)
	}

	return turns, nil
}

func readCSV(id string) (*gameArchive, [][]string, error) {
	path := getFilePath(id)
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, nil, err
	}

	reader := bufio.NewReader(f)

	// Read metadata from first line
	meta, err := readMetadata(reader)
	if err != nil {
		return nil, nil, err
	}

	// Skip a line because the CSV column headers are not needed
	_, err = reader.ReadBytes('\n')
	if err != nil {
		return nil, nil, err
	}

	// Read turns from rest of the lines
	turns, err := readTurns(reader, len(meta.Snakes))
	if err != nil {
		return nil, nil, err
	}

	return meta, turns, nil
}

// replay executes all the turns to derive the state at each frame.
func replay(meta *gameArchive, turns [][]string) (*pb.Game, []*pb.GameTick) {

}

// ReadGame logs the given game from its CSV file and recalculates the state
// at each frame.
func ReadGame(id string) (*pb.Game, []*pb.GameTick, error) {
	meta, turns, err := readCSV(id)
	if err != nil {
		return nil, nil, err
	}

	game, ticks := replay(meta, turns)
	return game, ticks, nil
}
