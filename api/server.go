package api

import (
	"net/http"

	"github.com/cloudflare/cfssl/log"
)

type Server struct {
	hs *http.Server
}

func New() *Server {
	return &Server{
		hs: &http.Server{
			Addr: ":5000",
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
