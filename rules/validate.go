package rules

import (
	"bytes"
	"encoding/json"
	"github.com/battlesnakeio/engine/controller/pb"
	"io/ioutil"
	"strings"
	"time"
)

func ValidateStart(gameID string, url string) *pb.SnakeResponseStatus {
	response := scoreResponse(gameID, url, "/start")
	return response
}

func ValidateMove(gameID string, url string) *pb.SnakeResponseStatus {
	response := scoreResponse(gameID, url, "/move")
	return response
}

func ValidateEnd(gameID string, url string) *pb.SnakeResponseStatus {
	response := scoreResponse(gameID, url, "/end")
	return response
}

func scoreResponse(gameID string, url string, endpoint string) *pb.SnakeResponseStatus {
	game, frame := createGameFrame(gameID, url)
	response := &pb.SnakeResponseStatus{
		Score: &pb.Score{},
	}
	if !isValidURL(url) {
		response.Score.ChecksFailed++
		response.Message = "Snake URL not valid"
		response.Errors = []string{"invalid url '" + url + "'"}
	} else {
		response.Score.ChecksPassed++
		rawResponse, responseCode, responseTime, responseError := makeSnakeCall(game, frame, url, endpoint)
		response.Message = "Perfect"
		response.RawResponse = rawResponse
		response.ResponseTime = responseTime
		response.ResponseCode = int32(responseCode)
		if responseError != nil {
			response.Score.ChecksFailed++
			if strings.HasPrefix(responseError.Error(), "invalid") {
				response.Message = "Bad response format - please ensure you return valid JSON"
				response.Errors = append(response.Errors, responseError.Error())
			} else {
				response.Message = "Unknown error"
				response.Errors = append(response.Errors, responseError.Error())
			}
		} else {
			response.Score.ChecksPassed++
		}
		if responseTime < 1000 {
			response.Score.ChecksPassed++
		} else {
			response.Message = "Slow snake"
			response.Errors = append(response.Errors, "snake took "+string(responseTime)+" ms to respond, try and get it < 1000 ms.")
			response.Score.ChecksFailed++
		}

	}
	return response
}

func makeSnakeCall(game *pb.Game, frame *pb.GameFrame, url string, endpoint string) (string, int, int32, error) {
	netClient := createClient(5000 * time.Millisecond)
	req := buildSnakeRequest(game, frame, "you")

	data, _ := json.Marshal(req)
	buf := bytes.NewBuffer(data)
	start := time.Now().UnixNano()
	response, err := netClient.Post(url+endpoint, "application/json", buf)
	if err != nil {
		return "", 0, 0, err
	}
	responseCode := response.StatusCode
	finish := time.Now().UnixNano()
	responseTime := int32((finish - start) / int64(time.Millisecond))
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", responseCode, 0, err
	}

	rawResponse := ""

	rawResponse = string(contents[:])
	if endpoint != "/end" {
		var raw map[string]interface{}
		err = json.Unmarshal(contents, &raw)
	}

	return rawResponse, responseCode, responseTime, err
}

func createGameFrame(gameID string, url string) (*pb.Game, *pb.GameFrame) {
	game := &pb.Game{
		Height: 10,
		Width:  10,
		ID:     gameID,
		Status: "Running",
	}
	snake := &pb.Snake{
		ID:   "you",
		URL:  url,
		Name: "you",
		Body: []*pb.Point{
			&pb.Point{X: 2, Y: 2},
		},
	}
	frame := &pb.GameFrame{
		Snakes: []*pb.Snake{
			snake,
		},
		Turn: 1,
		Food: []*pb.Point{&pb.Point{
			X: 1,
			Y: 1,
		},
		},
	}
	return game, frame
}
