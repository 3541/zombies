package entity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"github.com/gonum/graph"
	"github.com/gonum/graph/simple"
	"golang.org/x/image/colornames"
)

// Extends graph.Node with necessary properties
type PositionedNode struct {
	Id       int
	Name     string
	Selected bool // Used only for map editor
	Weight   int  // How difficult it is to attack this vertex

	// Store the name pre-rendered
	RenderedName *text.Text

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
	n.RenderedName = text.New(n.Pos, atlas)
	n.RenderedName.Color = colornames.Black
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
	n.RenderedName.Dot.X -= n.RenderedName.BoundsOf(s).W() / 2
	fmt.Fprintln(n.RenderedName, s)
	n.RenderedName.Dot.X -= n.RenderedName.BoundsOf(strconv.Itoa(n.Weight)).W() / 2
	fmt.Fprintln(n.RenderedName, n.Weight)
}

// Extends simple.Undirected graph, adding display-related things
type MapGraph struct {
	*simple.UndirectedGraph

	atlas      *text.Atlas
	Bounds     pixel.Rect
	VertexSize float64

	entities uint

	Log chan string

	Mutex *sync.RWMutex

	Changed bool
}

func NewMapGraph(atlas *text.Atlas, bounds pixel.Rect, vertexSize float64) *MapGraph {
	return &MapGraph{simple.NewUndirectedGraph(0, -1), atlas, bounds, vertexSize, 0, make(chan string, 20), &sync.RWMutex{}, true}
}

func (g *MapGraph) NewPositionedNode(name string, x float64, y float64, w int) *PositionedNode {
	n := &PositionedNode{g.UndirectedGraph.NewNodeID(), name, false, w, nil, make([]*Person, 0, 5), make([]*Zombie, 0, 5), make([]Item, 0, 2), pixel.V(x, y)}
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

// Add a new person to a vertex
func (g *MapGraph) AddPerson(job Profession, vertex *PositionedNode) {
	g.Mutex.Lock()
	vertex.People = append(vertex.People, NewPerson(g.entities, job, vertex.ID()))
	g.entities++
	g.Mutex.Unlock()
}

func (g *MapGraph) AddNewZombie(vertex *PositionedNode) {
	z := NewZombie(g.entities, vertex.ID())
	g.Mutex.Lock()
	vertex.Zombies = append(vertex.Zombies, z)
	g.entities++
	g.Changed = true
	g.Mutex.Unlock()
	go z.Unlive(g)
}

func (g *MapGraph) InfectPerson(p *Person) {

	z := NewZombieFromPerson(p)

	p.Kill <- "INFECTED by ZOMBIE"

	g.Mutex.Lock()
	n := g.Node(p.Location)
	n.Zombies = append(n.Zombies, z)
	g.Changed = true
	g.Mutex.Unlock()

	go z.Unlive(g)

}

func (g *MapGraph) RemovePerson(p *Person) {
	g.Mutex.RLock()
	n := g.Node(p.Location)
	i := 0
	for _, v := range n.People {
		if v.Id == p.Id {
			break
		}
		i++
	}
	g.Mutex.RUnlock()
	g.Mutex.Lock()
	if len(n.People) > 1 {
		n.People = append(n.People[:i], n.People[i+1:]...)
	} else {
		n.People = n.People[:0]
	}
	g.Mutex.Unlock()
}

func (g *MapGraph) RemoveZombie(z *Zombie) {
	g.Mutex.RLock()
	n := g.Node(z.Location)
	i := 0
	for _, v := range n.Zombies {
		if v.Id == z.Id {
			break
		}
		i++
	}
	g.Mutex.RUnlock()
	g.Mutex.Lock()
	if len(n.Zombies) > 1 {
		n.Zombies = append(n.Zombies[:i], n.Zombies[i+1:]...)
	} else {
		n.Zombies = n.Zombies[:0]
	}
	g.Mutex.Unlock()
}
func (g *MapGraph) StartEntities() {
	for _, v := range g.Nodes() {
		for _, p := range v.People {
			go p.Live(g)
		}

		for _, z := range v.Zombies {
			go z.Unlive(g)
		}
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
