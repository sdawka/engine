package filestore

import (
	"bufio"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockReader struct {
	*bufio.Reader
}

func (m *mockReader) Close() error {
	return nil
}

func newMockReader(text string) *mockReader {
	return &mockReader{
		Reader: bufio.NewReader(strings.NewReader(text)),
	}
}

type failReader struct{}

func (f *failReader) ReadBytes(delimiter byte) ([]byte, error) {
	return nil, errors.New("FAIL")
}

func (f *failReader) Close() error {
	return errors.New("FAIL")
}

func fileOpener(files map[string]string) func(string) (reader, error) {
	return func(id string) (reader, error) {
		text, ok := files[id]
		if !ok {
			return nil, errors.New("file not found")
		}
		return newMockReader(text), nil
	}
}

func gameInfoTestJSON() string {
	info := gameInfo{
		ID:     "myid",
		Width:  10,
		Height: 12,
		Snakes: []snakeInfo{
			snakeInfo{
				ID:    "snake1",
				Name:  "snake1",
				Color: "yellow",
				URL:   "http://snake1",
			},
			snakeInfo{
				ID:    "snake2",
				Name:  "snake2",
				Color: "green",
				URL:   "http://snake2",
			},
		},
	}

	infoJSON, _ := json.Marshal(info)
	return string(infoJSON) + "\n"
}

func framesTestJSON() string {
	f1 := frame{
		Turn: 1,
		Snakes: []snakeState{
			snakeState{
				ID:     "snake1",
				Body:   []point{point{1, 1}, point{1, 2}},
				Health: 100,
				Death:  nil,
			},
			snakeState{
				ID:     "snake2",
				Body:   []point{point{3, 1}, point{3, 2}},
				Health: 100,
				Death:  nil,
			},
		},
		Food: []point{point{X: 6, Y: 6}},
	}

	f2 := frame{
		Turn: 2,
		Snakes: []snakeState{
			snakeState{
				ID:     "snake1",
				Body:   []point{point{1, 2}, point{1, 3}},
				Health: 99,
				Death:  nil,
			},
			snakeState{
				ID:     "snake2",
				Body:   []point{point{3, 2}, point{3, 3}},
				Health: 99,
				Death:  &death{Cause: "test", Turn: 2},
			},
		},
		Food: []point{point{X: 6, Y: 6}},
	}

	j1, _ := json.Marshal(f1)
	j2, _ := json.Marshal(f2)

	return string(j1) + "\n" + string(j2) + "\n"
}

func TestReadGameFramesBadReader(t *testing.T) {
	openFileReader = func(id string) (reader, error) {
		return &failReader{}, nil
	}
	_, err := ReadGameFrames("myid")

	require.NotNil(t, err)
}

func TestReadGameFramesOpenReaderError(t *testing.T) {
	openFileReader = func(id string) (reader, error) {
		return nil, errors.New("fail")
	}
	_, err := ReadGameFrames("myid")

	require.NotNil(t, err)
}

func TestReadGameFramesWithoutHeader(t *testing.T) {
	j := framesTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	frames, _ := ReadGameFrames("myid")

	require.Len(t, frames, 1, "first frame is in header spot and should be ignored")
}

func TestReadGameFrames(t *testing.T) {
	j := gameInfoTestJSON() + framesTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	frames, _ := ReadGameFrames("myid")

	require.Len(t, frames, 2)
	require.Equal(t, int64(1), frames[0].Turn)
	require.Equal(t, int64(2), frames[1].Turn)
	require.Equal(t, "snake1", frames[0].Snakes[0].ID)
	require.Equal(t, "snake2", frames[0].Snakes[1].ID)
	require.Equal(t, "snake1", frames[1].Snakes[0].ID)
	require.Equal(t, "snake2", frames[1].Snakes[1].ID)
	require.Nil(t, frames[0].Snakes[0].Death)
	require.Nil(t, frames[0].Snakes[1].Death)
	require.Nil(t, frames[1].Snakes[0].Death)
	require.NotNil(t, frames[1].Snakes[1].Death)
}

func testGarbageEnding(t *testing.T, garbage string) {
	j := gameInfoTestJSON() + framesTestJSON() + garbage

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	frames, _ := ReadGameFrames("myid")

	require.Len(t, frames, 2, "3rd frame is invalid and should be ignored")
	require.Equal(t, int64(1), frames[0].Turn)
	require.Equal(t, int64(2), frames[1].Turn)
}

func TestReadGameFramesPlusGarbage(t *testing.T) {
	testGarbageEnding(t, "...")
	testGarbageEnding(t, "{")
	testGarbageEnding(t, "{ foo }")
}

func TestReadGameFramesGarbageAfterHeader(t *testing.T) {
	j := gameInfoTestJSON() + "\n\n{\n" + framesTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	frames, _ := ReadGameFrames("myid")

	require.Len(t, frames, 2, "garbage should be ignored")
	require.Equal(t, int64(1), frames[0].Turn)
	require.Equal(t, int64(2), frames[1].Turn)
}

func TestReadGameFramesEmpty(t *testing.T) {
	j := gameInfoTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	frames, _ := ReadGameFrames("myid")

	require.Len(t, frames, 0)
}

func TestReadGameFramesBlankFile(t *testing.T) {
	openFileReader = fileOpener(map[string]string{
		"myid": "",
	})
	frames, _ := ReadGameFrames("myid")

	require.Len(t, frames, 0)
}

func TestReadGameInfoOneLine(t *testing.T) {
	infoJSON := gameInfoTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(infoJSON),
	})
	game, err := ReadGameInfo("myid")

	require.NoError(t, err)
	require.Equal(t, "myid", game.ID)
	require.Equal(t, int64(10), game.Width)
	require.Equal(t, int64(12), game.Height)
}

func TestReadGameInfoManyLines(t *testing.T) {
	j := gameInfoTestJSON() + framesTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	game, err := ReadGameInfo("myid")

	require.NoError(t, err)
	require.Equal(t, "myid", game.ID)
}

func TestReadGameInfoManyLinesPlusGarbage(t *testing.T) {
	j := gameInfoTestJSON() + framesTestJSON() + "asdf"

	openFileReader = fileOpener(map[string]string{
		"myid": string(j),
	})
	game, err := ReadGameInfo("myid")

	require.NoError(t, err, "garbage data after last frame should not break anything")
	require.Equal(t, "myid", game.ID)
}

func TestReadGameInfoPlusGarbage(t *testing.T) {
	text := gameInfoTestJSON() + "\nfoo\nbar"

	openFileReader = fileOpener(map[string]string{
		"myid": string(text),
	})
	game, err := ReadGameInfo("myid")

	require.NoError(t, err, "garbage data after header should not break anything")
	require.Equal(t, "myid", game.ID)
}

func TestReadGameInfoCorruptJSON(t *testing.T) {
	text := "foo\nbar"

	openFileReader = fileOpener(map[string]string{
		"myid": string(text),
	})
	_, err := ReadGameInfo("myid")

	require.NotNil(t, err)
}

func TestReadGameInfoMissingFile(t *testing.T) {
	infoJSON := gameInfoTestJSON()

	openFileReader = fileOpener(map[string]string{
		"myid": string(infoJSON),
	})
	_, err := ReadGameInfo("wrongid")

	require.NotNil(t, err)
}
