package worker

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/battlesnakeio/engine/rules"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
)

var snakeURL string

func init() {
	tst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"move": "up"}`))
	}))
	snakeURL = tst.URL
}

func TestWorker_RunnerErrors(t *testing.T) {
	client, store := server()

	ctx := context.Background()

	w := &Worker{
		ControllerClient: client,
		PollInterval:     1 * time.Millisecond,
		RunGame:          Runner,
	}

	t.Run("NoGame", func(t *testing.T) {
		err := Runner(ctx, client, "")
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = NotFound desc = controller: game not found",
			err.Error(),
		)
	})

	t.Run("GameTickError", func(t *testing.T) {
		store.CreateGame(ctx, &pb.Game{ID: "1", Status: rules.GameStatusRunning}, nil)

		err := w.run(ctx, 1)
		require.NotNil(t, err)
		require.Equal(t, "rules: invalid state, previous tick is nil", err.Error())
	})

	t.Run("GameTickLocked", func(t *testing.T) {
		store.CreateGame(ctx,
			&pb.Game{ID: "2", Status: rules.GameStatusRunning},
			[]*pb.GameTick{{}},
		)
		w.RunGame = func(c context.Context, cl pb.ControllerClient, id string) error {
			md, _ := metadata.FromOutgoingContext(c)
			store.Unlock(c, id, md[pb.TokenKey][0])
			store.Lock(c, id, "")

			return Runner(c, cl, id)
		}

		err := w.run(ctx, 1)
		require.NotNil(t, err)
		require.Equal(t,
			"rpc error: code = ResourceExhausted desc = controller: game is locked",
			err.Error(),
		)
	})

}

func TestWorker_Runner(t *testing.T) {
	t.Skip("skipping due to race conditions")

	client, _ := server()

	games := map[string]*pb.CreateRequest{
		"Simple": {
			Width:  5,
			Height: 5,
			Food:   5,
			Snakes: []*pb.SnakeOptions{
				{
					Name: "1",
					URL:  snakeURL,
					ID:   "1",
				},
				{
					Name: "2",
					URL:  snakeURL,
					ID:   "2",
				},
			},
		},
		"InvalidURL": {
			Width:  5,
			Height: 5,
			Food:   5,
			Snakes: []*pb.SnakeOptions{
				{
					Name: "1",
					URL:  snakeURL,
					ID:   "1",
				},
				{
					Name: "2",
					URL:  "invalid",
					ID:   "2",
				},
			},
		},
		"LargerBoard": {
			Width:  25,
			Height: 25,
			Food:   10,
			Snakes: []*pb.SnakeOptions{
				{
					Name: "1",
					URL:  snakeURL,
					ID:   "1",
				},
				{
					Name: "2",
					URL:  snakeURL,
					ID:   "2",
				},
			},
		},
	}
	ctx := context.Background()

	w := &Worker{
		ControllerClient: client,
		PollInterval:     1 * time.Millisecond,
		RunGame:          Runner,
	}

	for key, game := range games {
		t.Run(fmt.Sprintf("%s", key), func(t *testing.T) {
			g, err := client.Create(ctx, game)
			require.Nil(t, err)

			_, err = client.Start(ctx, &pb.StartRequest{ID: g.ID})
			require.Nil(t, err)

			err = w.run(ctx, 1)
			require.Nil(t, err)

			st, err := client.Status(ctx, &pb.StatusRequest{ID: g.ID})
			require.Nil(t, err)

			spew.Dump(st)
		})
	}
}
