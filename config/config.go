package config

import (
	"os"
	"strconv"

	"golang.org/x/time/rate"
)

// Configuration variables. These aren't user facing but useful for tuning the
// details of engine performance.
var (
	MaxOpenConns = getEnvInt("MAX_OPEN_CONNS", 20)
	MaxIdleConns = getEnvInt("MAX_IDLE_CONNS", 20)
	PopRate      = rate.Limit(getEnvInt("POP_RPS", 40))
	PopBurstRate = getEnvInt("POP_BURST", 10)
)

func getEnvInt(varName string, defaults int) int {
	val := os.Getenv(varName)
	if val == "" {
		return defaults
	}
	intVal, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return defaults
	}
	return int(intVal)
}
