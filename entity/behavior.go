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
	p.Kill = make(chan string, 5)
	pause(rand.Intn(1000), time.Millisecond)
	tick := time.NewTicker(500 * time.Millisecond)
	for _ = range tick.C {
		//		g.Log <- fmt.Sprintf("%s is at %s with %v", p.Profession, g.Node(p.Location).Name, p.Items)

		if p.checkKilled(g) {
			return
		}

		if rand.Intn(100) == 1 {
			n := g.From(g.Node(p.Location))
			t := n[rand.Intn(len(n))].(*PositionedNode)
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
	g.Mutex.Lock()
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
	g.Mutex.Unlock()
}

func (z *Zombie) UnLive(g *MapGraph) {

}
