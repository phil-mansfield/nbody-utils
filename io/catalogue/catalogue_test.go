package catalogue

import (
	"testing"
)

func TestBlockSize(t *testing.T) {
	config := DefaultConfig
	config.MaxLineSize = 5
	config.MaxBlockSize = 15

	rd := TextFile("test_files/block_size_test.txt", config)
	txt, _ := rd.(*textReader)

	targetStarts := []int{0, 14, 29, 44}
	targetEnds := []int{14, 29, 44, 50}

	if !intsEq(txt.blockStarts, targetStarts) {
		t.Errorf("textReader.blockStarts = %d, but should be %d",
			txt.blockStarts, targetStarts)
	}

	if !intsEq(txt.blockEnds, targetEnds) {
		t.Errorf("textReader.blockEnds = %d, but should be %d",
			txt.blockEnds, targetEnds)
	}
}

func intsEq(x, y []int) bool {
	if len(x) != len(y) { return false }
	for i := range x {
		if x[i] != y[i] { return false }
	}
	return true
}
