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
		if p.checkKilled(g) {
			return
		}

		if rand.Intn(100) == 1 {
			g.Mutex.RLock()
			n := g.From(g.Node(p.Location))
			t := n[rand.Intn(len(n))].(*PositionedNode)
			g.Mutex.RUnlock()
			// Humans only pay attention to edge weights because zombies are (presumably) too stupid to fortify
			pause(int(g.Edge(g.Node(p.Location), t).Weight()), time.Second)
			// Did it die before getting to the next vertex?
			if p.checkKilled(g) {
				return
			}
			g.Log <- fmt.Sprintf("%s moves from %s to %s", p.Profession, g.Node(p.Location).Name, t.Name)
			p.moveTo(g, t)
		}
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
	g.Mutex.RLock()
	pn := g.Node(p.Location)

	// Because Go actually doesn't implement this in the standard library
	i := 0
	for _, v := range pn.People {
		if p.Id == v.Id {
			break
		}
		i++
	}
	g.Mutex.RUnlock()
	g.Mutex.Lock()
	// Delete from old vertex
	if len(pn.People) > 1 {
		pn.People = append(pn.People[:i], pn.People[i+1:]...)
	} else {
		pn.People = pn.People[:0]
	}

	p.Location = t.ID()
	t.People = append(t.People, p)

	g.Changed = true
	g.Mutex.Unlock()
}

func (z *Zombie) Unlive(g *MapGraph) {
	z.Kill = make(chan string, 3)
	pause(rand.Intn(2000), time.Millisecond)
	tick := time.NewTicker(500 * time.Millisecond)
	for _ = range tick.C {
		if z.checkKilled(g) {
			return
		}

		if len(g.Node(z.Location).People) == 0 {
			t := z.nearestPersonTraverseFirstStep(g)
			fortification := 0
			if len(t.People) > 0 {
				fortification = t.Weight
			}
			pause(int(g.Edge(g.Node(z.Location), t).Weight())+fortification, time.Second)
			if z.checkKilled(g) {
				return
			}
			g.Log <- fmt.Sprintf("ZOMBIE moves from %s to %s", g.Node(z.Location).Name, t.Name)
			z.moveTo(g, t)
		}
	}
}

func (z *Zombie) moveTo(g *MapGraph, t *PositionedNode) {
	g.Mutex.RLock()
	pn := g.Node(z.Location)

	// Because Go actually doesn't implement this in the standard library
	i := 0
	for _, v := range pn.Zombies {
		if z.Id == v.Id {
			break
		}
		i++
	}
	g.Mutex.RUnlock()
	g.Mutex.Lock()
	// Delete from old vertex
	if len(pn.Zombies) > 1 {
		pn.Zombies = append(pn.Zombies[:i], pn.Zombies[i+1:]...)
	} else {
		pn.Zombies = pn.Zombies[:0]
	}

	z.Location = t.ID()
	t.Zombies = append(t.Zombies, z)

	g.Changed = true
	g.Mutex.Unlock()
}

// Returns the first vertex on a path towards the nearest person by traversal (Dijkstra's Shortest Path)
func (z *Zombie) nearestPersonTraverseFirstStep(g *MapGraph) *PositionedNode {
	g.Mutex.RLock()
	unvisited := g.Nodes()
	distance := make(map[int]uint)
	previous := make(map[int]int)
	for _, v := range unvisited {
		distance[v.ID()] = ^uint(0)
	}

	distance[z.Location] = 0

	// Find the nearest person
	var nearestPerson *PositionedNode
	for len(unvisited) > 0 {
		min := ^uint(0) // uint max value
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
	g.Mutex.RUnlock()
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
