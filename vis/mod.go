// Graph and graph rendering

package vis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"

	"github.com/gonum/graph"
	"github.com/gonum/graph/simple"
)

// Extends graph.Node with necessary properties
type PositionedNode struct {
	Id       int
	Name     string
	Selected bool // Used only for map editor
	Weight   int  // How difficult it is to attack this vertex

	// Store the name pre-rendered
	renderedName *text.Text

	People  []*Person
	Zombies []*Zombie
	Items   []Item

	Pos pixel.Vec
}

// Necessary to implement graph.Node
func (n PositionedNode) ID() int {
	return n.Id
}

// Pre-render the vertex labels
func (n *PositionedNode) RenderName(atlas *text.Atlas) {
	n.renderedName = text.New(n.Pos, atlas)
	n.renderedName.Color = colornames.Black
	// Write to a buffer so that the size can be checked for centering purposes
	var w bytes.Buffer
	w.WriteString(n.Name)
	if len(n.Items) > 0 {
		// Prevent double-prinitng of item duplicates.
		seen := make([]bool, N_ITEMS)
		seen[n.Items[0]] = true
		w.WriteString(fmt.Sprintf(" (%s", n.Items[0]))
		for _, i := range n.Items[1:] {
			if !seen[i] {
				w.WriteString(fmt.Sprintf(", %s", i))
				seen[i] = true
			}
		}
		w.WriteString(")")
	}
	s := w.String()
	// Center text
	n.renderedName.Dot.X -= n.renderedName.BoundsOf(s).W() / 2
	fmt.Fprintln(n.renderedName, s)
	n.renderedName.Dot.X -= n.renderedName.BoundsOf(strconv.Itoa(n.Weight)).W() / 2
	fmt.Fprintln(n.renderedName, n.Weight)
}

// Encapsulates graphics handles and such.
type VWindow struct {
	window     *pixelgl.Window
	draw       *imdraw.IMDraw
	atlas      *text.Atlas
	StatusText *text.Text
	Log        chan string

	VertexSize float64

	Graph *MapGraph
}

func NewVWindow(window *pixelgl.Window, draw *imdraw.IMDraw, statusAtlas *text.Atlas, labelAtlas *text.Atlas, bounds pixel.Rect) *VWindow {
	t := text.New(pixel.V(10, window.Bounds().H()-20), statusAtlas)
	t.Color = colornames.Black
	t.WriteString("Welcome!\n")
	t.WriteString("Press 's' to see the status of all people currently alive.\n")
	t.WriteString("Press 'k' to see a key detailing the shortened item names.\n")
	t.WriteString("Press 'c' to clear the status text.\n")
	t.WriteString("Use the arrow keys to move the camera, the '.' key to zoom, and the ',' key to zoom out.\n")

	return &VWindow{window, draw, statusAtlas, t, make(chan string), (25.0 / bounds.W()) * window.Bounds().H(), NewMapGraph(labelAtlas, bounds)}
}

// Implements io.Writer for VWindow, allowing fmt.Println(w, ...) & co.
func (w *VWindow) Write(p []byte) (int, error) {
	fmt.Fprint(w.StatusText, string(p))
	return len(p), nil
}

// Extends simple.Undirected graph, adding display-related things
type MapGraph struct {
	*simple.UndirectedGraph

	atlas  *text.Atlas
	Bounds pixel.Rect

	entities uint

	Mutex *sync.Mutex

	Changed bool
}

func NewMapGraph(atlas *text.Atlas, bounds pixel.Rect) *MapGraph {
	return &MapGraph{simple.NewUndirectedGraph(0, -1), atlas, bounds, 0, &sync.Mutex{}, true}
}

func (g *MapGraph) NewPositionedNode(name string, x float64, y float64, w int) *PositionedNode {
	n := &PositionedNode{g.UndirectedGraph.NewNodeID(), name, false, w, nil, make([]*Person, 5), make([]*Zombie, 5), make([]Item, 0, 2), pixel.V(x, y)}
	n.RenderName(g.atlas)
	return n
}

