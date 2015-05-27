package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/mor7/blinkenline/go/bline"
	"github.com/nsf/termbox-go"
)

type Block struct {
	Size  int
	Color byte
	Pos   int
}

var colorLookup = []int{
	0xFF0000,
	0x00FF00,
	0x0000FF,
	0xFFB000,
	0x00FFFF,
	0xFF00FF,
}

// Update frequency in Hz
var speed int

var dropBlock Block
var userBlocks []Block
var running bool
var randChange bool
var shouldDraw bool

func main() {
	initBlocks := flag.Int("blocks", 4, "Initial blocks")
	initspeed := flag.Int("speed", 40, "Initial speed")
	initRandChange := flag.Bool("rand", false, "Should the dropping block change it's color while falling")
	flag.Parse()

	speed = *initspeed
	randChange = *initRandChange

	if speed <= 0 {
		fmt.Println("Speed must be greater than 0")
		return
	}

	if *initBlocks <= 0 {
		fmt.Println("Initial blocks must be greater than 0")
		return
	}

	defer func() {
		fmt.Printf("Speed: %d\n", speed)
	}()

	// Termbox Init
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	events := make(chan termbox.Event, 10)
	go func() {
		for {
			event := termbox.PollEvent()
			if event.Type == termbox.EventKey && event.Key == termbox.KeyEsc {
				running = false
			} else {
				events <- event
			}
		}
	}()

	rand.Seed(time.Now().UnixNano())

	// Bline Init
	err = bline.Init()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer bline.Close()

	// Game Init
	userBlocks = make([]Block, *initBlocks)
	for i := range userBlocks {
		userBlocks[i] = getRandomBlock()
	}

	dropBlock = getRandomBlock()
	dropBlock.Pos = bline.LedCount
	running = true

	go func() {
		for {
			select {
			case ev := <-events:
				if ev.Type == termbox.EventKey {
					// Rotate userblock array
					if ev.Ch == 'a' || ev.Key == termbox.KeyArrowLeft {
						userBlocks = append(userBlocks[1:], userBlocks[0])
					} else if ev.Ch == 'd' || ev.Key == termbox.KeyArrowRight {
						l := len(userBlocks)
						userBlocks = append(userBlocks[l-1:l], userBlocks[:l-1]...)
					}
				}
			default:
			}
			if running {
				running = update()
			}
			time.Sleep(time.Second / time.Duration(speed))
		}
	}()

	// Draw
	for {
		if shouldDraw {
			bline.SendBuffer()
			shouldDraw = false
		}
		if !running {
			return
		}
		time.Sleep(time.Second / time.Duration(60))
	}
}

func getRandomBlock() Block {
	col := byte(rand.Int31() % int32(len(colorLookup)))
	size := 4
	return Block{Size: size, Color: col}
}

func update() bool {
	dropBlock.Pos--

	shouldDraw = false

	bline.ClearBuffer()
	drawBlock(dropBlock)

	pos := 0
	for _, v := range userBlocks {
		v.Pos = pos
		pos += v.Size + 1
		drawBlock(v)
	}
	shouldDraw = true

	if randChange && dropBlock.Pos > pos+60 && rand.Int()%100 == 0 {
		dropBlock.Color = byte(rand.Int31() % int32(len(colorLookup)))
	}

	// Logic
	upperBlock := userBlocks[len(userBlocks)-1]
	if dropBlock.Pos == pos {
		if dropBlock.Color == upperBlock.Color &&
			dropBlock.Size == upperBlock.Size {
			// Play Animation
			explodeTop()
			userBlocks = userBlocks[:len(userBlocks)-1]
			if len(userBlocks) == 0 {
				return false
			}
			dropBlock = getRandomBlock()
			dropBlock.Pos = bline.LedCount
			speed++
		} else {
			userBlocks = append(userBlocks, dropBlock)
			if dropBlock.Pos+dropBlock.Size >= bline.LedCount-1 {
				return false
			}
			dropBlock = getRandomBlock()
			dropBlock.Pos = bline.LedCount
			speed++
		}
	}
	return true
}

func explodeTop() {
	base := 2
	part_range := 4

	for i := 0; i < 8; i++ {
		shouldDraw = false
		bline.ClearBuffer()
		pos := 0
		for i, v := range userBlocks {
			if i == len(userBlocks)-1 {
				break
			}
			v.Pos = pos
			pos += v.Size + 1
			drawBlock(v)
		}

		expls := Block{Pos: pos, Size: 8 - i, Color: dropBlock.Color}
		drawBlock(expls)
		ppos := pos + dropBlock.Size
		col := colorLookup[dropBlock.Color]
		for p := 0; p < i; p++ {
			ppos += base + int(rand.Int31())%part_range
			if ppos < bline.LedCount {
				bline.SetColor(ppos, col)
			}
		}
		shouldDraw = true
		time.Sleep(time.Second / time.Duration(speed))
	}
}

func drawBlock(block Block) {
	col := colorLookup[block.Color]
	for i := block.Pos; i < bline.LedCount && i < block.Pos+block.Size; i++ {
		bline.SetColor(i, col)
	}
}
