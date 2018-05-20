package rules

type GameStatus string

var (
	// GameStatusStopped represents a stopped game
	GameStatusStopped GameStatus = "stopped"
	// GameStatusRunning represents a running game
	GameStatusRunning GameStatus = "running"
	// GameStatusError represents a game that ended because of an error
	GameStatusError GameStatus = "error"
	// GameStatusComplete represents a game that is done
	GameStatusComplete GameStatus = "complete"
)
