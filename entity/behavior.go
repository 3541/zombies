package entity

import (
	"fmt"
	"math/rand"
	"time"
)

func pause(t int, unit time.Duration) {
	<-time.NewTimer(time.Duration(t) * unit).C
}

func (p *Person) Live(g *MapGraph) {
	p.Kill = make(chan string, 3)
	pause(rand.Intn(2000), time.Millisecond)
	tick := time.NewTicker(500 * time.Millisecond)
	for _ = range tick.C {
		//		g.Log <- fmt.Sprintf("%s is at %s with %v", p.Profession, g.Node(p.Location).Name, p.Items)
		g.Mutex.Lock()
		if p.checkKilled(g) {
			g.Mutex.Unlock()
			return
		}

		if rand.Intn(100) == 1 {
			n := g.From(g.Node(p.Location))
			t := n[rand.Intn(len(n))].(*PositionedNode)
			// Humans only pay attention to edge weights because zombies are (presumably) too stupid to fortify
			g.Mutex.Unlock()
			pause(int(g.Edge(g.Node(p.Location), t).Weight()), time.Second)
			g.Mutex.Lock()
			// Did it die before getting to the next vertex?
			if p.checkKilled(g) {
				g.Mutex.Unlock()
				return
			}
			g.Log <- fmt.Sprintf("%s moves from %s to %s", p.Profession, g.Node(p.Location).Name, t.Name)
			p.moveTo(g, t)
		}
		g.Mutex.Unlock()
	}
}

func (p *Person) checkKilled(g *MapGraph) bool {
	select {
	case reason := <-p.Kill:
		g.RemovePerson(p)
		g.Log <- fmt.Sprintf("%s %s at %s", p.Profession, reason, g.Node(p.Location).Name)
		return true
	default:
		return false
	}
}

func (p *Person) moveTo(g *MapGraph, t *PositionedNode) {
	pn := g.Node(p.Location)

	// Because Go actually doesn't implement this in the standard library
	i := 0
	for _, v := range pn.People {
		if p.Id == v.Id {
			break
		}
		i++
	}
	// Delete from old vertex
	if len(pn.People) > 1 {
		pn.People = append(pn.People[:i], pn.People[i+1:]...)
	} else {
		pn.People = pn.People[:0]
	}

	p.Location = t.ID()
	t.People = append(t.People, p)

	g.Changed = true
}

func (z *Zombie) Unlive(g *MapGraph) {
	z.Kill = make(chan string, 3)
	pause(rand.Intn(2000), time.Millisecond)
	tick := time.NewTicker(500 * time.Millisecond)
	for _ = range tick.C {
		g.Mutex.Lock()
		fmt.Println("Acquired lock")
		if z.checkKilled(g) {
			g.Mutex.Unlock()
			return
		}
		fmt.Println("Not killed")

		if len(g.Node(z.Location).People) == 0 {
			t := z.nearestPersonTraverseFirstStep(g)
			fmt.Println("acquired target")
			g.Log <- fmt.Sprintf("ZOMBIE moves from %s to %s", g.Node(z.Location).Name, t.Name)
			z.moveTo(g, t)
		}
		g.Mutex.Unlock()
		fmt.Println("Released lock")
	}
}

func (z *Zombie) moveTo(g *MapGraph, t *PositionedNode) {
	pn := g.Node(z.Location)

	// Because Go actually doesn't implement this in the standard library
	i := 0
	for _, v := range pn.Zombies {
		if z.Id == v.Id {
			break
		}
		i++
	}
	// Delete from old vertex
	if len(pn.Zombies) > 1 {
		pn.Zombies = append(pn.Zombies[:i], pn.Zombies[i+1:]...)
	} else {
		pn.Zombies = pn.Zombies[:0]
	}

	z.Location = t.ID()
	t.Zombies = append(t.Zombies, z)

	g.Changed = true
}

// Returns the first vertex on a path towards the nearest person by traversal (Dijkstra's Shortest Path)
func (z *Zombie) nearestPersonTraverseFirstStep(g *MapGraph) *PositionedNode {
	unvisited := g.Nodes()
	distance := make(map[int]uint)
	previous := make(map[int]int)
	for i := range distance {
		distance[i] = ^uint(0)
	}

	distance[z.Location] = 0

	// Find the nearest person
	var nearestPerson *PositionedNode
	for len(unvisited) > 0 {
		min := ^uint(0)
		var minvI int
		for i, v := range unvisited {
			if distance[v.ID()] < min {
				minvI = i
			}
		}

		v := unvisited[minvI]
		if len(v.People) > 0 {
			nearestPerson = v
			break
		}
		unvisited = append(unvisited[:minvI], unvisited[minvI+1:]...)

		for _, t := range g.From(v) {
			d := distance[v.ID()] + uint(g.Edge(v, t).Weight())
			if d < distance[t.ID()] {
				distance[t.ID()] = d
				previous[t.ID()] = v.ID()
			}
		}
	}

	// Work back to the first step on that path
	ret := nearestPerson
	for previous[ret.ID()] != z.Location {
		ret = g.Node(previous[ret.ID()])
	}
	return ret
}

func (z *Zombie) checkKilled(g *MapGraph) bool {
	select {
	case reason := <-z.Kill:
		g.RemoveZombie(z)
		g.Log <- fmt.Sprintf("ZOMBIE %s at %s", reason, g.Node(z.Location).Name)
		return true
	default:
		return false
	}
}
