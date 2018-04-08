package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "gets the status of a game from the battlesnake engine",
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

		fmt.Println(string(data))
	},
}

var (
	gameID string
)

func init() {
	statusCmd.Flags().StringVarP(&gameID, "game-id", "g", "", "the game id of the game to get the status of")
}
