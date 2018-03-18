package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"

	"golang.org/x/image/colornames"
	"golang.org/x/image/font"

	"github.com/golang/freetype/truetype"

	"github.com/3541/zombies/vis"
)

const (
	CAMERA_SPEED = 600.0
	ZOOM_SPEED   = 1.01
)

func loadFont(path string, size float64) (font.Face, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	font, err := truetype.Parse(data)
	if err != nil {
		return nil, err
	}

	return truetype.NewFace(font, &truetype.Options{Size: size, GlyphCacheEntries: 1}), nil
}

func entry() {
	monitor := pixelgl.PrimaryMonitor()
	width, height := monitor.Size()

	config := pixelgl.WindowConfig{
		Title:  "Apocalypse Simulator 2018",
		Bounds: pixel.R(0, 0, width, height),
		//VSync:  true,
	}

	window, err := pixelgl.NewWindow(config)
	if err != nil {
		panic(err)
	}

	window.SetSmooth(true)

	// Enable fullscreen
	window.SetMonitor(monitor)

	// Shape drawing interface
	draw := imdraw.New(nil)
	draw.Color = colornames.Black

	consolas, err := loadFont("consola.ttf", 24)
	if err != nil {
		panic(err)
	}

	consolasSmall, err := loadFont("consola.ttf", 16)
	if err != nil {
		panic(err)
	}

	g := vis.NewMapGraph(window, draw, text.NewAtlas(consolasSmall, text.ASCII), pixel.R(0, 0, 1000, 1000))

	/*	n1 := g.NewPositionedNode("BUILDING 1", 500, 500)
		g.AddNode(n1)
		n2 := g.NewPositionedNode("BUILDING 2", 800, 600)
		g.AddNode(n2)
		g.AddEdge(n1, n2, 3)

		s, err := g.Serialize()
		if err != nil {
			panic(err)
		}

		err = ioutil.WriteFile("./test_map.json", s, 0666)
		if err != nil {
			panic(err)
		}*/

	// Load and parse the map
	s, _ := ioutil.ReadFile("./map.json")

	err = g.Deserialize(s)
	if err != nil {
		panic(err)
	}

	// Start the map editor when running a debug build (see edit_release.go and edit_debug.go)
	editInit(window, g, consolas)

	// To track framerate
	frames := 0
	timer := time.Tick(time.Second)

	// Used to print framerate to screen
	logAtlas := text.NewAtlas(consolas, text.ASCII)
	logText := text.New(pixel.V(10, 10), logAtlas)
	logText.Color = colornames.Black

	cameraPosition := window.Bounds().Center()
	cameraZoom := 1.0

	// Track time since last frame for constant-speed movements, even without VSync
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

		// Scale viewport to match height of map space
		viewportScale := window.Bounds().H() / g.Bounds.H()
		camera := pixel.IM.Scaled(window.Bounds().Min, viewportScale).Moved(window.Bounds().Center().Sub(cameraPosition)).Scaled(window.Bounds().Center(), cameraZoom)
		window.SetMatrix(camera)

		window.Clear(colornames.White)

		g.Draw()

		// untransform so fps counter appears in bottom-left of viewport regardless of pan/zoom
		window.SetMatrix(pixel.IM)
		logText.Draw(window, pixel.IM)

		// Do map editor things, if in a debug build
		editGraph(camera)

		window.Update()

		// Every second, display the frames rendered in that second
		frames++
		select {
		case <-timer:
			logText.Clear()
			fmt.Fprintf(logText, "%d", frames)
			frames = 0
		default:
		}
	}
	editEnd(window, g)
}

func main() {
	pixelgl.Run(entry)
}
