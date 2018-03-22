package vis

// Data structures representing people, zombies, and items

type Item uint

const (
	Chainsaw Item = iota
	Handgun
	Rifle
	EnergyBar
	Water
	RustyPipe
	Hatchet
	AerosolFlamethrower
	Bandage
)

type Profession uint

const (
	Police Profession = iota
	Firefighter
	Soldier
	Doctor
	Engineer
)

type Person struct {
	Health     int
	Hunger     int
	Items      []Item
	Profession Profession
}

type Zombie struct {
	Health  int
	Hunger  int
	Holding Item
}
