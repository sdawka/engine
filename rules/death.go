package rules

import "github.com/battlesnakeio/engine/controller/pb"

type deathUpdate struct {
	Snake *pb.Snake
	Death *pb.Death
}

// checkForDeath looks through the snakes with the updated coords and checks to see if any have died
// possible death options are starvation (health has reached 0), wall collision, snake body collision
// snake head collision (other snake is same size or greater)
func checkForDeath(width, height int64, tick *pb.GameTick) []deathUpdate {
	updates := []deathUpdate{}
	for _, s := range tick.AliveSnakes() {
		if s.Health == 0 {
			updates = append(updates, deathUpdate{
				Snake: s,
				Death: &pb.Death{
					Turn:  tick.Turn,
					Cause: DeathCauseStarvation,
				},
			})
			continue
		}
		if s.Head().X < 0 || s.Head().X >= width || s.Head().Y < 0 || s.Head().Y >= height {
			updates = append(updates, deathUpdate{
				Snake: s,
				Death: &pb.Death{
					Turn:  tick.Turn,
					Cause: DeathCauseWallCollision,
				},
			})
			continue
		}

		for _, other := range tick.AliveSnakes() {
			if other.ID == s.ID {
				continue
			}

			for i, b := range other.Body {
				if i == 0 && s.Head().Equal(b) {
					if len(s.Body) <= len(other.Body) {
						updates = append(updates, deathUpdate{
							Snake: s,
							Death: &pb.Death{
								Turn:  tick.Turn,
								Cause: DeathCauseHeadToHeadCollision,
							},
						})
						break
					}
				}

				if s.Head().Equal(b) {
					updates = append(updates, deathUpdate{
						Snake: s,
						Death: &pb.Death{
							Turn:  tick.Turn,
							Cause: DeathCauseSnakeCollision,
						},
					})
					break
				}
			}
		}
	}
	return updates
}
