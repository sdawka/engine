package api

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	hs *http.Server
}

func readRequestBody(b io.ReadCloser) interface{} {
	return nil
}

func New(addr string, c pb.ControllerClient) *Server {
	router := httprouter.New()
	router.GET("/game/create", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
		// TODO: use a context with timeout
		c.Create(context.Background(), req)
	})
	router.POST("/game/start/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.WithError(err).Error("Unable to read request body")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		req := &pb.StartRequest{}
		// TODO: handle error
		json.Unmarshal(body, req)
		// TODO: use a context with timeout
		c.Start(context.Background(), req)
	})

	return &Server{
		hs: &http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}

func (s *Server) WaitForExit() {
	log.Infof("Battlesnake engine listening on %s", s.hs.Addr)
	err := s.hs.ListenAndServe()
	if err != nil {
		log.Errorf("Error while listening: %v", err)
	}
}
