// Stub implementations of map editor functions. For release builds.

package main

import (
	"math"

	"github.com/3541/zombies/entity"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/font"
)

var window *pixelgl.Window
var graph *entity.MapGraph

func editGraph(camera pixel.Matrix) {
	// Add zombies on mouse click
	if window.JustPressed(pixelgl.MouseButtonLeft) {
		clicked := clickedVertex(camera.Unproject(window.MousePosition()))
		// If the vertex already has people, convert them all to zombies
		if len(clickedVertex.People) > 0 {
		} else {

		}
	}
}

func editInit(w *pixelgl.Window, g *entity.MapGraph, font font.Face) {
	window = w
	graph = g
}

func editEnd(window *pixelgl.Window, g *entity.MapGraph) {}

func clickedVertex(pos pixel.Vec) *entity.PositionedNode {
	for _, v := range graph.Nodes() {
		if math.Sqrt(math.Pow(pos.X-v.Pos.X, 2)+math.Pow(pos.Y-v.Pos.Y, 2)) < graph.VertexSize {
			return v
		}
	}
	return nil
}
