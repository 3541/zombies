// +build !debug

// Stub implementations of map editor functions. For release builds.

package main

import (
	"github.com/faiface/pixel/pixelgl"

	"github.com/3541/zombies/vis"
)

func editGraph(window *pixelgl.Window, g *vis.MapGraph) {
}

func editInit(window *pixelgl.Window) {
}

func editEnd(g *vis.MapGraph) {}
