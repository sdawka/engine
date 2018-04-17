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
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		tr := &pb.ListGameTicksResponse{}
		var game *pb.Game
		{
			resp, err := client.Get(fmt.Sprintf("%s/games/%s/ticks", apiAddr, gameID))
			if err != nil {
				fmt.Println("error while getting ticks", err)
				return
			}
			err = json.NewDecoder(resp.Body).Decode(tr)
			resp.Body.Close()
			if err != nil {
				fmt.Println("error while decoding ticks", err)
				return
			}
		}
		{
			resp, err := client.Get(fmt.Sprintf("%s/games/%s", apiAddr, gameID))
			if err != nil {
				fmt.Println("error while getting status", err)
				return
			}
			s := &pb.StatusResponse{}
			err = json.NewDecoder(resp.Body).Decode(s)
			resp.Body.Close()
			if err != nil {
				fmt.Println("error while getting status", err)
				return
			}
			game = s.Game
		}

		if err := termbox.Init(); err != nil {
			panic(err)
		}
		defer termbox.Close()

		for _, gt := range tr.Ticks {
			if err := render(game, gt); err != nil {
				panic(err)
			}

			time.Sleep(200 * time.Millisecond)
		}
	},
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
