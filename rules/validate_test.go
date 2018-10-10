package rules

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
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
func TestValidateSlowUrl(t *testing.T) {
	url := snakeURL + "/move"
	createClient = singleSlowEndpointMockClient(t, url, "{}", 200, 1000*time.Millisecond)
	response := ValidateMove("1234", snakeURL)
	errorZero := response.Errors[0]
	var digitsRegexp = regexp.MustCompile(`snake took (\d+) ms`)
	digitString := digitsRegexp.FindStringSubmatch(errorZero)
	digits, _ := strconv.Atoi(digitString[1])
	fmt.Printf("%d\n", digits)
	assert.True(t, digits > 1000, "Should have found larger amount than the slow snake limit")
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

func singleSlowEndpointMockClient(t *testing.T, url, bodyJSON string, statusCode int, sleep time.Duration) func(time.Duration) httpClient {
	body := readCloser{Buffer: &bytes.Buffer{}}
	body.WriteString(bodyJSON)

	return func(time.Duration) httpClient {
		return mockHTTPClient{
			resp: func(reqURL string) *http.Response {
				time.Sleep(sleep)
				if reqURL != url {
					require.Fail(t, "invalid url")
				}
				return &http.Response{
					Body:       body,
					StatusCode: statusCode,
				}
			},
		}
	}
}
