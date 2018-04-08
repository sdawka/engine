package pb

// Clone clones a point and returns a new point
func (p *Point) Clone() *Point {
	return &Point{X: p.X, Y: p.Y}
}
