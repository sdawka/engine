package rules

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	log "github.com/sirupsen/logrus"
)

// NotifyGameEnd sends the /end requests to all the snakes.
func NotifyGameEnd(game *pb.Game, frame *pb.GameTick) {
	netClient := createClient(200 * time.Millisecond)

	for _, s := range frame.Snakes {
		req := buildSnakeRequest(game, frame, s.ID)
		data, err := json.Marshal(req)

		if err != nil {
			log.WithError(err).Errorf("error while marshaling snake request: %s", s.ID)
			return
		}

		buf := bytes.NewBuffer(data)
		_, err = netClient.Post(getURL(s.URL, "end"), "application/json", buf)
		if err != nil {
			log.WithError(err).Errorf("error POSTing to /end for snake %s", s.ID)
		}
	}
}
