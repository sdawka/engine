package pb

// Move the snake 1 space in the specified direction, move does not remove the end point of the snake, that will be done
// after snakes have eaten
func (*Snake) Move(direction string) {}

// DefaultMove the snake will move 1 space in the direction it was already heading
func (*Snake) DefaultMove() {}

// Head returns the first point in the body
func (s *Snake) Head() *Point {
	return s.Body[0]
}
