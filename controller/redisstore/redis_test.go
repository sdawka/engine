package redisstore

import (
	"context"
	"testing"

	"github.com/battlesnakeio/engine/controller"
	"github.com/battlesnakeio/engine/controller/pb"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestLock(t *testing.T) {
	store := storeOrFail(t)
	gameKey := uuid.NewV4().String()

	// No previous lock
	tkn, err := store.Lock(context.Background(), gameKey, "")
	assert.NoError(t, err, "this should be a new lock")
	assert.NotZero(t, tkn, "expect a reasonable token string back")

	// Same game (locked), no token
	_, err = store.Lock(context.Background(), gameKey, "")
	assert.Error(t, err, "we need a token to get this lock now")
	assert.Equal(t, controller.ErrIsLocked, err, "specifically this error")

	// Same game (locked, but with token)
	tkn, err = store.Lock(context.Background(), gameKey, tkn)
	assert.NoError(t, err, "the lock should be allowed using previous token")
	assert.NotNil(t, tkn, "should still get a reasonable token back")
}

// Unlock will unlock a game if it is locked and the token used to lock it
// is correct.
func TestUnlock(t *testing.T) {
	store := storeOrFail(t)
	gameKey := uuid.NewV4().String()

	// No previous lock
	tkn, err := store.Lock(context.Background(), gameKey, "")
	assert.NoError(t, err, "this should be a new lock")
	assert.NotZero(t, tkn, "expect a reasonable token string back")

	store.Unlock(context.Background(), gameKey, tkn)

	// No previous lock again (unlocked)
	tkn, err = store.Lock(context.Background(), gameKey, "")
	assert.NoError(t, err, "this should be a new lock")
	assert.NotZero(t, tkn, "expect a reasonable token string back")
}

// PopGameID returns a new game that is unlocked and running. Workers call
// this method through the controller to find games to process.
func TestPopGameID(t *testing.T) {
	store := storeOrFail(t)

	// Empty state
	gameID, err := store.PopGameID(context.Background())
	assert.NoError(t, err, "no error for empty games")
	assert.Zero(t, gameID, "no game should be returned when empty")

	// Add a game
	game := &pb.Game{
		ID: uuid.NewV4().String(),
	}
	err = store.CreateGame(context.Background(), game, nil)
	assert.NoError(t, err, "no error for creating games")

	// Pop our game out
	poppedID, err := store.PopGameID(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, gameID, poppedID, "1 game in store, should pop that one")

	// Lock it
	_, err = store.Lock(context.Background(), game.ID, "")
	assert.NoError(t, err)

	// no unlocked games left
	gameID, err = store.PopGameID(context.Background())
	assert.NoError(t, err, "no error for empty unlocked games")
	assert.Zero(t, gameID, "no game should be returned when empty unlocked games")
}

// SetGameStatus is used to set a specific game status. This operation
// should be atomic.
func TestSetGameStatus(t *testing.T) {
	store := storeOrFail(t)

	// Add a game
	game := &pb.Game{
		ID:     uuid.NewV4().String(),
		Status: "old",
	}
	err := store.CreateGame(context.Background(), game, nil)
	assert.NoError(t, err, "no error for creating games")

	// Set a status
	status := "TEST STATUS"
	err = store.SetGameStatus(context.Background(), game.ID, status)
	assert.NoError(t, err)

	// Validate new status is present
	game, _ = store.GetGame(context.Background(), game.ID)
	assert.Equal(t, status, game.GetStatus())
}

// Test Create/Get games
func TestCreateGame(t *testing.T) {
	store := storeOrFail(t)

	// Iterate over each game case and ensure they persist correctly
	for _, gameCase := range gameCases {
		err := store.CreateGame(context.Background(), gameCase.game, gameCase.frames)
		assert.NoError(t, err, "no error for creating games")

		// Validate games returned are the same as what we saved
		game, err := store.GetGame(context.Background(), gameCase.game.ID)
		assert.NoError(t, err, "all games should have created and be retrievable")
		assert.Equal(t, gameCase.game, game)
	}
}

// Tests PushGameFrame and ListGameFrames
func TestPushGameFrame(t *testing.T) {
	store := storeOrFail(t)
	game := &pb.Game{ID: uuid.NewV4().String()}
	err := store.CreateGame(context.Background(), game, nil)
	assert.NoError(t, err)

	// No frames
	frames, err := store.ListGameFrames(context.Background(), game.ID, 10, 0)
	assert.NoError(t, err)
	assert.Zero(t, frames, "no frames yet")

	// 1 frame
	err = store.PushGameFrame(context.Background(), game.ID, testFrames[0])
	assert.NoError(t, err)
	frames, err = store.ListGameFrames(context.Background(), game.ID, 10, 0)
	assert.NoError(t, err)
	assert.Contains(t, testFrames[0], frames)
	assert.Equal(t, 1, len(frames), "only 1 frame should be present")

	// remaining frames
	err = store.PushGameFrame(context.Background(), game.ID, testFrames[1])
	assert.NoError(t, err)
	err = store.PushGameFrame(context.Background(), game.ID, testFrames[2])
	assert.NoError(t, err)
	frames, err = store.ListGameFrames(context.Background(), game.ID, 10, 0)
	assert.NoError(t, err)
	assert.ElementsMatch(t, frames, testFrames)

	// smaller limit
	frames, err = store.ListGameFrames(context.Background(), game.ID, 2, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(frames))

	// offset
	frames, err = store.ListGameFrames(context.Background(), game.ID, 2, 1)
	assert.NoError(t, err)
	for _, f := range []*pb.GameFrame{testFrames[1], testFrames[2]} {
		assert.Contains(t, frames, f, "the frames should match the test frames, so offset by 1 and limit 2 should mean 2nd and 3rd frames")
	}

	// negative offset
	frames, err = store.ListGameFrames(context.Background(), game.ID, 1, -1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(frames), "should only be 1 game")
	assert.Equal(t, testFrames[3], frames[0])

	// bigger limit
	_, err = store.ListGameFrames(context.Background(), game.ID, 1000000000, 0)
	assert.NoError(t, err)

	// No such game
	_, err = store.ListGameFrames(context.Background(), game.ID, 10, 0)
	assert.Error(t, err)
	assert.Equal(t, controller.ErrNotFound, err)
}

func storeOrFail(t *testing.T) *RedisStore {
	store, err := NewRedisStore("127.0.0.1:6379")
	assert.NoError(t, err)
	return store
}

type testCase struct {
	game   *pb.Game
	frames []*pb.GameFrame
}

// 🐍🐍🐍 got some tough UTF stuff going on here :D 🐍🐍🐍
var gameCases = []testCase{
	{
		game: &pb.Game{ID: uuid.NewV4().String()},
	},
	{
		game: &pb.Game{ID: uuid.NewV4().String(), Status: "Test status", Width: 10, Height: 10, SnakeTimeout: 10, TurnTimeout: 10, Mode: "Test"},
		frames: []*pb.GameFrame{
			{
				Turn: 0, Food: []*pb.Point{{X: 0, Y: 0}},
				Snakes: []*pb.Snake{
					{ID: uuid.NewV4().String(), URL: "http://example.com/snek", Health: 50, Color: "red"},
				},
			},
		},
	},
	{
		game: &pb.Game{ID: uuid.NewV4().String(), Status: "ᚠᛇᚻ᛫ᛒᛦᚦ᛫ᚠᚱᚩᚠᚢᚱ᛫ᚠᛁᚱᚪ᛫ᚷᛖᚻᚹᛦᛚᚳᚢᛗ", Width: 10, Height: 10, SnakeTimeout: 10, TurnTimeout: 10, Mode: "Test"},
	},
	{
		game: &pb.Game{ID: uuid.NewV4().String(), Status: "Test status", Width: 10, Height: 10, SnakeTimeout: 10, TurnTimeout: 10},
	},
	{
		game:   &pb.Game{ID: uuid.NewV4().String(), Status: "Test status", Width: 10, Height: 10, Mode: "🐍"},
		frames: testFrames,
	},
	{
		game: &pb.Game{ID: uuid.NewV4().String(), Status: "Snek 🐍🐍🐍🐍🐍", SnakeTimeout: 100, Mode: "🐍🐍🐍🐍🐍"},
	},
}

var testFrames = []*pb.GameFrame{
	{
		Turn: 0, Food: []*pb.Point{{X: 10, Y: 10}},
		Snakes: []*pb.Snake{
			{ID: uuid.NewV4().String(), URL: "http://example.com/snek1", Health: 26, Color: "blue"},
			{ID: uuid.NewV4().String(), URL: "http://example.com/snek2", Health: 33, Color: "red"},
		},
	},
	{
		Turn: 1, Food: []*pb.Point{{X: 1, Y: 0}},
		Snakes: []*pb.Snake{
			{ID: uuid.NewV4().String(), URL: "http://example.com/snek", Health: 1, Color: "orange"},
		},
	},
	{
		Turn: 2, Food: []*pb.Point{{X: 2, Y: 22}},
		Snakes: []*pb.Snake{
			{ID: uuid.NewV4().String(), URL: "http://example.com/snek", Health: 10, Color: "green"},
		},
	},
}

/*
ID     string   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Name   string   `protobuf:"bytes,2,opt,name=Name,proto3" json:"Name,omitempty"`
	URL    string   `protobuf:"bytes,3,opt,name=URL,proto3" json:"URL,omitempty"`
	Body   []*Point `protobuf:"bytes,4,rep,name=Body" json:"Body,omitempty"`
	Health int64    `protobuf:"varint,5,opt,name=Health,proto3" json:"Health,omitempty"`
	Death  *Death   `protobuf:"bytes,6,opt,name=Death" json:"Death,omitempty"`
	Color  string   `protobuf:"bytes,7,opt,name=Color,proto3" json:"Color,omitempty"`
*/
