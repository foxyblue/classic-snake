package main

import (
	"container/list"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math/rand"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

const (
	size  float64 = 16
	speed float64 = 16
	minX  float64 = 0
	minY  float64 = 0
	maxX  float64 = 1024
	maxY  float64 = 768
)

type Direction int

const (
	UP Direction = iota
	DOWN
	LEFT
	RIGHT
)

type Head struct {
	x      float64
	y      float64
	d      Direction
	bodies *list.List
	size   int
}

type Body struct {
	x float64
	y float64
	d Direction
}

type Food struct {
	x      float64
	y      float64
	sprite *pixel.Sprite
}

func createFood(head *Head, sprite *pixel.Sprite) *Food {
	var x float64 = -1
	var y float64 = -1

	for head.inside(x, y) || (x < 0 && y < 0) {
		x = float64(rand.Intn(int(maxX/size))) * size
		y = float64(rand.Intn(int(maxY/size))) * size
	}

	return &Food{
		x:      x,
		y:      y,
		sprite: sprite,
	}
}

func (head *Head) hasEaten(food *Food) bool {
	if (head.x == food.x) && (head.y == food.y) {
		return true
	}
	return false
}

func createHead(x, y float64, size int) *Head {
	bodies := list.New()
	for i := 0; i <= size; i++ {
		body := createBody(x-float64(i*size), y)
		bodies.PushBack(body)
	}
	return &Head{
		x:      x,
		y:      y,
		d:      RIGHT,
		bodies: bodies,
		size:   size,
	}
}

func (head *Head) grow() {
	body := createBody(head.x, head.y)
	head.size++
	head.bodies.PushBack(body)
}

func (head *Head) collision(x, y float64) bool {
	b := head.bodies.Front()
	for b != nil {
		body := b.Value.(*Body)
		if (body.x == x) && (body.y == y) {
			return true
		}
		b = b.Next()
	}
	return false
}

func (head *Head) inside(x, y float64) bool {
	if (head.x == x) && (head.y == y) {
		return true
	}
	return head.collision(x, y)
}

func (head *Head) update() {
	body := createBody(head.x, head.y)
	head.bodies.PushBack(body)

	switch head.d {
	case UP:
		head.y += size
		if head.y >= maxY {
			head.y = minY
		}
	case RIGHT:
		head.x += size
		if head.x >= maxX {
			head.x = minX
		}
	case DOWN:
		head.y -= size
		if head.y < minY {
			head.y = maxY - size
		}
	case LEFT:
		head.x -= size
		if head.x < minX {
			head.x = maxX - size
		}
	}
	fmt.Printf("Head: x=%v y=%v\n", head.x, head.y)

	e := head.bodies.Front()
	head.bodies.Remove(e)
}

func createBody(x, y float64) *Body {
	return &Body{
		x: x,
		y: y,
	}
}

func (head *Head) draw(imd *imdraw.IMDraw) {
	imd.Color = pixel.RGB(0.13, 0.54, 0.13)
	imd.Push(pixel.V(head.x, head.y), pixel.V(head.x+size, head.y+size))
	imd.Rectangle(0)

	e := head.bodies.Front()
	for e != nil {
		body := e.Value.(*Body)
		body.draw(imd)
		e = e.Next()
	}
}

func (body *Body) draw(imd *imdraw.IMDraw) {
	imd.Color = pixel.RGB(0.13, 0.54, 0.13)
	imd.Push(pixel.V(body.x, body.y), pixel.V(body.x+size, body.y+size))
	imd.Rectangle(0)
}

func (food *Food) draw(win *pixelgl.Window) {
	food.sprite.Draw(win, pixel.IM.Moved(pixel.V(food.x+(size/2), food.y+(size/2))))
}

func displayText(win *pixelgl.Window, p string) {
	dt := float64(len(p)) / 2
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	basicTxt := text.New(
		pixel.V(float64(maxY/2), float64(maxX/2)-dt),
		basicAtlas,
	)
	basicTxt.Color = color.Black
	fmt.Fprintln(basicTxt, p)
	basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 4))
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

func run() {
	var alive bool

	cfg := pixelgl.WindowConfig{
		Title:  "Snake",
		Bounds: pixel.R(0, 0, maxX, maxY),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	spriteSheet, err := loadPicture("snake-sprites.png")
	if err != nil {
		panic(err)
	}
	foodFrame := pixel.R(spriteSheet.Bounds().Min.X, spriteSheet.Bounds().Min.Y, size, size)
	foodSprite := pixel.NewSprite(spriteSheet, foodFrame)

	imd := imdraw.New(nil)

	head := createHead(64, 64, 3)
	food := createFood(head, foodSprite)
	alive = true

	for !win.Closed() {
		win.SetClosed(win.JustPressed(pixelgl.KeyEscape))

		if win.JustPressed(pixelgl.KeyUp) {
			if head.d != DOWN {
				head.d = UP
			}
		}
		if win.JustPressed(pixelgl.KeyDown) {
			if head.d != UP {
				head.d = DOWN
			}
		}
		if win.JustPressed(pixelgl.KeyLeft) {
			if head.d != RIGHT {
				head.d = LEFT
			}
		}
		if win.JustPressed(pixelgl.KeyRight) {
			if head.d != LEFT {
				head.d = RIGHT
			}
		}

		if alive {
			imd.Clear()
			win.Clear(colornames.Aliceblue)

			head.update()
			if head.hasEaten(food) {
				head.grow()
				food = createFood(head, foodSprite)
			} else if head.collision(head.x, head.y) {
				alive = false
			}
			head.draw(imd)
			food.draw(win)
		} else {
			displayText(win, "Game Over")
		}

		imd.Draw(win)
		win.Update()
		time.Sleep(64 * time.Millisecond)
	}
}

func main() {
	pixelgl.Run(run)
}
