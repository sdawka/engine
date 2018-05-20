package redisstore

import (
	"context"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/go-redis/redis"
	"github.com/sendwithus/errors"
)

type RedisStore struct {
	client *redis.Client
}

// NewRedisStore will create a new instance of an underlying redis client, so it should not be re-created across "threads"
// - connectURL see: github.com/go-redis/redis/options.go for URL specifics
// The underlying redis client will be immediately tested for connectivity, so don't call this until you know redis can connect.
// Returns a new instance OR an error if unable (meaning an issue connecting to your redis URL)
func NewRedisStore(connectURL string) (*RedisStore, error) {
	o, err := redis.ParseURL(connectURL)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse redis URL")
	}

	client := redis.NewClient(o)

	// Validate it's connected
	err = client.Ping().Err()
	if err != nil {
		return nil, errors.Wrap(err, "unable to connect ")
	}

	return &RedisStore{client: client}, nil
}

// Lock will lock a specific game, returning a token that must be used to
// write frames to the game.
func (rs *RedisStore) Lock(ctx context.Context, key, token string) (string, error) {
	return "", nil
}

// Unlock will unlock a game if it is locked and the token used to lock it
// is correct.
func (rs *RedisStore) Unlock(ctx context.Context, key, token string) error {
	return nil
}

// PopGameID returns a new game that is unlocked and running. Workers call
// this method through the controller to find games to process.
func (rs *RedisStore) PopGameID(context.Context) (string, error) {
	return "", nil
}

// SetGameStatus is used to set a specific game status. This operation
// should be atomic.
func (rs *RedisStore) SetGameStatus(c context.Context, id, status string) error {
	return nil
}

// CreateGame will insert a game with the default game frames.
func (rs *RedisStore) CreateGame(context.Context, *pb.Game, []*pb.GameFrame) error {
	return nil
}

// PushGameFrame will push a game frame onto the list of frames.
func (rs *RedisStore) PushGameFrame(c context.Context, id string, t *pb.GameFrame) error {
	return nil
}

// ListGameFrames will list frames by an offset and limit, it supports
// negative offset.
func (rs *RedisStore) ListGameFrames(c context.Context, id string, limit, offset int) ([]*pb.GameFrame, error) {
	return nil, nil
}

// GetGame will fetch the game.
func (rs *RedisStore) GetGame(context.Context, string) (*pb.Game, error) {
	return nil, nil
}
