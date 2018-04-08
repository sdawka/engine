package pb

// AliveSnakes returns all the alive snakes
func (gt *GameTick) AliveSnakes() []*Snake {
	snakes := []*Snake{}

	for _, s := range gt.Snakes {
		if s.Death == nil {
			snakes = append(snakes, s)
		}
	}

	return snakes
}

// DeadSnakes returns all the dead snakes
func (gt *GameTick) DeadSnakes() []*Snake {
	snakes := []*Snake{}

	for _, s := range gt.Snakes {
		if s.Death != nil {
			snakes = append(snakes, s)
		}
	}

	return snakes
}
