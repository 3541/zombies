package entity

import (
	"math/rand"
)

// Data structures representing people, zombies, and items

type Item uint

const N_ITEMS int = 15

const (
	Chainsaw Item = iota
	Pistol
	Rifle
	EnergyBar
	Water
	WaterBottle
	RustyPipe
	Hatchet
	AerosolFlamethrower
	Bandage
	Wrench
	Hacksaw
	RPG
	ATGM
	HolyWater
)

func (i Item) StringLong() string {
	switch i {
	case Chainsaw:
		return "CHAINSAW"
	case Pistol:
		return "PISTOL"
	case Rifle:
		return "RIFLE"
	case EnergyBar:
		return "ENERGY BAR"
	case Water:
		return "WATER SOURCE"
	case WaterBottle:
		return "WATER BOTTLE"
	case RustyPipe:
		return "RUSTY PIPE"
	case Hatchet:
		return "HATCHET"
	case AerosolFlamethrower:
		return "IMPROVISED AEROSOL FLAMETHROWER"
	case Bandage:
		return "BANDAGE"
	case Wrench:
		return "WRENCH"
	case Hacksaw:
		return "HACKSAW"
	case RPG:
		return "ROCKET-PROPELLED GRENADE LAUNCHER"
	case ATGM:
		return "ANTI-TANK GUIDED MISSILE"
	case HolyWater:
		return "HOLY WATER"
	default:
		return "INVALID ITEM"
	}
	return "INVALID ITEM"
}

func (i Item) String() string {
	switch i {
	case Chainsaw:
		return "CS"
	case Pistol:
		return "PSTL"
	case Rifle:
		return "RFL"
	case EnergyBar:
		return "EB"
	case Water:
		return "WS"
	case WaterBottle:
		return "WB"
	case RustyPipe:
		return "RP"
	case Hatchet:
		return "HTCHT"
	case AerosolFlamethrower:
		return "IAF"
	case Bandage:
		return "BDG"
	case Wrench:
		return "WRNC"
	case Hacksaw:
		return "HS"
	case RPG:
		return "RPG"
	case ATGM:
		return "ATGM"
	case HolyWater:
		return "HW"
	default:
		return "INVALID ITEM"
	}
	return "INVALID ITEM"
}

type Profession uint

const (
	Police Profession = iota
	Firefighter
	Soldier
	Doctor
	Engineer
	Priest
	Other
)

func (p Profession) String() string {
	switch p {
	case Police:
		return "POLICE OFFICER"
	case Firefighter:
		return "FIREFIGHTER"
	case Soldier:
		return "SOLDIER"
	case Doctor:
		return "DOCTOR"
	case Engineer:
		return "ENGINEER"
	case Other:
		return "OTHER"
	case Priest:
		return "PRIEST"
	default:
		return "INVALID PROFESSION"
	}
	return "INVALID PROFESSION"
}

type Person struct {
	Id         uint
	Health     uint
	Hunger     uint
	Thirst     uint
	Items      []Item
	Profession Profession
	Location   int
}

func (p *Person) AddItem(items ...Item) {
	p.Items = append(p.Items, items...)
}

func NewPerson(id uint, job Profession, pos int) *Person {
	ret := &Person{id, 100, 0, 0, make([]Item, 0, 2), job, pos}
	switch job {
	case Police:
		ret.AddItem(Pistol)
	case Firefighter:
		if rand.Intn(5) == 2 {
			ret.AddItem(Chainsaw)
		} else {
			ret.AddItem(Hatchet)
		}
	case Soldier:
		ret.AddItem(Rifle)
		if rand.Intn(5) == 0 {
			if rand.Intn(3) == 0 {
				ret.AddItem(ATGM)
			} else {
				ret.AddItem(RPG)
			}
		}
	case Doctor:
		ret.AddItem(Bandage, Bandage, Bandage, Bandage, Hacksaw)
	case Engineer:
		ret.AddItem(Hatchet, Wrench)
	case Priest:
		ret.AddItem(HolyWater)
	}

	if rand.Intn(3) == 1 {
		ret.AddItem(EnergyBar)
	}

	if rand.Intn(2) == 0 {
		ret.AddItem(WaterBottle)
	}

	if rand.Intn(10) == 1 {
		ret.AddItem(RustyPipe)
	}
	return ret
}

type Zombie struct {
	Id       uint
	Health   int
	Hunger   int
	Holding  Item
	Location int
}

func NewZombieFromPerson(victim Person) *Zombie {
	return &Zombie{victim.Id, 100, 0, victim.Items[rand.Intn(len(victim.Items))], victim.Location}
}

func NewZombie(id int, location int) *Zombie {
	return &Zombie{id, 100, 0, nil, location}
}
