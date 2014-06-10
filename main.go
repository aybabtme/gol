package main

import (
	"flag"
	"github.com/bradfitz/iter"
	"github.com/nsf/termbox-go"
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"
)

func main() {

	var (
		debug   = flag.Bool("debug", false, "if true, will print debug output to gol.log")
		density = flag.Float64("density", 0.75, "a value between 0 and 1.0, the density of alive cells at the beginning")
		seed    = flag.Int64("seed", time.Now().UnixNano(), "seed for the random number generator")
	)
	flag.Parse()
	rand.Seed(*seed)
	runtime.GOMAXPROCS(runtime.NumCPU())

	if *debug {
		file, err := os.OpenFile("gol.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		log.SetOutput(file)
	} else {
		log.SetOutput(os.Stderr)
	}

	if err := termbox.Init(); err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	go func() {
		for {
			ev := termbox.PollEvent()
			switch ev.Key {
			case termbox.KeyCtrlZ,
				termbox.KeyCtrlC,
				termbox.KeyCtrlD:
				os.Exit(0)
			}
		}
	}()

	h, w := termbox.Size()

	lastBoard := make([][]termbox.Cell, h)
	for i := range lastBoard {
		lastBoard[i] = make([]termbox.Cell, w)
	}
	board := make([][]termbox.Cell, h)
	for i := range board {
		board[i] = make([]termbox.Cell, w)
	}
	for i := range iter.N(h) {
		for j := range iter.N(w) {
			c := lastBoard[i][j]
			c.Ch = ' '
			if rand.Intn(100) > int(*density*100.0) {
				c.Fg = termbox.ColorWhite
				c.Bg = termbox.ColorBlack
			} else {
				c.Fg = termbox.ColorBlack
				c.Bg = termbox.ColorWhite
			}
			board[i][j] = c
		}
	}

	last := time.Now()
	compute := 0
	for _ = range time.Tick(time.Millisecond * 33) {
		compute++
		update(lastBoard, board)

		draw(lastBoard, board)

		if time.Since(last) > time.Second {
			if *debug {
				log.Printf("%d compute/s", compute)
			}
			compute = 0
			last = time.Now()
		}

		termbox.Flush()
	}

}

func draw(last, current [][]termbox.Cell) {
	for x := range last {
		for y := range last[x] {
			future := current[x][y]
			termbox.SetCell(x, y, future.Ch, future.Fg, future.Bg)
			last[x][y] = future
		}
	}
}

func update(last, current [][]termbox.Cell) {
	for x := range last {
		for y := range last[x] {
			var aliveNeigh int
			acc := func(c termbox.Cell, ok bool) {
				if ok && isAlive(c) {
					aliveNeigh++
				}
			}
			acc(get(x-1, y-1, last))
			acc(get(x-1, y, last))
			acc(get(x-1, y+1, last))
			acc(get(x, y-1, last))
			acc(get(x, y+1, last))
			acc(get(x+1, y-1, last))
			acc(get(x+1, y, last))
			acc(get(x+1, y+1, last))

			old := last[x][y]
			future := current[x][y]

			if isAlive(old) {
				if aliveNeigh != 2 && aliveNeigh != 3 {
					kill(&future)
				}
			} else {
				if aliveNeigh == 3 {
					revive(&future)
				}
			}

			current[x][y] = future
		}
	}
}

func isAlive(c termbox.Cell) bool {
	return c.Fg == termbox.ColorWhite && c.Bg == termbox.ColorBlack
}

func revive(c *termbox.Cell) {
	c.Bg = termbox.ColorBlack
	c.Fg = termbox.ColorWhite
}

func kill(c *termbox.Cell) {
	c.Bg = termbox.ColorWhite
	c.Fg = termbox.ColorBlack
}

func get(x, y int, b [][]termbox.Cell) (termbox.Cell, bool) {

	rx := x % len(b)
	ry := y % len(b[0])

	if rx < 0 {
		rx = len(b) + rx
	}
	if ry < 0 {
		ry = len(b[0]) + ry
	}
	return b[rx][ry], true
}
