package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	termbox "github.com/nsf/termbox-go"
	"github.com/spf13/cobra"
)

func init() {
	replayCmd.Flags().StringVarP(&gameID, "game-id", "g", "", "the game id of the game to get the status of")
}

var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "replays an existing game on the battlesnake engine",
	Args: func(c *cobra.Command, args []string) error {
		if len(gameID) == 0 {
			return errors.New("game id is required")
		}
		return nil
	},
	Run: func(*cobra.Command, []string) {
		replayGame()
	},
}

func moveFrameForwards(frameIndex int, frames []*pb.GameTick) (int, *pb.GameTick, bool) {
	frameIndex++
	if frameIndex >= len(frames) {
		return frameIndex, nil, true
	}
	return frameIndex, frames[frameIndex], false
}

func moveFrameBackwards(frameIndex int, frames []*pb.GameTick) (int, *pb.GameTick) {
	frameIndex--
	if frameIndex <= 0 {
		frameIndex = 0
	}
	return frameIndex, frames[frameIndex]
}

func getCharacter(gameTick *pb.GameTick, x, y int64) string {
	for _, f := range gameTick.Food {
		if f.X == x && f.Y == y {
			return "●"
		}
	}

	for _, s := range gameTick.AliveSnakes() {
		for _, p := range s.Body {
			if p.X == x && p.Y == y {
				return "◼"
			}
		}
	}
	return " "
}

func loadGame() (*pb.Game, []*pb.GameTick, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	tr := &pb.ListGameTicksResponse{}
	var game *pb.Game
	{
		resp, err := client.Get(fmt.Sprintf("%s/games/%s/ticks", apiAddr, gameID))
		if err != nil {
			fmt.Println("error while getting ticks", err)
			return nil, nil, err
		}
		err = json.NewDecoder(resp.Body).Decode(tr)
		resp.Body.Close()
		if err != nil {
			fmt.Println("error while decoding ticks", err)
			return nil, nil, err
		}
	}
	{
		resp, err := client.Get(fmt.Sprintf("%s/games/%s", apiAddr, gameID))
		if err != nil {
			fmt.Println("error while getting status", err)
			return nil, nil, err
		}
		s := &pb.StatusResponse{}
		err = json.NewDecoder(resp.Body).Decode(s)
		resp.Body.Close()
		if err != nil {
			fmt.Println("error while getting status", err)
			return nil, nil, err
		}
		game = s.Game
	}
	return game, tr.Ticks, nil
}

func replayGame() {
	game, frames, err := loadGame()
	if err != nil {
		panic(err)
	}

	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	eventQueue := make(chan termbox.Event)
	go func(ev chan<- termbox.Event) {
		for {
			ev <- termbox.PollEvent()
		}
	}(eventQueue)

	cycle := time.NewTicker(200 * time.Millisecond)
	if len(frames) == 0 {
		return
	}
	currentFrame := frames[0]
	frameIndex := 0
	paused := false
	done := false

	for {
		if done {
			break
		}

		select {
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey {
				switch ev.Key {
				case termbox.KeyEsc:
					done = true
				case termbox.KeySpace:
					paused = !paused
					if paused {
						cycle.Stop()
					} else {
						cycle = time.NewTicker(200 * time.Millisecond)
					}

				case termbox.KeyArrowLeft:
					paused = true
					frameIndex, currentFrame = moveFrameBackwards(frameIndex, frames)
					if err := render(game, currentFrame); err != nil {
						panic(err)
					}
				case termbox.KeyArrowRight:
					paused = true
					frameIndex, currentFrame, done = moveFrameForwards(frameIndex, frames)
					if err := render(game, currentFrame); err != nil {
						panic(err)
					}
				}

			}
		case <-cycle.C:
			if paused {
				continue
			}
			if err := render(game, currentFrame); err != nil {
				panic(err)
			}
			frameIndex, currentFrame, done = moveFrameForwards(frameIndex, frames)

		}
	}

	if !done {
		tbprint(0, 0, defaultColor, defaultColor, "Press any key to exit...")
		termbox.Flush()
		termbox.PollEvent()
	}
}
