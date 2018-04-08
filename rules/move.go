package rules

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	nu "net/url"
	"strings"
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
func GatherSnakeMoves(timeout time.Duration, game *pb.Game, gameTick *pb.GameTick) <-chan SnakeUpdate {
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
				getMove(snake, buildSnakeRequest(game, gameTick, snake.ID), timeout, respChan)
				wg.Done()
			}(s)
		}
		wg.Wait()
		log.Info("closing snake update channel")
		close(respChan)
	}()
	return respChan
}

// GetMove queries the snake url and returns the resp on the channel
func getMove(snake *pb.Snake, req SnakeRequest, timeout time.Duration, resp chan<- SnakeUpdate) {
	netClient.SetTimeout(timeout)
	data, err := json.Marshal(req)
	if err != nil {
		log.WithError(err).Errorf("error while marshaling snake request: %s", snake.ID)
		resp <- SnakeUpdate{
			Snake: snake,
			Err:   err,
		}
		return
	}
	buf := bytes.NewBuffer(data)
	response, err := netClient.Post(getURL(snake.URL, "move"), "application/json", buf)
	if err != nil {
		log.WithError(err).Errorf("error while querying %s for move", snake.ID)
		resp <- SnakeUpdate{
			Snake: snake,
			Err:   err,
		}
		return
	}
	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.WithError(err).Errorf("error while decoding response body for %s", snake.ID)
		resp <- SnakeUpdate{
			Snake: snake,
			Err:   err,
		}
		return
	}

	mr := &MoveResponse{}
	err = json.Unmarshal(data, mr)
	if err != nil {
		log.WithError(err).WithField("body", string(data)).Errorf("error while converting response body to json for %s", snake.ID)
		resp <- SnakeUpdate{
			Snake: snake,
			Err:   err,
		}
		return
	}
	resp <- SnakeUpdate{
		Snake: snake,
		Move:  mr.Move,
	}
}

func cleanURL(url string) string {
	if !strings.HasSuffix(url, "/") {
		return fmt.Sprintf("%s/", url)
	}
	return url
}

func getURL(url, path string) string {
	u := cleanURL(url)
	return fmt.Sprintf("%s%s", u, path)
}
