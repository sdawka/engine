package rules

import (
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/battlesnakeio/engine/controller/pb"
	"strings"
)

var snakeUrl = "http://good-server"

func TestValidateEnd(t *testing.T) {
	json := ""
	expected := &pb.SnakeResponseStatus{
		Message: "Perfect",
	}
	validateWithJson(t, ValidateEnd, snakeUrl+"/end", json, expected, 0, 3)
}
func TestValidateMove(t *testing.T) {
	json := "{  }"
	expected := &pb.SnakeResponseStatus{
		Message: "Perfect",
	}
	validateWithJson(t, ValidateMove, snakeUrl+"/move", json, expected, 0, 3)
}

func TestValidateStart(t *testing.T) {
	json := "{ \"color\": \"blue\" }"
	expected := &pb.SnakeResponseStatus{
		Message: "Perfect",
	}
	validateWithJson(t, ValidateStart, snakeUrl+"/start", json, expected, 0, 3)
}

func TestValidateStartBadJson(t *testing.T) {
	json := "{ color\": \"blue\" }"
	expected := &pb.SnakeResponseStatus{
		Message: "Bad response format",
		Errors:  []string{"invalid character 'c' looking for beginning of object key string"},
	}
	validateWithJson(t, ValidateStart, snakeUrl+"/start", json, expected, 1, 2)
}

func TestValidateStartBadUrl(t *testing.T) {
	response := ValidateStart("1234", "start")
	require.True(t, strings.Contains(response.Message, "Snake URL not valid"), response.Message)
	require.Equal(t, []string{"invalid url 'start'"}, response.Errors)
}

func validateWithJson(t *testing.T, status func(id string, url string) *pb.SnakeResponseStatus, url string, json string, expected *pb.SnakeResponseStatus, expectedFailed int32, expectedChecksPassed int32) {
	createClient = singleEndpointMockClient(t, url, json, 200)
	response := status("1234", snakeUrl)
	require.True(t, strings.Contains(response.Message, expected.Message), "got: "+response.Message+", expected: "+expected.Message)
	require.Equal(t, expected.Errors, response.Errors)
	require.Equal(t, expectedChecksPassed, response.Score.ChecksPassed, "IncrementPassed count mismatch")
	require.Equal(t, expectedFailed, response.Score.ChecksFailed, "Expected count mismatch")
}
