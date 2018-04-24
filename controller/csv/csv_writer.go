package csv

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/battlesnakeio/engine/controller/pb"
)

func findSnake(snakeID string, frame *pb.GameTick) *pb.Snake {
	for snakeIndex := range frame.Snakes {
		s := frame.Snakes[snakeIndex]
		if s.ID == snakeID {
			return s
		}
	}
	return nil
}

func directionBetween(a, b *pb.Point) string {
	if b.X < a.X {
		return "l"
	} else if b.X > a.X {
		return "r"
	} else if b.Y < a.Y {
		return "d"
	} else {
		return "u"
	}
}

func findMove(snakeID string, thisFrame, previousFrame *pb.GameTick) string {
	thisFrameSnake := findSnake(snakeID, thisFrame)
	previousFrameSnake := findSnake(snakeID, previousFrame)
	if thisFrameSnake == nil || previousFrameSnake == nil {
		return "_"
	}

	thisHead := thisFrameSnake.Body[0]
	previousHead := previousFrameSnake.Body[0]
	return directionBetween(previousHead, thisHead)
}

func getSnakes(frames []*pb.GameTick) []snakeArchive {
	snakes := []snakeArchive{}

	if len(frames) > 0 {
		for snakeIndex := range frames[0].Snakes {
			s := frames[0].Snakes[snakeIndex]
			snakes = append(snakes, snakeArchive{
				ID:    s.ID,
				Name:  s.Name,
				Color: "#ff0000",
			})
		}
	}

	return snakes
}

func getMoves(tick *pb.GameTick) []string {
	moves := []string{}
	for _, s := range snakes {
		moves = append(moves, findMove(s.ID, thisFrame, previousFrame))
	}
	return move
}

// TO DO: REMOVE THIS UNLESS ACTUALLY NEEDED
func getTurns(snakes []snakeArchive, frames []*pb.GameTick) [][]string {
	turns := [][]string{}

	if len(frames) > 1 {
		for frameIndex := range frames[1:] {
			thisFrame := frames[frameIndex+1]
			previousFrame := frames[frameIndex]
			snakeTurns := []string{}
			for snakeIndex := range snakes {
				s := snakes[snakeIndex]
				move := findMove(s.ID, thisFrame, previousFrame)
				snakeTurns = append(snakeTurns, move)
			}
			turns = append(turns, snakeTurns)
		}
	}

	return turns
}

// TO DO: REMOVE THIS UNLESS ACTUALLY NEEDED
func toArchive(game *pb.Game, frames []*pb.GameTick) (gameArchive, [][]string) {
	snakes := getSnakes(frames)
	turns := getTurns(snakes, frames)
	metadata := gameArchive{
		Board: boardArchive{
			ID:     game.ID,
			Width:  game.Width,
			Height: game.Height,
		},
		Snakes: snakes,
	}

	return metadata, turns
}

func requireSaveDir() error {
	fmt.Println("requireSaveDir()")
	path := "/home/graeme/.battlesnake/games"
	return os.MkdirAll(path, 0775)
}

func writeCSVMetadata(f *os.File, metaJSON []byte) error {
	line := "#" + string(metaJSON) + "\n"
	_, err := f.WriteString(line)
	return err
}

func writeCSVColumnHeaders(f *os.File, snakeCount int) error {
	line := "turn"
	for i := 0; i < snakeCount; i++ {
		line += ",player" + strconv.Itoa(i+1)
	}
	line += "\n"
	_, err := f.WriteString(line)
	return err
}

func writeCSVRow(f *os.File, moves []string, turnNumber int) error {
	line := strconv.Itoa(turnNumber)
	for _, move := range moves {
		line += "," + move
	}
	line += "\n"

	_, err := f.WriteString(line)
	return err
}

func writeTick(f *os.File, tick *pb.GameTick) error {
	return writeCSVRow(f, getMoves(tick), tick.Turn)
}

func appendOnlyFileHandle(id string) (*os.File, error) {
	path := getFilePath(id)
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
}

// TO DO: REMOVE ME I AM A TEMP FUNCTION
func saveGameCSV(id string, metaJSON []byte, turns [][]string) error {
	fmt.Println("saveGameCSV()")
	err := requireSaveDir()
	if err != nil {
		return err
	}

	fmt.Println("writing...")

	f, err := appendOnlyFileHandle(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// First line: commented out metadata.
	err = writeCSVMetadata(f, metaJSON)
	if err != nil {
		return err
	}

	// Second line: column headers
	err = writeCSVColumnHeaders(f, len(turns[0]))
	if err != nil {
		return err
	}

	// Subsequent lines: actual CSV data
	for moveIndex, moves := range turns {
		err = writeCSVRow(f, moves, moveIndex+1)
		if err != nil {
			return err
		}
	}

	return nil
}

func saveGameToFile(game *pb.Game, ticks []*pb.GameTick) error {
	fmt.Println("saveGameToFile()")
	if len(ticks) <= 1 {
		// This game didn't actually happen so don't bother saving anything.
		return nil
	}

	metadata, turns := toArchive(game, ticks)
	metaJSON, err := json.Marshal(metadata)

	if err != nil {
		return err
	}

	return saveGameCSV(game.ID, metaJSON, turns)
}
