package chess

import "testing"

const TEST_DEPTH = 5

// func TestStartingMoves(t *testing.T) {
// 	expectedCounts := []int{1, 20, 400, 8902, 197281, 4865609, 119060324}
//
// 	c := NewGame()
// 	// expectedCounts = []int{1, 17}
// 	// c.SetStateFromFEN("rnbqkb1r/p1pppppp/p6n/8/4P3/P7/1PPP1PPP/RNBQK1NR b KQkq - 0 1")
//
// 	for depth, expected := range expectedCounts[:TEST_DEPTH] {
// 		actual, results := c.Perft(depth, depth)
// 		if expected != actual {
// 			t.Errorf("Expected legal move count (%d) does not equal computed count (%d) at depth %d for board\n%s\nPerft results:\n%s", expected, actual, depth, c.EBE.ToFEN(), results)
// 		}
// 	}
// }

func TestPosition2(t *testing.T) {
	expectedCounts := []int{1, 48, 2039, 97862, 4085603, 193690690, 8031647685}

	c := NewGame()
	c.SetStateFromFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - ")

	// expectedCounts = []int{1, 42, 1964, 81066, 3768825}
	// c.SetStateFromFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/1PN2Q1p/P1PBBPPP/R3K2R b KQkq - 0 1")

	for depth, expected := range expectedCounts[:TEST_DEPTH] {
		actual, results := c.Perft(depth, depth)
		if expected != actual {
			t.Errorf("Expected legal move count (%d) does not equal computed count (%d) at depth %d for board\n%s\nPerft results:\n%s", expected, actual, depth, c.EBE.ToFEN(), results)
		}
	}
}
