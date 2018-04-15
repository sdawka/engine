package commands

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	runCmd.Flags().StringVarP(&gameID, "game-id", "g", "", "the game id of the game to get the status of")
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "runs an existing game on the battlesnake engine",
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

		resp, err := client.Post(fmt.Sprintf("%s/games/%s/start", apiAddr, gameID), "application/json", &bytes.Buffer{})
		if err != nil {
			fmt.Println("error while posting to start endpoint", err)
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
