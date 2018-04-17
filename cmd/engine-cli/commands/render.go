package commands

import (
	"math/rand"

	"github.com/battlesnakeio/engine/controller/pb"
	"github.com/mattn/go-runewidth"
	termbox "github.com/nsf/termbox-go"
)

const (
	defaultColor = termbox.ColorDefault
	bgColor      = termbox.ColorDefault
	snakeColor   = termbox.ColorGreen
)

func render(game *pb.Game, tick *pb.GameTick) error {
	termbox.Clear(defaultColor, defaultColor)

	var (
		w, h   = termbox.Size()
		midY   = h / 2
		left   = (w - int(game.Width)) / 2
		top    = midY - (int(game.Height) / 2)
		bottom = midY + (int(game.Height) / 2) + 1
	)

	renderTitle(left, top)
	renderBoard(game, top, bottom, left)
	for _, s := range tick.Snakes {
		renderSnake(left, bottom, s)
	}
	renderFood(left, bottom, tick.Food)

	return termbox.Flush()
}

func renderSnake(left, bottom int, s *pb.Snake) {
	for _, b := range s.Body {
		termbox.SetCell(left+int(b.X), bottom-int(b.Y), ' ', snakeColor, snakeColor)
	}
}

func randomFoodEmoji() rune {
	f := []rune{
		'🍒',
		'🍍',
		'🍑',
		'🍇',
		'🍏',
		'🍌',
		'🍫',
		'🍭',
		'🍕',
		'🍩',
		'🍗',
		'🍖',
		'🍬',
		'🍤',
		'🍪',
	}

	return f[rand.Intn(len(f))]
}

func renderFood(left, bottom int, food []*pb.Point) {
	for _, f := range food {
		termbox.SetCell(left+int(f.X), bottom-int(f.Y), randomFoodEmoji(), defaultColor, bgColor)
	}
}

func renderBoard(game *pb.Game, top, bottom, left int) {
	for i := top; i < bottom; i++ {
		termbox.SetCell(left-1, i, '│', defaultColor, bgColor)
		termbox.SetCell(left+int(game.Width), i, '│', defaultColor, bgColor)
	}

	termbox.SetCell(left-1, top, '┌', defaultColor, bgColor)
	termbox.SetCell(left-1, bottom, '└', defaultColor, bgColor)
	termbox.SetCell(left+int(game.Width), top, '┐', defaultColor, bgColor)
	termbox.SetCell(left+int(game.Width), bottom, '┘', defaultColor, bgColor)

	fill(left, top, int(game.Width), 1, termbox.Cell{Ch: '─'})
	fill(left, bottom, int(game.Width), 1, termbox.Cell{Ch: '─'})
}

func renderTitle(left, top int) {
	tbprint(left, top-1, defaultColor, defaultColor, "Snake Game")
}

func fill(x, y, w, h int, cell termbox.Cell) {
	for ly := 0; ly < h; ly++ {
		for lx := 0; lx < w; lx++ {
			termbox.SetCell(x+lx, y+ly, cell.Ch, cell.Fg, cell.Bg)
		}
	}
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}
