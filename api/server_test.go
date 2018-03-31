package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type MockController struct {
	pb.ControllerClient

	Error          error
	CreateResponse *pb.CreateResponse
	StartResponse  *pb.StartResponse
	StatusResponse *pb.StatusResponse
}

func (mc *MockController) Create(ctx context.Context, req *pb.CreateRequest, opts ...grpc.CallOption) (*pb.CreateResponse, error) {
	return mc.CreateResponse, mc.Error
}

func (mc *MockController) Start(ctx context.Context, req *pb.StartRequest, opts ...grpc.CallOption) (*pb.StartResponse, error) {
	return mc.StartResponse, mc.Error
}

func (mc *MockController) Status(ctx context.Context, req *pb.StatusRequest, opts ...grpc.CallOption) (*pb.StatusResponse, error) {
	return mc.StatusResponse, mc.Error
}

func createAPIServer() (*Server, *MockController) {
	var client = &MockController{
		CreateResponse: &pb.CreateResponse{},
		StartResponse:  &pb.StartResponse{},
		StatusResponse: &pb.StatusResponse{},
	}
	s := New(":1234", client)
	return s, client
}

func TestCreate(t *testing.T) {
	s, _ := createAPIServer()

	buf := &bytes.Buffer{}
	buf.WriteString("{}")
	req, _ := http.NewRequest("POST", "/game/create", buf)
	rr := httptest.NewRecorder()

	s.hs.Handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestStart(t *testing.T) {
	s, _ := createAPIServer()

	req, _ := http.NewRequest("POST", "/game/start/abc_123", nil)
	rr := httptest.NewRecorder()

	s.hs.Handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestStatus(t *testing.T) {
	s, _ := createAPIServer()

	req, _ := http.NewRequest("GET", "/game/status/abc_123", nil)
	rr := httptest.NewRecorder()

	s.hs.Handler.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
}
