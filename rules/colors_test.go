package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNextColor(t *testing.T) {
	resetPalette([]string{"red", "green", "blue"})

	// test each color
	require.Equal(t, "red", nextColor())
	require.Equal(t, "green", nextColor())
	require.Equal(t, "blue", nextColor())

	// test wrap around
	require.Equal(t, "red", nextColor())
}

func resetPalette(colors []string) {
	colorMutex.Lock()
	defer colorMutex.Unlock()

	palette = colors
	colorIndex = 0
}
