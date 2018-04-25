package filestore

import (
	"context"
	"errors"
	"testing"

	"github.com/battlesnakeio/engine/controller"
	"github.com/battlesnakeio/engine/rules"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestFileStore(t *testing.T) {
	fs, w := testFileStore()
	ticks := []*pb.GameTick{basicTicks[0]}
	err := fs.CreateGame(context.Background(), basicGame, ticks)
	require.NoError(t, err)

	game, err := fs.GetGame(context.Background(), "myid")
	require.NoError(t, err)
	require.Equal(t, basicGame, game)

	err = fs.PushGameTick(context.Background(), "myid", basicTicks[1])
	require.NoError(t, err)

	newTicks, err := fs.ListGameTicks(context.Background(), "myid", 5, 0)
	require.NoError(t, err)
	require.Len(t, newTicks, 2)
	require.Equal(t, basicTicks, newTicks)

	fs.SetGameStatus(context.Background(), "myid", rules.GameStatusComplete)
	require.True(t, w.closed)
}

func TestCreateGameHandlesWriteError(t *testing.T) {
	fs, w := testFileStore()
	w.err = errors.New("fail")
	ticks := []*pb.GameTick{basicTicks[0]}
	err := fs.CreateGame(context.Background(), basicGame, ticks)
	require.NotNil(t, err)
}

func TestCreateGameHandlesOpenFileError(t *testing.T) {
	openFileWriter = func(id string) (writer, error) {
		return nil, errors.New("fail")
	}
	openFileReader = func(id string) (reader, error) {
		return nil, errors.New("fail")
	}
	fs := NewFileStore()
	ticks := []*pb.GameTick{basicTicks[0]}
	err := fs.CreateGame(context.Background(), basicGame, ticks)
	require.NotNil(t, err)
}

func TestCreateGetGameFound(t *testing.T) {
	fs, _ := testFileStore()

	_, err := fs.GetGame(context.Background(), "notfound")
	require.NotNil(t, err)
}

func TestPushGameTickInvalidGame(t *testing.T) {
	fs, _ := testFileStore()

	err := fs.PushGameTick(context.Background(), "notfound", basicTicks[1])
	require.NotNil(t, err)
}

func TestListGameTicksInvalidGame(t *testing.T) {
	fs, _ := testFileStore()

	_, err := fs.ListGameTicks(context.Background(), "notfound", 5, 0)
	require.NotNil(t, err)
}

func TestSetGameStatusInvalidGame(t *testing.T) {
	fs, _ := testFileStore()

	err := fs.SetGameStatus(context.Background(), "notfound", rules.GameStatusComplete)
	require.NotNil(t, err)
}

func TestLockUnlock(t *testing.T) {
	fs, _ := testFileStore()
	_, err := fs.Lock(context.Background(), "asdf", "")
	require.NoError(t, err)
	fs.Unlock(context.Background(), "asdf", "")
	//requireNoError(t, err)
}

func TestPopGameID(t *testing.T) {
	fs, _ := testFileStore()
	_, err := fs.PopGameID(context.Background())
	require.NotNil(t, err)
}

func testFileStore() (controller.Store, *mockWriter) {
	w := &mockWriter{
		closed: false,
	}
	openFileWriter = func(id string) (writer, error) {
		return w, nil
	}
	openFileReader = func(id string) (reader, error) {
		return newMockReader(w.text), nil
	}
	return NewFileStore(), w
}
