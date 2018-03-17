package vis

import (
	"encoding/json"
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
	Id   int
	Name string

	renderedName *text.Text

	Pos pixel.Vec
}

func (n PositionedNode) ID() int {
	return n.Id
}

func (n *PositionedNode) RenderName(atlas *text.Atlas) {
	n.renderedName = text.New(n.Pos, atlas)
	n.renderedName.Color = colornames.Black
	fmt.Fprint(n.renderedName, n.Name)
	n.renderedName.Orig = n.renderedName.Orig.Sub(n.renderedName.Dot.Sub(n.renderedName.Orig).ScaledXY(pixel.V(0.5, 1)))
	fmt.Fprintln(n.renderedName)
	n.renderedName.Orig = n.renderedName.Orig.Add(n.renderedName.Dot.Sub(n.renderedName.Orig).ScaledXY(pixel.V(0, 0.25)))
	n.renderedName.Clear()
	fmt.Fprint(n.renderedName, n.Name)
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
	return &MapGraph{simple.NewUndirectedGraph(0, -1), window, draw, atlas, bounds, true}
}

func (g *MapGraph) NewPositionedNode(name string, x float64, y float64) *PositionedNode {
	n := &PositionedNode{g.UndirectedGraph.NewNodeID(), name, nil, pixel.V(x, y)}
	n.RenderName(g.atlas)
	return n
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

func (g *MapGraph) AddEdge(from *PositionedNode, to *PositionedNode, weight float64) {
	g.SetEdge(simple.Edge{simple.Node(from.ID()), simple.Node(to.ID()), weight})
}

func (g *MapGraph) Draw() {
	if g.changed {
		g.draw.Reset()
		g.draw.Color = colornames.Lightslategray
		for _, n := range g.Nodes() {
			g.draw.Push(n.Pos)
			g.draw.Circle(20, 0)

			for _, t := range g.From(n) {
				t := t.(*PositionedNode)
				g.draw.Push(n.Pos)
				g.draw.Push(t.Pos)
				w, _ := g.Weight(n, t)
				g.draw.Line(w)
			}
		}

		g.draw.Push(pixel.V(0, 0))
		g.draw.Push(pixel.V(1000, 1000))
		g.draw.Rectangle(2)

		g.changed = false
	}

	g.draw.Draw(g.window)

	for _, n := range g.Nodes() {
		n.renderedName.Draw(g.window, pixel.IM)
	}
}

func (g *MapGraph) Serialize() ([]byte, error) {
	return json.Marshal(g.UndirectedGraph)
}

func (g *MapGraph) Deserialize(data []byte) error {
	err := json.Unmarshal(data, &g.UndirectedGraph)
	if err != nil {
		return err
	}

	d := make(map[string]*json.RawMessage)
	err = json.Unmarshal(data, &d)
	if err != nil {
		return err
	}

	serNodes := make(map[int]*json.RawMessage)
	err = json.Unmarshal(*d["Nodes"], &serNodes)
	if err != nil {
		return err
	}

	for _, v := range serNodes {
		n := new(PositionedNode)
		err = json.Unmarshal(*v, &n)
		if err != nil {
			return err
		}
		n.RenderName(g.atlas)
		g.AddNode(n)
	}

	serEdges := make(map[int]*json.RawMessage)
	err = json.Unmarshal(*d["Edges"], &serEdges)
	if err != nil {
		return err
	}

	for _, v := range serEdges {
		se := make(map[string]*json.RawMessage)
		err = json.Unmarshal(*v, &se)
		if err != nil {
			return err
		}

		for _, v := range se {
			e := new(simple.Edge)
			sv := make(map[string]float64)
			err = json.Unmarshal(*v, &sv)
			if err != nil {
				return err
			}

			e.F = simple.Node(sv["F"])
			e.T = simple.Node(sv["T"])
			e.W = sv["W"]

			if !g.HasEdgeBetween(e.F, e.T) {
				g.SetEdge(e)
			}
		}
	}

	return nil
}
