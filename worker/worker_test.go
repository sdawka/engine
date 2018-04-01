package worker

import (
	"context"
	"testing"
	"time"

	"github.com/battlesnakeio/engine/controller"
	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
)

var client pb.ControllerClient

func init() {
	ctrl := controller.New(controller.InMemStore())
	go func() {
		if err := ctrl.Serve(":0"); err != nil {
			panic(err)
		}
	}()
	var err error
	client, err = pb.Dial(ctrl.DialAddress())
	if err != nil {
		panic(err)
	}
}

func TestWorker_RunNoGame(t *testing.T) {
	w := &Worker{
		ControllerClient:  client,
		PollInterval:      200 * time.Millisecond,
		HeartbeatInterval: 200 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := w.run(ctx, 1)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "not found")
}

func TestWorker_Run(t *testing.T) {
	w := &Worker{
		ControllerClient:  client,
		PollInterval:      1 * time.Millisecond,
		HeartbeatInterval: 1 * time.Millisecond,
	}

	resp, _ := client.Create(context.Background(), &pb.CreateRequest{})
	client.Start(context.Background(), &pb.StartRequest{
		ID: resp.ID,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := w.run(ctx, 1)
	require.Equal(t, context.DeadlineExceeded, err)
}

func TestWorker_RunLoop(t *testing.T) {
	w := &Worker{
		ControllerClient:  client,
		PollInterval:      1 * time.Millisecond,
		HeartbeatInterval: 1 * time.Millisecond,
	}

	resp, _ := client.Create(context.Background(), &pb.CreateRequest{})
	client.Start(context.Background(), &pb.StartRequest{
		ID: resp.ID,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	w.Run(ctx, 1)
}
