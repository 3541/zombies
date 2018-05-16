package entity

import (
	"fmt"
	"math/rand"
	"time"
)

func (p *Person) Live(g *MapGraph) {
	tick := time.NewTicker(500 * time.Millisecond)
	for _ = range tick.C {
		//		g.Log <- fmt.Sprintf("%s is at %s with %v", p.Profession, g.Node(p.Location).Name, p.Items)

		if p.checkKilled(g) {
			return
		}

		if rand.Intn(100) == 1 {
			n := g.From(g.Node(p.Location))
			t := n[rand.Intn(len(n))].(*PositionedNode)
			g.Log <- fmt.Sprintf("%s moves from %s to %s", p.Profession, g.Node(p.Location).Name, t.Name)
			p.moveTo(g, t)
		}
	}
}

func (p *Person) checkKilled(g *MapGraph) bool {
	select {
	case reason := <-p.Kill:
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

func (z *Zombie) UnLive(g *MapGraph) {

}
