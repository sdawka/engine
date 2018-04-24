package csv

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/battlesnakeio/engine/controller"
	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/battlesnakeio/engine/rules"
	"github.com/gogo/protobuf/proto"
	uuid "github.com/satori/go.uuid"
)

/*
CSV file will look like this except json will all be on one line:

#{
	board: {
		id: "1234",
		width: 10,
		height: 10
	},
	snakes: [
		{
			id: "123",
			name: "123",
			color: "#ff0000",
			start: { x: 2, y: 3 }
		},
		{
			id: "246",
			name: "asdf",
			color: "#ffff00",
			start: { x: 4, y: 3 }
		}
	]
}
turn,player1,player2
1,l,r
2,u,r
3,l,d
4,l,d
...
*/

// NewCSVStore returns a CSV file based store implementation (1 file per game).
func NewCSVStore() controller.Store {
	return &csvStore{
		games: map[string]*pb.Game{},
		ticks: map[string][]*pb.GameTick{},
		locks: map[string]*lock{},
	}
}

type lock struct {
	token   string
	expires time.Time
}

type csvStore struct {
	games   map[string]*pb.Game
	ticks   map[string][]*pb.GameTick
	handles map[string]*os.File
	locks   map[string]*lock
	lock    sync.Mutex
}

// closeGame removes the game from in-memory cache and closes the handle to its
// file. Should be called when game is complete.
func (cs *csvStore) closeGame(id string) {
	cs.handles[id].Close()
	delete(cs.games, id)
	delete(cs.ticks, id)
	delete(cs.handles, id)
}

func (cs *csvStore) Lock(ctx context.Context, key, token string) (string, error) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	now := time.Now()

	l, ok := cs.locks[key]
	if ok {
		// We have a lock token, if it's expired just delete it and continue as
		// if nothing happened.
		if l.expires.Before(now) {
			delete(cs.locks, key)
		} else {
			// If the token is not expired and matched our active token, let's
			// just bump the expiration.
			if l.token == token {
				l.expires = time.Now().Add(controller.LockExpiry)
				return l.token, nil
			}
			// If it's not our token, we should throw an error.
			return "", controller.ErrIsLocked
		}
	}
	if token == "" {
		token = uuid.NewV4().String()
	}
	// Lock was expired or non-existant, create a new token.
	l = &lock{
		token:   token,
		expires: now.Add(controller.LockExpiry),
	}
	cs.locks[key] = l
	return l.token, nil
}

func (cs *csvStore) isLocked(key string) bool {
	l, ok := cs.locks[key]
	return ok && l.expires.After(time.Now())
}

func (cs *csvStore) Unlock(ctx context.Context, key, token string) error {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	now := time.Now()

	l, ok := cs.locks[key]
	// No lock? Don't care.
	if !ok {
		return nil
	}
	// We have a lock that matches our token, even if it's expired we are safe
	// to remove it. If it's expired, remove it as well.
	if l.expires.Before(now) || l.token == token {
		delete(cs.locks, key)
		return nil
	}
	// The token is valid and doesn't match our lock.
	return controller.ErrIsLocked
}

// PopGameID gives the next running game. Since running games should always be
// cached in memory it is not necessary to scan file system.
func (cs *csvStore) PopGameID(ctx context.Context) (string, error) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	// For every game we need to make sure it's active and is not locked before
	// returning it. We get randomness due to go's built in random map.
	for id, g := range cs.games {
		if !cs.isLocked(id) && g.Status == rules.GameStatusRunning {
			return id, nil
		}
	}
	return "", controller.ErrNotFound
}

func (cs *csvStore) CreateGame(ctx context.Context, g *pb.Game, ticks []*pb.GameTick) error {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	handle, err := appendOnlyFileHandle(g.ID)
	if err != nil {
		return err
	}
	cs.handles[g.ID] = handle
	cs.games[g.ID] = g
	cs.ticks[g.ID] = ticks
	return nil
}

func (cs *csvStore) SetGameStatus(ctx context.Context, id, status string) error {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	if g, ok := cs.games[id]; ok {
		g.Status = status
		if status != rules.GameStatusRunning {
			cs.closeGame(id)
		}
		return nil
	}
	return controller.ErrNotFound
}

func (cs *csvStore) PushGameTick(ctx context.Context, id string, g *pb.GameTick) error {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	// Add tick to in-memory cache
	cs.ticks[id] = append(cs.ticks[id], g)

	// Add tick to archive file
	return writeTick(cs.handles[id], g)
}

func (cs *csvStore) ListGameTicks(ctx context.Context, id string, limit, offset int) ([]*pb.GameTick, error) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	if _, ok := cs.games[id]; !ok {
		return nil, controller.ErrNotFound
	}
	ticks := cs.ticks[id]
	if len(ticks) == 0 {
		return nil, nil
	}
	if offset < 0 {
		offset = len(ticks) + offset
	}
	if offset >= len(ticks) {
		return nil, nil
	}
	if offset+limit >= len(ticks) {
		limit = len(ticks) - offset
	}
	return ticks[offset : offset+limit], nil
}

func (cs *csvStore) GetGame(ctx context.Context, id string) (*pb.Game, error) {
	cs.lock.Lock()
	defer cs.lock.Unlock()

	if g, ok := cs.games[id]; ok {
		// Clone the game, since this could be modified after this is returned
		// and upset internal state inside the store.
		clone := proto.Clone(g).(*pb.Game)
		return clone, nil
	}
	return nil, controller.ErrNotFound
}

type gameArchive struct {
	Board  boardArchive   `json:"board"`
	Snakes []snakeArchive `json:"snakes"`
}

type boardArchive struct {
	ID     string `json:"id"`
	Width  int64  `json:"width"`
	Height int64  `json:"height"`
}

type pointArchive struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type snakeArchive struct {
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Color string       `json:"color"`
	Start pointArchive `json:"start"`
}

func getFilePath(id string) string {
	return "/home/graeme/.battlesnake/games/" + id + ".csv"
}
