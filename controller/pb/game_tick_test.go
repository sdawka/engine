package pb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGameTickAliveSnakes(t *testing.T) {
	gt := &GameTick{
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

func TestGameTickDeadSnakes(t *testing.T) {
	gt := &GameTick{
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
