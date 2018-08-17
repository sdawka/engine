package rules

import (
	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

var snakeURL = "http://good-server"

func TestValidateEnd(t *testing.T) {
	expected := &pb.SnakeResponseStatus{
		Message:    "Perfect",
		StatusCode: 200,
		Raw:        "",
		Score: &pb.Score{
			ChecksPassed: 3,
			ChecksFailed: 0,
		},
	}
	validateWithJSON(t, ValidateEnd, snakeURL+"/end", expected)
}
func TestValidateMove(t *testing.T) {
	expected := &pb.SnakeResponseStatus{
		Message:    "Perfect",
		Raw:        "{  }",
		StatusCode: 200,
		Score: &pb.Score{
			ChecksPassed: 3,
			ChecksFailed: 0,
		},
	}
	validateWithJSON(t, ValidateMove, snakeURL+"/move", expected)
}

func TestValidateStart(t *testing.T) {
	expected := &pb.SnakeResponseStatus{
		Message:    "Perfect",
		Raw:        "{ \"color\": \"blue\" }",
		StatusCode: 200,
		Score: &pb.Score{
			ChecksPassed: 3,
			ChecksFailed: 0,
		},
	}
	validateWithJSON(t, ValidateStart, snakeURL+"/start", expected)
}

func TestValidateStartBadJson(t *testing.T) {
	expected := &pb.SnakeResponseStatus{
		Message:    "Bad response format",
		Raw:        "{ color\": \"blue\" }",
		StatusCode: 200,
		Errors:     []string{"invalid character 'c' looking for beginning of object key string"},
		Score: &pb.Score{
			ChecksPassed: 2,
			ChecksFailed: 1,
		},
	}
	validateWithJSON(t, ValidateStart, snakeURL+"/start", expected)
}

func TestValidateStartBadUrl(t *testing.T) {
	response := ValidateStart("1234", "start")
	require.True(t, strings.Contains(response.Message, "Snake URL not valid"), response.Message)
	require.Equal(t, []string{"invalid url 'start'"}, response.Errors)
}

func validateWithJSON(t *testing.T, status func(id string, url string) *pb.SnakeResponseStatus, url string, expected *pb.SnakeResponseStatus) {
	createClient = singleEndpointMockClient(t, url, expected.Raw, 200)
	response := status("1234", snakeURL)
	require.True(t, strings.Contains(response.Message, expected.Message), "got: "+response.Message+", expected: "+expected.Message)
	require.Equal(t, expected.Errors, response.Errors)
	require.Equal(t, expected.Raw, response.Raw)
	require.Equal(t, expected.Score.ChecksPassed, response.Score.ChecksPassed, "Passed count mismatch")
	require.Equal(t, expected.Score.ChecksFailed, response.Score.ChecksFailed, "Failed count mismatch")
}
