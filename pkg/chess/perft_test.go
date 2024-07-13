package chess

import "testing"

func TestStartingMoves(t *testing.T) {
	expectedCounts := []int{1, 20, 400, 8902, 197281, 4865609}

	c := NewGame()
	// expectedCounts = []int{1, 23}
	// c.SetStateFromFEN("rnbqkb1r/p1pppppp/5n2/Pp6/8/8/1PPPPPPP/RNBQKBNR w KQkq b6 0 1")

	for depth, expected := range expectedCounts {
		actual, results := c.Perft(depth, depth)
		if expected != actual {
			t.Errorf("Expected legal move count (%d) does not equal computed count (%d) at depth %d for board\n%s\nPerft results:\n%s", expected, actual, depth, c.EBE.ToFEN(), results)
		}
	}
}
