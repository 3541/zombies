package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"

	"golang.org/x/image/colornames"
	"golang.org/x/image/font"

	_ "image/png"

	"github.com/golang/freetype/truetype"

	"github.com/3541/zombies/behavior"
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

func loadImage(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	image, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return pixel.PictureDataFromImage(image), nil
}

func entry() {
	monitor := pixelgl.PrimaryMonitor()
	width, height := monitor.Size()

	config := pixelgl.WindowConfig{
		Title:  "Black Friday Simulator 2019",
		Bounds: pixel.R(0, 0, width, height),
		VSync:  true,
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

	consolasScaled, err := loadFont("consola.ttf", 14)
	if err != nil {
		panic(err)
	}

	/*	mapImage, err := loadImage("vantage.png")
		if err != nil {
			panic(err)
		}*/
	//mapSprite := pixel.NewSprite(mapImage, mapImage.Bounds())

	w := vis.NewVWindow(window, draw, text.NewAtlas(consolasScaled, text.ASCII), text.NewAtlas(consolas, text.ASCII), pixel.R(0, 0, 1000, 1000))
	go behavior.Start(w.Log, w.Graph)

	// Load and parse the map
	s, _ := ioutil.ReadFile("./map.json")

	err = w.Graph.Deserialize(s)
	if err != nil {
		panic(err)
	}

	// Start the map editor when running a debug build (see edit_release.go and edit_debug.go)
	editInit(window, w.Graph, consolasScaled)

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

	//	g.PopulateMap()

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

		if window.JustPressed(pixelgl.KeyS) {
			w.PrintEntityStatus()
		}

		if window.JustPressed(pixelgl.KeyC) {
			w.StatusText.Clear()
		}

		if window.JustPressed(pixelgl.KeyK) {
			fmt.Fprintln(w, "CS: CHAINSAW")
			fmt.Fprintln(w, "PSTL: PISTOL")
			fmt.Fprintln(w, "RFL: RIFLE")
			fmt.Fprintln(w, "EB: ENERGY BAR")
			fmt.Fprintln(w, "WS: WATER SOURCE")
			fmt.Fprintln(w, "WB: WATER BOTTLE")
			fmt.Fprintln(w, "RP: RUSTY PIPE")
			fmt.Fprintln(w, "HTCHT: HATCHET")
			fmt.Fprintln(w, "IAF: IMPROVISED AEROSOL FLAMETHROWER")
			fmt.Fprintln(w, "BDG: BANDAGE")
			fmt.Fprintln(w, "WRNC: WRENCH")
			fmt.Fprintln(w, "HS: HACKSAW")
			fmt.Fprintln(w, "RPG: ROCKET-PROPELLED GRENADE LAUNCHER")
			fmt.Fprintln(w, "ATGM: ANTI-TANK GUIDED MISSILE")
			fmt.Fprintln(w, "HW: HOLY WATER")
		}

		// Scale viewport to match height of map space
		viewportScale := window.Bounds().H() / w.Graph.Bounds.H()
		camera := pixel.IM.Scaled(window.Bounds().Min, viewportScale).Moved(window.Bounds().Center().Sub(cameraPosition)).Scaled(window.Bounds().Center(), cameraZoom)
		window.SetMatrix(camera)

		window.Clear(colornames.White)

		//	mapSprite.Draw(window, pixel.IM.Moved(mapImage.Bounds().Center()))
		w.Draw()

		// untransform so fps counter appears in bottom-left of viewport regardless of pan/zoom
		window.SetMatrix(pixel.IM)
		logText.Draw(window, pixel.IM)
		w.StatusText.Draw(window, pixel.IM)

		// Do map editor things, if in a debug build
		editGraph(camera)

		select {
		case t := <-w.Log:
			fmt.Fprintln(w, t)
		default:
		}

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
	editEnd(window, w.Graph)
}

func main() {
	pixelgl.Run(entry)
}
