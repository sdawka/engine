package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

		resp, err := client.Get(fmt.Sprintf("%s/game/status/%s", apiAddr, gameID))
		if err != nil {
			fmt.Println("error while posting to status endpoint", err)
			return
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("unable to read response body", err)
			return
		}

		s := &pb.StatusResponse{}
		err = json.Unmarshal(data, s)
		if err != nil {
			fmt.Println(string(data))
			fmt.Println("unable to unmarshal status response: ", err)
			return
		}

		tm.Clear() // Clear current screen

		for _, gt := range s.Game.Ticks {
			// By moving cursor to top-left position we ensure that console output
			// will be overwritten each time, instead of adding new.
			tm.MoveCursor(1, 1)

			tm.Print("┌")
			for x := int64(0); x < s.Game.Width; x++ {
				tm.Print("─")
			}
			tm.Println("┐")
			for y := int64(0); y < s.Game.Height; y++ {
				tm.Print("│")
				for x := int64(0); x < s.Game.Width; x++ {
					c := getCharacter(gt, x, y)
					tm.Print(c)
				}
				tm.Println("│")
			}

			tm.Print("└")
			for x := int64(0); x < s.Game.Width; x++ {
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
