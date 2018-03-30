package api

import (
	"io"
	"net/http"

	"github.com/battlesnakeio/engine/controller"
	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	hs *http.Server
}

func readRequestBody(b io.ReadCloser) interface{} {
	return nil
}

func New(c *controller.Controller) *Server {
	router := httprouter.New()
	router.GET("/game/create", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		c.Create(readRequestBody(r.Body).(controller.GameConfiguration))
	})
	router.POST("/game/start/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		c.Start(readRequestBody(r.Body))
	})

	return &Server{
		hs: &http.Server{
			Addr:    ":5000",
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
