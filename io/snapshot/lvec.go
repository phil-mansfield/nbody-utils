package snapshot

const (
	LVecMagicNumber = 0xbadf00d
	LVecVersion = 1

	lvecBoxMethod= iota

	lvecX = iota
	lvecV
)

type lvecHeader struct {
	magic   uint64 // Magic number confirming that this file is a .lvec file
	version uint64 // Version number of the code that generated this file.
	varType uint64 // Flag showing the type of variable stored in the file.
	method  uint64 // Flag showing the method used to store data within
	               // the file.

	idx   uint64 // The index of the current file in the Lagrangian grid.
	cells uint64 // The number of file-sized cells on one side of the
	             // Lagrangian grid.

	subCells           uint64 // The number of sub-cells used inside this file.
	subCellVectorsMin  uint64 // 
	subCellVectorsBits uint64 // Number of bits used to represent the vectors to
	                          // each sub-cell.

	bitsMin  uint64 // Minimum bits value.
	bitsBits uint64 // Number of bits used to represent the number of bits in
	                // each cell.
	

	pix        uint64 // Number of "pixels" on one side of the quantiation grid.
	limits [2]float64 // The minimum and maximum value of each side of the box.
	delta     float64 // The user-specified delta parameter. Each component of
                      // each vector will be stored to at least this accuracy.

	offsets [4]uint64 // The offsets to the start of the sub-cell vector block,
	                  // the bits block, and the data block, respetively. This
	                  // isn't neccessary, but it sure is convenient.

	hd Header // The header for the simulation.
}
