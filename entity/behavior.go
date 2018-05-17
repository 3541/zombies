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
	p.Kill = make(chan string, 100)
	p.Damage = make(chan DamageMessage, 100)
	pause(rand.Intn(2000), time.Millisecond)
	tick := time.NewTicker(500 * time.Millisecond)
	for _ = range tick.C {
		if p.checkKilled(g) {
			return
		}

		//		g.Log <- fmt.Sprintf("%s is at %s with %v", p.Profession, g.Node(p.Location).Name, p.Items)
		g.Mutex.RLock()
		if p.Hunger >= 400 {
			p.Kill <- "STARVED to DEATH"
		} else if p.Thirst >= 300 {
			p.Kill <- "DIED of THIRST"
		}
		g.Mutex.RUnlock()

		select {
		case m := <-p.Damage:
			g.Mutex.Lock()
			p.Health -= int(m.Value)
			if p.Health <= 0 {
				p.Kill <- fmt.Sprintf("killed by %s with %s", m.Attacker, m.Item.StringLong())
			} else {
				g.Log <- fmt.Sprintf("%s at %s takes %d damage from %s wielding %s. Now at %d HP", p.Profession, g.Node(p.Location).Name, m.Value, m.Attacker, m.Item.StringLong(), p.Health)
			}
			g.Mutex.Unlock()
		default:
		}

		if p.checkKilled(g) {
			return
		}

		currentNode := g.Node(p.Location)

		g.Mutex.Lock()

		p.Hunger++
		p.Thirst++

		if len(currentNode.Zombies) > 0 {
			weapon := p.BestWeapon()
			minHealth := currentNode.Zombies[0].Health
			weakest := 0
			for i := range currentNode.Zombies {
				if currentNode.Zombies[i].Health < minHealth {
					minHealth = currentNode.Zombies[i].Health
					weakest = i
				}
			}
			currentNode.Zombies[weakest].Damage <- DamageMessage{weapon.Damage(), p.Profession.String(), weapon}
			if weapon.Consumable() {
				p.ConsumeItem(weapon)
			}
			g.Mutex.Unlock()
			continue
		}

		if p.Hunger >= 100 && p.Holding(EnergyBar) {
			p.ConsumeItem(EnergyBar)
			p.Hunger -= 100
			//			g.Log <- fmt.Sprintf("%s ate an ENERGY BAR at %s", p.Profession, currentNode.Name)
			g.Mutex.Unlock()
			continue
		}

		if p.Thirst >= 100 {
			if currentNode.ItemPresent(Water) {
				p.Thirst -= 100
				//				g.Log <- fmt.Sprintf("%s took a drink at %s", p.Profession, currentNode.Name)
				g.Mutex.Unlock()
				continue
			} else if p.Holding(WaterBottle) {
				p.ConsumeItem(WaterBottle)
				p.Thirst -= 100
				//				g.Log <- fmt.Sprintf("%s drank a WATER BOTTLE at %s", p.Profession, currentNode.Name)
				g.Mutex.Unlock()
				continue
			}
		}

		if len(p.Items) < 3 && len(currentNode.Items) > 0 {
			i := rand.Intn(len(currentNode.Items))
			if currentNode.Items[i] != Water {
				p.Items = append(p.Items, currentNode.Items[i])
				//				g.Log <- fmt.Sprintf("%s picked up %s at %s", p.Profession, currentNode.Items[i].StringLong(), currentNode.Name)
				currentNode.Items = append(currentNode.Items[:i], currentNode.Items[i+1:]...)
				currentNode.RenderName(g.atlas)
				g.Mutex.Unlock()
				continue
			}
		}
		g.Mutex.Unlock()

		if rand.Intn(100) == 1 {
			g.Mutex.RLock()
			n := g.From(g.Node(p.Location))
			t := n[rand.Intn(len(n))].(*PositionedNode)
			g.Mutex.RUnlock()
			if len(t.Zombies) > 0 {
				return
			}
			// Humans only pay attention to edge weights because zombies are (presumably) too stupid to fortify
			pause(int(g.Edge(g.Node(p.Location), t).Weight()), time.Second)
			// Did it die before getting to the next vertex?
			if p.checkKilled(g) {
				return
			}
			//			g.Log <- fmt.Sprintf("%s moves from %s to %s", p.Profession, g.Node(p.Location).Name, t.Name)
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

func (z *Zombie) Unlive(g *MapGraph) {
	pause(rand.Intn(2000), time.Millisecond)
	tick := time.NewTicker(500 * time.Millisecond)
	for _ = range tick.C {
		g.Mutex.RLock()
		if z.Hunger >= 150 {
			z.Kill <- "STARVED to DEATH"
		}
		g.Mutex.RUnlock()

		// Check for and take any incoming damage, displaying messages required
		select {
		case m := <-z.Damage:
			g.Mutex.Lock()
			z.Health -= int(m.Value)
			if z.Health <= 0 {
				z.Kill <- fmt.Sprintf("killed by %s with %s", m.Attacker, m.Item.StringLong())
			} else {
				g.Log <- fmt.Sprintf("ZOMBIE at %s takes %d damage from %s wielding %s. Now at %d HP", g.Node(z.Location).Name, m.Value, m.Attacker, m.Item.StringLong(), z.Health)
			}
			g.Mutex.Unlock()
		default:
		}

		if z.checkKilled(g) {
			return
		}

		currentNode := g.Node(z.Location)
		g.Mutex.Lock()
		z.Hunger++

		if len(currentNode.People) > 0 {
			t := currentNode.People[rand.Intn(len(currentNode.People))]
			if int(z.Holding.Damage()) >= t.Health {
				g.InfectPerson(t)
				z.Health = 100
				z.Hunger = 0
				g.Log <- fmt.Sprintf("ZOMBIE INFECTED %s at %s", t.Profession, currentNode.Name)
			} else {
				t.Damage <- DamageMessage{z.Holding.Damage(), "ZOMBIE", z.Holding}
				if z.Holding.Consumable() {
					z.Holding = Nothing
				}
			}
			g.Mutex.Unlock()
			continue
		}

		if z.Holding == Nothing && len(currentNode.Items) > 0 {
			i := rand.Intn(len(currentNode.Items))
			if currentNode.Items[i] != Water {
				z.Holding = currentNode.Items[i]
				//				g.Log <- fmt.Sprintf("ZOMBIE picked up %s at %s", currentNode.Items[i].StringLong(), currentNode.Name)
				currentNode.Items = append(currentNode.Items[:i], currentNode.Items[i+1:]...)
				currentNode.RenderName(g.atlas)
				g.Mutex.Unlock()
				continue
			}

		}
		g.Mutex.Unlock()

		if len(g.Node(z.Location).People) == 0 {
			t := z.nearestPersonTraverseFirstStep(g)
			if t != nil {
				fortification := 0
				if len(t.People) > 0 {
					fortification = t.Weight
					g.Log <- fmt.Sprintf("ZOMBIE is trying to break into %s from %s", t.Name, g.Node(z.Location).Name)
				}
				pause(int(g.Edge(g.Node(z.Location), t).Weight())+fortification*2, time.Second)
				if z.checkKilled(g) {
					return
				}
				if len(t.People) > 0 {
					g.Log <- fmt.Sprintf("ZOMBIE successfully broke into %s from %s", t.Name, g.Node(z.Location).Name)
				}
				z.moveTo(g, t)
			}
		}
	}
}

func (z *Zombie) moveTo(g *MapGraph, t *PositionedNode) {
	g.Mutex.Lock()
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
	var ret *PositionedNode
	if nearestPerson != nil {
		ret = nearestPerson
		for previous[ret.ID()] != z.Location {
			ret = g.Node(previous[ret.ID()])
		}
	} else {
		ret = nil
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
