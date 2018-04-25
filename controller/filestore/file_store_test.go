package filestore

import (
	"context"
	"testing"

	"github.com/battlesnakeio/engine/rules"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

func TestFileStore(t *testing.T) {
	w := &mockWriter{
		closed: false,
	}
	openFileWriter = func(id string) (writer, error) {
		return w, nil
	}
	fs := NewFileStore()
	ticks := []*pb.GameTick{basicTicks[0]}
	err := fs.CreateGame(context.Background(), basicGame, ticks)
	require.NoError(t, err)

	fs.SetGameStatus(context.Background(), "myid", rules.GameStatusComplete)
	require.True(t, w.clossed)
}