// Because writing Go is marginally easier than writing JSON.
/*func (g *MapGraph) PopulateMap() {
	for _, v := range g.Nodes() {
		v.People = make([]*Person, 0, 5)
	}
	g.entities = 0
	rand.Seed(time.Now().UnixNano())
	g.AddPerson(Engineer, g.GetVertexByName("Water Treatment Plant"))

	g.AddPerson(Engineer, g.GetVertexByName("Garage"))

	g.AddPerson(Other, g.GetVertexByName("General Store"))
	g.AddPerson(Other, g.GetVertexByName("General Store"))

	g.AddPerson(Priest, g.GetVertexByName("Church"))

	g.AddPerson(Other, g.GetVertexByName("Hardware Store"))
	g.AddPerson(Other, g.GetVertexByName("Hardware Store"))

	g.AddPerson(Police, g.GetVertexByName("Police Station"))
	g.AddPerson(Police, g.GetVertexByName("Police Station"))

	g.AddPerson(Other, g.GetVertexByName("Store 1"))
	g.AddPerson(Other, g.GetVertexByName("Store 2"))
	g.AddPerson(Other, g.GetVertexByName("Store 3"))

	g.AddPerson(Other, g.GetVertexByName("Restaurant 1"))
	g.AddPerson(Other, g.GetVertexByName("Restaurant 1"))

	g.AddPerson(Other, g.GetVertexByName("Restaurant 2"))
	g.AddPerson(Other, g.GetVertexByName("Restaurant 2"))

	g.AddPerson(Other, g.GetVertexByName("Gas Station"))
	g.AddPerson(Other, g.GetVertexByName("Convenience Store"))

	g.AddPerson(Engineer, g.GetVertexByName("Warehouse 1"))
	g.AddPerson(Other, g.GetVertexByName("Warehouse 1"))
	g.AddPerson(Other, g.GetVertexByName("Warehouse 1"))

	g.AddPerson(Other, g.GetVertexByName("Warehouse 2"))
	g.AddPerson(Other, g.GetVertexByName("Warehouse 2"))

	g.AddPerson(Other, g.GetVertexByName("Warehouse 3"))
	g.AddPerson(Other, g.GetVertexByName("Warehouse 3"))
	g.AddPerson(Engineer, g.GetVertexByName("Warehouse 3"))

	g.AddPerson(Doctor, g.GetVertexByName("Doctor's Office"))
	g.AddPerson(Other, g.GetVertexByName("Doctor's Office"))

	g.AddPerson(Soldier, g.GetVertexByName("Trailer 12"))

	g.AddPerson(Firefighter, g.GetVertexByName("Fire Station"))
	g.AddPerson(Firefighter, g.GetVertexByName("Fire Station"))

	g.AddPerson(Soldier, g.GetVertexByName("House 3"))

	for i := 1; i <= 15; i++ {
		if rand.Intn(3) == 0 {
			continue
		}
		c := rand.Intn(3) + 1
		for j := 0; j < c; j++ {
			p := Profession(rand.Intn(13))
			if p > Other {
				p = Other
			}
			g.AddPerson(p, g.GetVertexByName(fmt.Sprintf("House %d", i)))
		}
	}

	for i := 1; i <= 16; i++ {
		if rand.Intn(2) == 0 {
			continue
		}
		c := rand.Intn(2) + 1
		for j := 0; j < c; j++ {
			p := Profession(rand.Intn(20))
			if p > Other {
				p = Other
			}
			g.AddPerson(p, g.GetVertexByName(fmt.Sprintf("Trailer %d", i)))
		}
	}

	for i := 1; i <= 15; i++ {
		p := Profession(rand.Intn(20))
		if p > Other {
			p = Other
		}
		pos := g.Nodes()[rand.Intn(len(g.Nodes()))]
		for ; len(pos.People) >= 2; pos = g.Nodes()[rand.Intn(len(g.Nodes()))] {
		}
		g.AddPerson(p, pos)
	}

	fmt.Println(g.entities)

	for _, v := range g.Nodes() {
		for _, p := range v.People {
			fmt.Printf("%s is at %s with %s.\n", p.Profession, g.Nodes()[p.Location].Name, p.Items)
		}
	}
}*/

// Show a status string for all entities
func (w *VWindow) PrintEntityStatus() {
	//	w.StatusText.Clear()
	for _, v := range w.Graph.Nodes() {
		for _, p := range v.People {
			fmt.Fprintf(w, "%d (%s) is at %s with ", p.ID, p.Profession, strings.ToUpper(v.Name))
			if len(p.Items) > 0 {
				fmt.Fprint(w, p.Items[0].StringLong())
				for _, i := range p.Items[1:] {
					fmt.Fprintf(w, ", %s", i.StringLong())
				}
			} else {
				fmt.Fprint(w, "NOTHING")
			}
			fmt.Fprintln(w)
		}
	}
}

