package netlist

// Kind - вид списка подсетей. Значения хранятся в базе как есть,
// поэтому меняются только вместе с миграцией.
type Kind string

const (
	// Whitelist - подсети, попытки из которых всегда разрешены.
	Whitelist Kind = "white"

	// Blacklist - подсети, попытки из которых всегда отклонены.
	Blacklist Kind = "black"
)
