package rules

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	nu "net/url"
	"sync"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	log "github.com/sirupsen/logrus"
)

var netClient httpClient = &wrappedHTTPClient{
	Client: &http.Client{
		Timeout: 2 * time.Millisecond,
	},
}

// SnakeUpdate bundles together a snake with a move for processing
type SnakeUpdate struct {
	Snake *pb.Snake
	Move  string
	Err   error
}

func isValidURL(url string) bool {
	if len(url) == 0 {
		return false
	}

	parsed, err := nu.Parse(url)
	if err != nil {
		return false
	}

	if len(parsed.Scheme) == 0 {
		return false
	}

	return true
}

// GatherSnakeMoves goes and queries each snake for the snake move
func GatherSnakeMoves(timeout time.Duration, gameTick *pb.GameTick) <-chan SnakeUpdate {
	respChan := make(chan SnakeUpdate, len(gameTick.Snakes))
	go func() {
		wg := sync.WaitGroup{}
		for _, s := range gameTick.AliveSnakes() {
			if !isValidURL(s.URL) {
				respChan <- SnakeUpdate{
					Snake: s,
					Err:   errors.New("invalid snake url"),
				}
				continue
			}
			wg.Add(1)
			go func(snake *pb.Snake) {
				getMove(snake, timeout, respChan)
				wg.Done()
			}(s)
		}
		wg.Wait()
		close(respChan)
	}()
	return respChan
}

// GetMove queries the snake url and returns the resp on the channel
func getMove(snake *pb.Snake, timeout time.Duration, resp chan<- SnakeUpdate) {
	response, err := netClient.Get(snake.URL)
	if err != nil {
		log.WithError(err).Errorf("error while querying %s for move", snake.ID)
		resp <- SnakeUpdate{
			Snake: snake,
			Err:   err,
		}
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.WithError(err).Errorf("error while decoding response body for %s", snake.ID)
		resp <- SnakeUpdate{
			Snake: snake,
			Err:   err,
		}
	}

	mr := &MoveResponse{}
	err = json.Unmarshal(data, mr)
	if err != nil {
		log.WithError(err).Errorf("error while converting response body to json for %s", snake.ID)
		resp <- SnakeUpdate{
			Snake: snake,
			Err:   err,
		}
	}
	resp <- SnakeUpdate{
		Snake: snake,
		Move:  mr.Move,
	}
}
