package pb

// Equals checks if 2 points are the same x,y coordinate
func (p *Point) Equals(other *Point) bool {
	return p.X == other.X && p.Y == other.Y
}
