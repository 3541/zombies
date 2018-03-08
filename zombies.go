package main

import (
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"

	"github.com/gonum/graph/simple"

	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"

	"github.com/3541/zombies/vis"
)

const (
	CAMERA_SPEED = 600.0
	ZOOM_SPEED   = 1.01
)

func entry() {
	config := pixelgl.WindowConfig{
		Title:  "Apocalypse Simulator 2018",
		Bounds: pixel.R(0, 0, 1920, 1080),
		//VSync:  true,
	}

	window, err := pixelgl.NewWindow(config)
	if err != nil {
		panic(err)
	}

	window.SetSmooth(true)

	draw := imdraw.New(nil)
	draw.Color = colornames.Black
	g := vis.NewMapGraph(window, draw, text.NewAtlas(basicfont.Face7x13, text.ASCII), pixel.R(0, 0, 1000, 1000))

	n1 := g.NewPositionedNode("1", 500, 500)
	g.AddNode(n1)
	n2 := g.NewPositionedNode("2", 800, 600)
	g.AddNode(n2)
	g.SetEdge(simple.Edge{n1, n2, 3})

	frames := 0
	timer := time.Tick(time.Second)

	logAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	logText := text.New(pixel.V(10, 10), logAtlas)
	logText.Color = colornames.Black

	cameraPosition := window.Bounds().Center()
	cameraZoom := 1.0

	lastFrame := time.Now()

	for !window.Closed() {
		timeElapsed := time.Since(lastFrame).Seconds()
		lastFrame = time.Now()

		if window.Pressed(pixelgl.KeyLeft) {
			cameraPosition.X -= CAMERA_SPEED * timeElapsed
		}

		if window.Pressed(pixelgl.KeyRight) {
			cameraPosition.X += CAMERA_SPEED * timeElapsed
		}

		if window.Pressed(pixelgl.KeyDown) {
			cameraPosition.Y -= CAMERA_SPEED * timeElapsed
		}

		if window.Pressed(pixelgl.KeyUp) {
			cameraPosition.Y += CAMERA_SPEED * timeElapsed
		}

		if window.Pressed(pixelgl.KeyPeriod) {
			cameraZoom += ZOOM_SPEED * timeElapsed
		}

		if window.Pressed(pixelgl.KeyComma) {
			cameraZoom -= ZOOM_SPEED * timeElapsed
		}

		if cameraZoom < 0 {
			cameraZoom = 0
		}

		viewportScale := window.Bounds().H() / g.Bounds.H()
		camera := pixel.IM.Scaled(window.Bounds().Min, viewportScale).Moved(window.Bounds().Center().Sub(cameraPosition)).Scaled(window.Bounds().Center(), cameraZoom)
		window.SetMatrix(camera)

		window.Clear(colornames.White)

		g.Draw()

		// untransform so fps counter appears in bottom-left of viewport regardless of pan/zoom
		window.SetMatrix(pixel.IM)
		logText.Draw(window, pixel.IM)

		window.Update()

		frames++
		select {
		case <-timer:
			logText.Clear()
			fmt.Fprintf(logText, "%dfps", frames)
			frames = 0
		default:
		}
	}
}

func main() {
	pixelgl.Run(entry)
}
