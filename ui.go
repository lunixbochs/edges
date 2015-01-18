package main

import (
	"bytes"
	"github.com/lunixbochs/termbox-go"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

func getCell(x, y int) *termbox.Cell {
	buf := termbox.CellBuffer()
	width, height := termbox.Size()
	if x >= 0 && y >= 0 && x < width && y < height {
		return &buf[y*width+x]
	}
	return nil
}

func Output(x, y int, msg string) error {
	for i, c := range msg {
		xd := x + i
		if cell := getCell(xd, y); cell != nil {
			cell.Ch = c
		}
	}
	return nil
}

type P struct {
	x, y int
}

var start = &P{}
var maxLen int

func redraw(board []byte, cx, cy int, path []*P) {
	fg, bg := termbox.ColorDefault, termbox.ColorDefault
	termbox.Clear(fg, bg)
	bounds := &P{}
	for i, line := range bytes.Split(board, []byte{'\n'}) {
		if len(line) > 0 {
			bounds.x = len(line)
			bounds.y = i + 1
			Output(1, i, string(line))
			idx := bytes.Index(line, []byte{'@'})
			if idx != -1 {
				start.x = i + 1
				start.y = idx
			}
		}
	}
	if len(path) > 1 {
		bg := termbox.ColorCyan
		if len(path)-1 > maxLen {
			bg = termbox.ColorYellow
		}
		origin := path[0]
		x, y := origin.x, origin.y
		iters := 0
	Loop:
		for {
			for i, p := range path[1:] {
				iters++
				x += p.x
				y += p.y
				cell := getCell(x, y)
				if cell == nil || x > bounds.x || y > bounds.y || x < 0 || y < 0 {
					break Loop
				}
				if cell.Ch == 'x' || iters > 10000 && cell.Bg == bg {
					cell.Fg = termbox.ColorWhite
					cell.Bg = termbox.ColorRed
					break Loop
				}
				if i == 0 && cell.Ch == ' ' {
					cell.Ch = '+'
				}
				cell.Bg = bg
			}
		}
		str := ""
		for _, p := range path[1:] {
			c := "?"
			if p.x != 0 && p.y != 0 {
				c = "!"
			} else if p.x == 1 {
				c = "d"
			} else if p.x == -1 {
				c = "a"
			} else if p.y == 1 {
				c = "s"
			} else if p.y == -1 {
				c = "w"
			}
			str += c
		}
		Output(3, 0, str)
	}
	startC := getCell(start.x, start.y)
	if startC != nil && startC.Ch == '@' {
		startC.Fg = termbox.ColorCyan
	}
	cell := getCell(cx, cy)
	if cell != nil {
		cell.Bg = termbox.ColorRed
	}
	termbox.Flush()
}

func main() {
	if len(os.Args) > 1 {
		maxLen, _ = strconv.Atoi(os.Args[1])
	}
	board, err := ioutil.ReadFile("board.txt")
	if err != nil {
		log.Fatal(err)
	}
	termbox.Init()
	defer termbox.Close()
	redraw(board, 0, 0, nil)
	x := start.x
	y := start.y
	path := []*P{&P{x, y}}
	var undoStack [][]*P
	active := true
	for {
		origin := &P{x, y}
		redraw(board, x, y, path)
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Ch {
			case 'h':
				x--
			case 'j':
				y++
			case 'k':
				y--
			case 'l':
				x++
			case 'H':
				x -= 5
			case 'J':
				y += 5
			case 'K':
				y -= 5
			case 'L':
				x += 5
			case 'u':
				if len(undoStack) > 0 {
					active = false
					end := len(undoStack) - 1
					path = undoStack[end]
					undoStack = undoStack[:end]
				}
			default:
				origin = nil
			}
			if origin != nil && active {
				p := &P{x - origin.x, y - origin.y}
				top := path[len(path)-1]
				if len(path) > 1 && top.x == -p.x && top.y == -p.y {
					path = path[:len(path)-1]
				} else {
					path = append(path, p)
				}
			}
			switch ev.Key {
			/*
				case termbox.KeyArrowDown:
					pos++
				case termbox.KeyArrowUp:
					pos--
			*/
			case termbox.KeyCtrlC:
				return
			case termbox.KeyEsc:
				startC := getCell(start.x, start.y)
				if startC != nil && startC.Ch == '@' {
					x, y = start.x, start.y
				}
				if path != nil {
					undoStack = append(undoStack, path)
				}
				active = false
				path = nil
			case termbox.KeyBackspace2:
				if active && len(path) > 1 {
					end := len(path) - 1
					pop := path[end]
					x -= pop.x
					y -= pop.y
					path = path[:end]
				}
			case termbox.KeySpace:
				if !active {
					if path != nil {
						undoStack = append(undoStack, path)
					}
					path = []*P{&P{x, y}}
				}
				active = !active
			}
		case termbox.EventResize:
			termbox.Flush()
		}
	}

}
