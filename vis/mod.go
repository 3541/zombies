// Graph and graph rendering

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
	Id       int
	Name     string
	Selected bool

	// Store the name pre-rendered
	renderedName *text.Text

	Pos pixel.Vec
}

// Necessary to implement graph.Node
func (n PositionedNode) ID() int {
	return n.Id
}

// Yes, this is awful, I know.
// Render the text multiple times in order to reposition to center on the vertex.
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

// Extends simple.Undirected graph, adding handles to graphics things
type MapGraph struct {
	*simple.UndirectedGraph

	window *pixelgl.Window
	draw   *imdraw.IMDraw
	atlas  *text.Atlas
	Bounds pixel.Rect

	VertexSize float64

	Changed bool
}

func NewMapGraph(window *pixelgl.Window, draw *imdraw.IMDraw, atlas *text.Atlas, bounds pixel.Rect) *MapGraph {
	return &MapGraph{simple.NewUndirectedGraph(0, -1), window, draw, atlas, bounds, (40.0 / 3840) * window.Bounds().H(), true}
}

func (g *MapGraph) NewPositionedNode(name string, x float64, y float64) *PositionedNode {
	n := &PositionedNode{g.UndirectedGraph.NewNodeID(), name, false, nil, pixel.V(x, y)}
	n.RenderName(g.atlas)
	return n
}

// Allows re-rendering only when the graph is actually Changed
func (g *MapGraph) AddNode(n graph.Node) {
	g.UndirectedGraph.AddNode(n)
	g.Changed = true
}

// Returns vertices, asserting they are all PositionedNodes
func (g *MapGraph) Nodes() []*PositionedNode {
	nodes := g.UndirectedGraph.Nodes()
	ret := make([]*PositionedNode, len(nodes))
	for i, n := range nodes {
		ret[i] = n.(*PositionedNode)
	}

	return ret
}

// Returns edges, asserting that they are all concretely typed
// Necessary for nice serialization
func (g *MapGraph) Edges() []*simple.Edge {
	edges := g.UndirectedGraph.Edges()
	ret := make([]*simple.Edge, len(edges))
	for i, n := range edges {
		ret[i] = n.(*simple.Edge)
	}

	return ret
}

func (g *MapGraph) AddEdge(from *PositionedNode, to *PositionedNode, weight float64) {
	g.SetEdge(simple.Edge{simple.Node(from.ID()), simple.Node(to.ID()), weight})
}

func (g *MapGraph) Draw() {
	if g.Changed {
		g.draw.Reset()
		g.draw.Clear()
		g.draw.Color = colornames.Red
		for _, n := range g.Nodes() {
			// Draw vertex
			g.draw.Push(n.Pos)
			if n.Selected {
				g.draw.Circle(g.VertexSize, 4)
			} else {
				g.draw.Circle(g.VertexSize, 0)
			}

			// Draw edges from that vertex
			for _, t := range g.From(n) {
				t := t.(*PositionedNode)
				g.draw.Push(n.Pos)
				g.draw.Push(t.Pos)
				w, _ := g.Weight(n, t)
				g.draw.Line(w)
			}
		}

		g.draw.Push(pixel.V(0, 0))
		g.draw.Push(pixel.V(g.Bounds.W(), g.Bounds.H()))
		g.draw.Rectangle(2)

		g.Changed = false
	}

	g.draw.Draw(g.window)

	// Draw vertex names
	for _, n := range g.Nodes() {
		n.renderedName.Draw(g.window, pixel.IM)
	}
}

// Go's JSON library and type system conspire to make
// it impossible to serialize anything unexported
// as well as impossible to deserialize anything
// to a field without a concrete type.
// Hence, these monstrosities.

// An intermediate type to allow easier serialization
// and deserialization of the embedded UndirectedGraph.
type intermediateUndirectedGraph struct {
	Nodes []*PositionedNode
	Edges []*simple.Edge
}

func (g *MapGraph) Serialize() ([]byte, error) {
	return json.Marshal(struct {
		G *MapGraph
		U intermediateUndirectedGraph
	}{g, intermediateUndirectedGraph{g.Nodes(), g.Edges()}})
}

func (g *MapGraph) Deserialize(data []byte) error {
	p := make(map[string]json.RawMessage)
	err := json.Unmarshal(data, &p)
	if err != nil {
		return err
	}

	err = json.Unmarshal(p["G"], &g)
	if err != nil {
		return err
	}

	iug := new(intermediateUndirectedGraph)
	serIug := make(map[string]json.RawMessage)
	err = json.Unmarshal(p["U"], &serIug)
	if err != nil {
		return err
	}

	err = json.Unmarshal(serIug["Nodes"], &iug.Nodes)
	if err != nil {
		return err
	}

	for _, v := range iug.Nodes {
		v.RenderName(g.atlas)
		g.AddNode(v)
	}

	serEdges := make([]json.RawMessage, len(iug.Nodes)*2)
	err = json.Unmarshal(serIug["Edges"], &serEdges)
	if err != nil {
		return err
	}

	for _, se := range serEdges {
		em := make(map[string]float64)
		err = json.Unmarshal(se, &em)
		if err != nil {
			return err
		}

		e := new(simple.Edge)
		e.F = simple.Node(em["F"])
		e.T = simple.Node(em["T"])
		e.W = em["W"]

		if !g.HasEdgeBetween(e.F, e.T) {
			g.SetEdge(e)
		}
	}

	return nil
}
