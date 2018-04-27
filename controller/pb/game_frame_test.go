package pb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGameFrameAliveSnakes(t *testing.T) {
	gt := &GameFrame{
		Snakes: []*Snake{
			&Snake{},
			&Snake{
				Death: &Death{},
			},
		},
	}
	snakes := gt.AliveSnakes()
	require.Len(t, snakes, 1)
}

func TestGameFrameDeadSnakes(t *testing.T) {
	gt := &GameFrame{
		Snakes: []*Snake{
			&Snake{},
			&Snake{
				Death: &Death{},
			},
		},
	}
	snakes := gt.DeadSnakes()
	require.Len(t, snakes, 1)
}
