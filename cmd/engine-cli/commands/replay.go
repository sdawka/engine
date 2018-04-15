package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/battlesnakeio/engine/controller/pb"
	tm "github.com/buger/goterm"
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

		var ticks []*pb.GameTick
		var game *pb.Game
		{
			resp, err := client.Get(fmt.Sprintf("%s/games/%s/ticks", apiAddr, gameID))
			if err != nil {
				fmt.Println("error while getting ticks", err)
				return
			}
			err = json.NewDecoder(resp.Body).Decode(ticks)
			resp.Body.Close()
			if err != nil {
				fmt.Println("error while getting ticks", err)
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

		tm.Clear() // Clear current screen

		for _, gt := range ticks {
			// By moving cursor to top-left position we ensure that console output
			// will be overwritten each time, instead of adding new.
			tm.MoveCursor(1, 1)

			tm.Print("┌")
			for x := int64(0); x < game.Width; x++ {
				tm.Print("─")
			}
			tm.Println("┐")
			for y := int64(0); y < game.Height; y++ {
				tm.Print("│")
				for x := int64(0); x < game.Width; x++ {
					c := getCharacter(gt, x, y)
					tm.Print(c)
				}
				tm.Println("│")
			}

			tm.Print("└")
			for x := int64(0); x < game.Width; x++ {
				tm.Print("─")
			}
			tm.Println("┘")

			tm.Println()

			tm.Flush() // Call it every time at the end of rendering

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
