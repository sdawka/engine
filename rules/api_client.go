package rules

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"sync"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	log "github.com/sirupsen/logrus"
)

type snakeResponse struct {
	snake *pb.Snake
	data  []byte
	err   error
}

func gatherSnakeResponses(url string, timeout time.Duration, game *pb.Game, startState *pb.GameTick) []snakeResponse {
	respChan := make(chan snakeResponse, len(startState.Snakes))
	wg := sync.WaitGroup{}

	for _, snake := range startState.Snakes {
		if !isValidURL(snake.URL) {
			respChan <- snakeResponse{
				snake: snake,
				err:   errors.New("invalid snake URL: " + snake.URL),
			}
			continue
		}

		wg.Add(1)
		go func(s *pb.Snake) {
			getSnakeResponse(url, s, game, startState, timeout, respChan)
			wg.Done()
		}(snake)
	}

	wg.Wait()
	close(respChan)

	ret := []snakeResponse{}
	for response := range respChan {
		ret = append(ret, response)
	}
	return ret
}

func postToSnakeServer(url string, reqData []byte, s *pb.Snake, timeout time.Duration, resp chan<- snakeResponse) {
	buf := bytes.NewBuffer(reqData)
	netClient := createClient(timeout)
	postResponse, err := netClient.Post(getURL(s.URL, url), "application/json", buf)
	if err != nil {
		log.WithError(err).
			Errorf("error POSTing to %s for snake %s", s.ID)
		resp <- snakeResponse{
			snake: s,
			err:   err,
		}
	}

	responseData, err := ioutil.ReadAll(postResponse.Body)
	resp <- snakeResponse{
		snake: s,
		data:  responseData,
		err:   err,
	}
}

func getSnakeResponse(url string, s *pb.Snake, game *pb.Game, frame *pb.GameTick, timeout time.Duration, resp chan<- snakeResponse) {
	req := buildSnakeRequest(game, frame, s.ID)
	data, err := json.Marshal(req)

	if err != nil {
		log.WithError(err).
			Errorf("error while marshaling snake request: %s", s.ID)
		resp <- snakeResponse{
			snake: s,
			err:   err,
		}
		return
	}

	postToSnakeServer(url, data, s, timeout, resp)
}
