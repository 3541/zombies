package vis

import (
	"fmt"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"

	"github.com/gonum/graph"
	"github.com/gonum/graph/simple"
)

type PositionedNode struct {
	id   int
	name string

	renderedName *text.Text

	x float64
	y float64
}

func (n PositionedNode) ID() int {
	return n.id
}

type MapGraph struct {
	*simple.UndirectedGraph

	window *pixelgl.Window
	draw   *imdraw.IMDraw
	atlas  *text.Atlas
	Bounds pixel.Rect

	changed bool
}

func NewMapGraph(window *pixelgl.Window, draw *imdraw.IMDraw, atlas *text.Atlas, bounds pixel.Rect) *MapGraph {
	return &MapGraph{simple.NewUndirectedGraph(0, -1), window, draw, atlas, bounds, false}
}

func (g *MapGraph) NewPositionedNode(name string, x float64, y float64) *PositionedNode {
	renderedName := text.New(pixel.V(200, 200), g.atlas)
	renderedName.Color = colornames.Black
	fmt.Fprintln(renderedName, name)
	return &PositionedNode{g.UndirectedGraph.NewNodeID(), name, renderedName, x, y}
}

func (g *MapGraph) AddNode(n graph.Node) {
	g.UndirectedGraph.AddNode(n)
	g.changed = true
}

func (g *MapGraph) Nodes() []*PositionedNode {
	nodes := g.UndirectedGraph.Nodes()
	ret := make([]*PositionedNode, len(nodes))
	for i, n := range nodes {
		ret[i] = n.(*PositionedNode)
	}

	return ret
}

func (g *MapGraph) Draw() {
	if g.changed {
		g.draw.Reset()
		g.draw.Color = pixel.RGB(0, 0, 0)
		for _, n := range g.Nodes() {
			g.draw.Push(pixel.V(n.x, n.y))
		}

		g.draw.Circle(20, 0)

		g.draw.Push(pixel.V(0, 0))
		g.draw.Push(pixel.V(1000, 1000))
		g.draw.Rectangle(2)

		g.changed = false
	}

	for _, n := range g.Nodes() {
		n.renderedName.Draw(g.window, pixel.IM)
	}
}
