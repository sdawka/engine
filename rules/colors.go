package rules

import "sync"

var defaultColors = []string{
	"#8f4949",
	"#49628f",
	"#7f498f",
	"#8f7f49",
	"#628f49",
	"#491010",
	"#493810",
	"#164910",
	"#104947",
	"#3e1049",
	"#cd1e91",
	"#741ecd",
	"#1e4fcd",
	"#1ecdc7",
	"#1ecd3f",
	"#cdcb1e",
	"#cd681e",
}

var palette = defaultColors

var colorIndex = 0

var colorMutex = &sync.Mutex{}

func nextColorIndex() int {
	colorMutex.Lock()
	defer colorMutex.Unlock()

	current := colorIndex
	colorIndex = (colorIndex + 1) % len(palette)
	return current
}

func nextColor() string {
	return palette[nextColorIndex()]
}

func resetPalette(colors []string) {
	colorMutex.Lock()
	defer colorMutex.Unlock()

	palette = colors
	colorIndex = 0
}
