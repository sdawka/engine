package rules

import "github.com/battlesnakeio/engine/controller/pb"

// CheckForDeath looks through the snakes with the updated coords and checks to see if any have died
// possible death options are starvation (health has reached 0), wall collision, snake body collision
// snake head collision (other snake is same size or greater)
func CheckForDeath(game *pb.Game) {}
