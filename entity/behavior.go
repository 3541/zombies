package entity

import (
	"fmt"
	"math/rand"
	"time"
)

func (p *Person) Live(g *MapGraph) {
	tick := time.NewTicker(300 * time.Millisecond)
	for _ = range tick.C {
		//		g.Log <- fmt.Sprintf("%s is at %s with %v", p.Profession, g.Node(p.Location).Name, p.Items)
		if rand.Intn(50) == 1 {
			n := g.From(g.Node(p.Location))
			t := rand.Intn(len(n))
			g.Log <- fmt.Sprintf("%s moves from %s to %s", p.Profession, g.Node(p.Location).Name, g.Node(t).Name)
			p.moveTo(g, n[t].(*PositionedNode))
		}
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
	if len(pn.People) != 0 {
		pn.People = append(pn.People[:i], pn.People[i+1:]...)
	} else {
		pn.People = pn.People[:0]
	}

	t.People = append(t.People, p)
	p.Location = t.Id

	g.Changed = true
}

func (z *Zombie) UnLive(g *MapGraph) {

}
