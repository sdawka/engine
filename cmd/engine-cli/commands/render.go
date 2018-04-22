package commands

import (
	"fmt"
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

	renderTitle(left, top, int(tick.Turn))
	renderBoard(game, top, bottom, left)
	for i, s := range tick.Snakes {
		renderSnake(left, top, i, s)
	}
	renderFood(left, top, tick.Food)

	return termbox.Flush()
}

func renderSnake(left, top, snakeIndex int, s *pb.Snake) {
	for _, b := range s.Body {
		termbox.SetCell(left+int(b.X), top+int(b.Y)+1, ' ', snakeColor, snakeColor)
	}

	for _, c := range s.Name {

	}
}

func renderFood(left, top int, food []*pb.Point) {
	for _, f := range food {
		termbox.SetCell(left+int(f.X), top+int(f.Y)+1, randomFoodEmoji(), defaultColor, bgColor)
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

func renderBoard(game *pb.Game, top, bottom, left int) {
	for i := top + 1; i < bottom; i++ {
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

func renderTitle(left, top, turn int) {
	tbprint(left, top-1, defaultColor, defaultColor, fmt.Sprintf("Battlesnake! - Turn %d", turn))
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
