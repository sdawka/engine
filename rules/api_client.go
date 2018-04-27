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

type multiSnakeRequest struct {
	url     string
	timeout time.Duration
	game    *pb.Game
	frame   *pb.GameFrame
}

type snakePostOptions struct {
	url     string
	snake   *pb.Snake
	timeout time.Duration
}

type snakePostRequest struct {
	options snakePostOptions
	data    []byte
}

func gatherAllSnakeResponses(multiReq multiSnakeRequest) []snakeResponse {
	return gatherSnakeResponses(multiReq, multiReq.frame.Snakes)
}

func gatherAliveSnakeResponses(multiReq multiSnakeRequest) []snakeResponse {
	return gatherSnakeResponses(multiReq, multiReq.frame.AliveSnakes())
}

func gatherSnakeResponses(multiReq multiSnakeRequest, snakes []*pb.Snake) []snakeResponse {
	respChan := make(chan snakeResponse, len(multiReq.frame.Snakes))
	wg := sync.WaitGroup{}

	for _, snake := range snakes {
		if !isValidURL(snake.URL) {
			respChan <- snakeResponse{
				snake: snake,
				err:   errors.New("invalid snake URL: " + snake.URL),
			}
			continue
		}

		wg.Add(1)
		go func(s *pb.Snake) {
			options := snakePostOptions{
				url:     multiReq.url,
				snake:   s,
				timeout: multiReq.timeout,
			}
			getSnakeResponse(options, multiReq.game, multiReq.frame, respChan)
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

func postToSnakeServer(req snakePostRequest, resp chan<- snakeResponse) {
	buf := bytes.NewBuffer(req.data)
	netClient := createClient(req.options.timeout)
	postURL := getURL(req.options.snake.URL, req.options.url)
	postResponse, err := netClient.Post(postURL, "application/json", buf)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"url": postURL,
			"id":  req.options.snake.ID,
		}).Error("error POSTing to snake")
		resp <- snakeResponse{
			snake: req.options.snake,
			err:   err,
		}
		return
	}

	responseData, err := ioutil.ReadAll(postResponse.Body)
	resp <- snakeResponse{
		snake: req.options.snake,
		data:  responseData,
		err:   err,
	}
}

func getSnakeResponse(options snakePostOptions, game *pb.Game, frame *pb.GameFrame, resp chan<- snakeResponse) {
	req := buildSnakeRequest(game, frame, options.snake.ID)
	data, err := json.Marshal(req)

	if err != nil {
		log.WithError(err).WithField("snakeID", options.snake.ID).
			Error("error while marshaling snake request")
		resp <- snakeResponse{
			snake: options.snake,
			err:   err,
		}
		return
	}

	postToSnakeServer(snakePostRequest{
		options: options,
		data:    data,
	}, resp)
}
