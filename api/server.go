package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// Server this is the api server
type Server struct {
	hs *http.Server
}

// New creates a new api server
func New(addr string, c pb.ControllerClient) *Server {
	router := httprouter.New()
	router.POST("/game/create", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.WithError(err).Error("Unable to read request body")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		req := &pb.CreateRequest{}
		// TODO: handle error
		json.Unmarshal(body, req)
		// TODO: use a context with timeout, and handle error
		resp, _ := c.Create(context.Background(), req)

		// TODO: Handle error
		j, _ := json.Marshal(resp)
		w.Write(j)
	})
	router.POST("/game/start/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")
		req := &pb.StartRequest{
			ID: id,
		}
		// TODO: use a context with timeout
		_, err := c.Start(context.Background(), req)
		if err != nil {
			log.WithError(err).WithField("req", req).Error("Error while calling controller start")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	router.GET("/game/status/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		id := ps.ByName("id")
		req := &pb.StatusRequest{
			ID: id,
		}
		// TODO: use a context with timeout
		resp, err := c.Status(context.Background(), req)
		if err != nil {
			log.WithError(err).WithField("req", req).Error("Error while calling controller status")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		// TODO: Handle error
		j, _ := json.Marshal(resp)
		w.Write(j)
	})

	return &Server{
		hs: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}

// WaitForExit starts up the server and blocks until the server shuts down.
func (s *Server) WaitForExit() {
	log.Infof("Battlesnake engine api listening on %s", s.hs.Addr)
	err := s.hs.ListenAndServe()
	if err != nil {
		log.Errorf("Error while listening: %v", err)
	}
}