// Add a new person to a vertex
func (g *MapGraph) AddPerson(job Profession, vertex *PositionedNode) {
	g.Mutex.Lock()
	vertex.People = append(vertex.People, NewPerson(g.entities, job, vertex.ID()))
	g.entities++
	g.Mutex.Lock()
}

func (g *MapGraph) GetVertexByName(name string) *PositionedNode {
	for _, v := range g.Nodes() {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (g *MapGraph) AddNode(n graph.Node) {
	g.Mutex.Lock()
	g.UndirectedGraph.AddNode(n)
	// Allows re-rendering only when the graph is actually Changed
	g.Changed = true
	g.Mutex.Unlock()
}

/*
** Return a vertex, asserting that it is a PositionedNode
** because the Go type system is elegant and well-designed
 */
func (g *MapGraph) Node(id int) *PositionedNode {
	return g.UndirectedGraph.Node(id).(*PositionedNode)
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

/*
** Returns edges, asserting that they are all concretely typed
** Necessary for nice serialization
 */
func (g *MapGraph) Edges() []*simple.Edge {
	edges := g.UndirectedGraph.Edges()
	ret := make([]*simple.Edge, len(edges))
	for i, n := range edges {
		ret[i] = n.(*simple.Edge)
	}

	return ret
}

func (g *MapGraph) AddEdge(from *PositionedNode, to *PositionedNode, weight float64) {
	g.Mutex.Lock()
	g.SetEdge(simple.Edge{simple.Node(from.ID()), simple.Node(to.ID()), weight})
	g.Mutex.Unlock()
}

func (w *VWindow) Draw() {
	if w.Graph.Changed {
		w.draw.Reset()
		w.draw.Clear()
		w.draw.Color = colornames.Lightslategray
		for _, n := range w.Graph.Nodes() {
			// Draw vertex
			if len(n.Zombies) > 0 {
				w.draw.Color = colornames.Red
				w.draw.Push(n.Pos)
				w.draw.Circle(w.VertexSize+2, float64(2*len(n.Zombies)))
				w.draw.Color = colornames.Lightslategray
			}
			if len(n.People) > 0 {
				w.draw.Color = colornames.Green
				w.draw.Push(n.Pos)
				w.draw.Circle(w.VertexSize+4, float64(2*len(n.People)))
				w.draw.Color = colornames.Lightslategray
			}
			w.draw.Push(n.Pos)
			if n.Selected {
				w.draw.Circle(w.VertexSize, 4)
			} else {
				w.draw.Circle(w.VertexSize, 0)
			}

			// Draw edges from that vertex
			for _, t := range w.Graph.From(n) {
				t := t.(*PositionedNode)
				w.draw.Push(n.Pos)
				w.draw.Push(t.Pos)
				weight, _ := w.Graph.Weight(n, t)
				w.draw.Line(weight * 2)
			}
		}

		w.draw.Push(pixel.V(0, 0))
		w.draw.Push(pixel.V(w.Graph.Bounds.W(), w.Graph.Bounds.H()))
		w.draw.Rectangle(2)

		w.Graph.Changed = false
	}

	w.draw.Draw(w.window)

	// Draw vertex names
	for _, n := range w.Graph.Nodes() {
		n.renderedName.Draw(w.window, pixel.IM)
	}
}

/*
** Go's JSON library and type system conspire to make
** it impossible to serialize anything unexported
** as well as impossible to deserialize anything
** to a field without a concrete type.
** Hence, these monstrosities.
 */

/*
** An intermediate type to allow easier serialization
** and deserialization of the embedded UndirectedGraph.
 */
type intermediateUndirectedGraph struct {
	Nodes []*PositionedNode
	Edges []*simple.Edge
}

// Wow it's actually not ridiculous
func (g *MapGraph) Serialize() ([]byte, error) {
	return json.Marshal(struct {
		G *MapGraph
		U intermediateUndirectedGraph
	}{g, intermediateUndirectedGraph{g.Nodes(), g.Edges()}})
}

// This is a disaster but there genuinely seems to be no real good way to do it.
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
