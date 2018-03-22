// +build !debug

// Stub implementations of map editor functions. For release builds.

package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/font"

	"github.com/3541/zombies/vis"
)

func editGraph(camera pixel.Matrix) {
}

func editInit(window *pixelgl.Window, g *vis.MapGraph, font font.Face) {
}

func editEnd(window *pixelgl.Window, g *vis.MapGraph) {}
