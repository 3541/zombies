// +build debug

// Graph editing. For debug builds

package main

import (
	"fmt"

	"github.com/3541/zombies/vis"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var statusAtlas *text.Atlas
var statusText *text.Text

func editGraph(window *pixelgl.Window, g *vis.MapGraph, camera pixel.Matrix) {
	statusText.Draw(window, pixel.IM.Scaled(statusText.Orig, 2))

	if window.JustPressed(pixelgl.MouseButtonLeft) {
		statusText.Clear()
		pos := camera.Unproject(window.MousePosition())
		fmt.Fprintln(statusText, pos)
	}
}

func editInit(window *pixelgl.Window) {
	statusAtlas = text.NewAtlas(basicfont.Face7x13, text.ASCII)
	statusText = text.New(pixel.V(10, window.Bounds().H()-50), statusAtlas)
	statusText.Color = colornames.Black

	fmt.Fprintln(statusText, "NOTE: This is a DEBUG build with the map editor enabled.")
	fmt.Fprintln(statusText, "Click to add a vertex at that position.")
	fmt.Fprintln(statusText, "Or press the 'a' key to add a vertex at a specific position.")
}
