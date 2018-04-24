package filestore

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

// NewFileStore returns a CSV file based store implementation (1 file per game).
func NewFileStore() controller.Store {
	return &fileStore{
		games: map[string]*pb.Game{},
		ticks: map[string][]*pb.GameTick{},
		locks: map[string]*lock{},
	}
}

type lock struct {
	token   string
	expires time.Time
}

type fileStore struct {
	games   map[string]*pb.Game
	ticks   map[string][]*pb.GameTick
	handles map[string]*os.File
	locks   map[string]*lock
	lock    sync.Mutex
}

// closeGame removes the game from in-memory cache and closes the handle to its
// file. Should be called when game is complete.
func (fs *fileStore) closeGame(id string) {
	fs.handles[id].Close()
	delete(fs.games, id)
	delete(fs.ticks, id)
	delete(fs.handles, id)
}

func (fs *fileStore) Lock(ctx context.Context, key, token string) (string, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	now := time.Now()

	l, ok := fs.locks[key]
	if ok {
		// We have a lock token, if it's expired just delete it and continue as
		// if nothing happened.
		if l.expires.Before(now) {
			delete(fs.locks, key)
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
	fs.locks[key] = l
	return l.token, nil
}

func (fs *fileStore) isLocked(key string) bool {
	l, ok := fs.locks[key]
	return ok && l.expires.After(time.Now())
}

func (fs *fileStore) Unlock(ctx context.Context, key, token string) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	now := time.Now()

	l, ok := fs.locks[key]
	// No lock? Don't care.
	if !ok {
		return nil
	}
	// We have a lock that matches our token, even if it's expired we are safe
	// to remove it. If it's expired, remove it as well.
	if l.expires.Before(now) || l.token == token {
		delete(fs.locks, key)
		return nil
	}
	// The token is valid and doesn't match our lock.
	return controller.ErrIsLocked
}

// PopGameID gives the next running game. Since running games should always be
// cached in memory it is not necessary to scan file system.
func (fs *fileStore) PopGameID(ctx context.Context) (string, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	// For every game we need to make sure it's active and is not locked before
	// returning it. We get randomness due to go's built in random map.
	for id, g := range fs.games {
		if !fs.isLocked(id) && g.Status == rules.GameStatusRunning {
			return id, nil
		}
	}
	return "", controller.ErrNotFound
}

func (fs *fileStore) CreateGame(ctx context.Context, g *pb.Game, ticks []*pb.GameTick) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	handle, err := appendOnlyFileHandle(g.ID)
	if err != nil {
		return err
	}
	fs.handles[g.ID] = handle
	fs.games[g.ID] = g
	fs.ticks[g.ID] = ticks
	return nil
}

func (fs *fileStore) SetGameStatus(ctx context.Context, id, status string) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	if g, ok := fs.games[id]; ok {
		g.Status = status
		if status != rules.GameStatusRunning {
			fs.closeGame(id)
		}
		return nil
	}
	return controller.ErrNotFound
}

func (fs *fileStore) PushGameTick(ctx context.Context, id string, g *pb.GameTick) error {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	// If this is the first tick, then first write the game info header.
	if !fs.hasAnyTicks(id) {
		if err := writeGameInfo(fs.handles[id], fs.games[id], g.Snakes); err != nil {
			return err
		}
	}

	// Add tick to in-memory cache
	fs.ticks[id] = append(fs.ticks[id], g)

	// Add tick to archive file
	return writeTick(fs.handles[id], g)
}

func (fs *fileStore) ListGameTicks(ctx context.Context, id string, limit, offset int) ([]*pb.GameTick, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()
	if _, ok := fs.games[id]; !ok {
		return nil, controller.ErrNotFound
	}
	ticks := fs.ticks[id]
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

func (fs *fileStore) GetGame(ctx context.Context, id string) (*pb.Game, error) {
	fs.lock.Lock()
	defer fs.lock.Unlock()

	if g, ok := fs.games[id]; ok {
		// Clone the game, since this could be modified after this is returned
		// and upset internal state inside the store.
		clone := proto.Clone(g).(*pb.Game)
		return clone, nil
	}
	return nil, controller.ErrNotFound
}

func (fs *fileStore) hasAnyTicks(gameID string) bool {
	ticks, ok := fs.ticks[gameID]
	return ok && len(ticks) > 0
}

type gameArchive struct {
	info   gameInfo
	frames []frame
}

type gameInfo struct {
	ID     string      `json:"id"`
	Width  int64       `json:"width"`
	Height int64       `json:"height"`
	Snakes []snakeInfo `json:"snakes"`
}

type point struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type snakeInfo struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	URL   string `json:"url"`
}

type frame struct {
	Turn   int64        `json:"turn"`
	Snakes []snakeState `json:"snakes"`
	Food   []point      `json:"food"`
}

type snakeState struct {
	ID     string  `json:"id"`
	Body   []point `json:"body"`
	Health int64   `json:"health"`
	Death  *death  `json:"dead"`
}

type death struct {
	Cause string `json:"cause"`
	Turn  int64  `json:"turn"`
}

func getFilePath(id string) string {
	return "/home/graeme/.battlesnake/games/" + id + ".bs"
}
