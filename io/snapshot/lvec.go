package snapshot

const (
	LVecMagicNumber = 0xbadf00d
	LVecVersion = 1

	lvecBoxMethod= iota

	lvecX = iota
	lvecV
)

type lvecHeader struct {
	magic, version, method, varType uint64
	cells, subCells, idx int64
	hd Header
)
