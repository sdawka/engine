package server

import (
	log "github.com/sirupsen/logrus"

	"github.com/battlesnakeio/engine/controller"
	"github.com/spf13/cobra"
)

var (
	controllerListen  = ":3004"
	controllerBackend = "inmem"
)

func init() {
	controllerCmd.Flags().StringVarP(&controllerListen, "listen", "l", controllerListen, "address for the controller to bind to")
	controllerCmd.Flags().StringVarP(&controllerBackend, "backend", "b", controllerBackend, "controller backend, as one of: [inmem]")
	RootCmd.Flags().AddFlagSet(controllerCmd.Flags())
}

var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "runs the engine controller",
	Run: func(c *cobra.Command, args []string) {
		var store controller.Store
		switch controllerBackend {
		case "inmem":
			store = controller.InMemStore()
		default:
			log.WithField("backend", controllerBackend).Fatal("invalid backend")
		}

		ctrl := controller.New(store)
		log.WithField("listen", controllerListen).
			Info("Battlesnake controller serving")
		if err := ctrl.Serve(controllerListen); err != nil {
			log.WithError(err).
				WithField("listen", controllerListen).
				Fatal("Controller failed to serve")
		}
	},
}
