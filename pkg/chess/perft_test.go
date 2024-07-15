package chess

import (
	"os"
	"runtime/pprof"
	"testing"
)

const TEST_DEPTH = 7
const DEBUG = false

func TestPerftStarting(t *testing.T) {
	expectedCounts := []int{1, 20, 400, 8902, 197281, 4865609, 119060324, 3195901860, 84998978956, 2439530234167, 69352859712417, 2097651003696806, 62854969236701747, 1981066775000396239, 61, 885021521585529237} // 2015099950053364471960

	c := NewGame()
	// expectedCounts = []int{1, 17}
	// c.SetStateFromFEN("rnbqkb1r/p1pppppp/p6n/8/4P3/P7/1PPP1PPP/RNBQK1NR b KQkq - 0 1")

	for depth, expected := range expectedCounts[:min(TEST_DEPTH, len(expectedCounts))] {
		actual, results := c.Perft(depth, depth, DEBUG)
		if expected != actual {
			t.Errorf("Expected legal move count (%d) does not equal computed count (%d) at depth %d for board\n%s\nPerft results:\n%s", expected, actual, depth, c.EBE.ToFEN(), results)
		}
	}
}

func TestPerftPosition2(t *testing.T) {
	expectedCounts := []int{1, 48, 2039, 97862, 4085603, 193690690, 8031647685}

	c := NewGame()
	c.SetStateFromFEN("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - ")

	// expectedCounts = []int{1, 45, 2066}
	// c.SetStateFromFEN("r3k2r/p1ppqpb1/1n2pnp1/1b1PN3/Pp2P3/2N2QPp/1PPBBP1P/R3K2R b KQkq a3 0 1")

	for depth, expected := range expectedCounts[:min(TEST_DEPTH, len(expectedCounts))] {
		actual, results := c.Perft(depth, depth, DEBUG)
		if expected != actual {
			t.Errorf("Expected legal move count (%d) does not equal computed count (%d) at depth %d for board\n%s\nPerft results:\n%s", expected, actual, depth, c.EBE.ToFEN(), results)
		}
	}
}

func TestPerftPosition3(t *testing.T) {
	expectedCounts := []int{1, 14, 191, 2812, 43238, 674624, 11030083, 178633661, 3009794393}

	c := NewGame()
	c.SetStateFromFEN("8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - -")

	// expectedCounts = []int{1, 45, 2066}
	// c.SetStateFromFEN("r3k2r/p1ppqpb1/1n2pnp1/1b1PN3/Pp2P3/2N2QPp/1PPBBP1P/R3K2R b KQkq a3 0 1")

	for depth, expected := range expectedCounts[:min(TEST_DEPTH, len(expectedCounts))] {
		actual, results := c.Perft(depth, depth, DEBUG)
		if expected != actual {
			t.Errorf("Expected legal move count (%d) does not equal computed count (%d) at depth %d for board\n%s\nPerft results:\n%s", expected, actual, depth, c.EBE.ToFEN(), results)
		}
	}
}

func TestPerftPosition4(t *testing.T) {
	expectedCounts := []int{1, 6, 264, 9467, 422333, 15833292, 706045033}

	c := NewGame()
	c.SetStateFromFEN("r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1")

	// expectedCounts = []int{1, 40}
	// c.SetStateFromFEN("r3k1Nr/Pppp1ppp/1b3nb1/nPP5/BB2P3/q4N2/P2P2PP/b2Q1RK1 b kq - 0 1")

	for depth, expected := range expectedCounts[:min(TEST_DEPTH, len(expectedCounts))] {
		actual, results := c.Perft(depth, depth, DEBUG)
		if expected != actual {
			t.Errorf("Expected legal move count (%d) does not equal computed count (%d) at depth %d for board\n%s\nPerft results:\n%s", expected, actual, depth, c.EBE.ToFEN(), results)
		}
	}
}

func TestPerftPosition4Mirrored(t *testing.T) {
	expectedCounts := []int{1, 6, 264, 9467, 422333, 15833292, 706045033}

	c := NewGame()
	c.SetStateFromFEN("r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq - 0 1")

	// expectedCounts = []int{1, 45, 2066}
	// c.SetStateFromFEN("r3k2r/p1ppqpb1/1n2pnp1/1b1PN3/Pp2P3/2N2QPp/1PPBBP1P/R3K2R b KQkq a3 0 1")

	for depth, expected := range expectedCounts[:min(TEST_DEPTH, len(expectedCounts))] {
		actual, results := c.Perft(depth, depth, DEBUG)
		if expected != actual {
			t.Errorf("Expected legal move count (%d) does not equal computed count (%d) at depth %d for board\n%s\nPerft results:\n%s", expected, actual, depth, c.EBE.ToFEN(), results)
		}
	}
}

func TestPerftPosition5(t *testing.T) {
	expectedCounts := []int{1, 44, 1486, 62379, 2103487, 89941194}

	c := NewGame()
	c.SetStateFromFEN("rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ - 1 8")

	// expectedCounts = []int{1, 45, 2066}
	// c.SetStateFromFEN("r3k2r/p1ppqpb1/1n2pnp1/1b1PN3/Pp2P3/2N2QPp/1PPBBP1P/R3K2R b KQkq a3 0 1")

	for depth, expected := range expectedCounts[:min(TEST_DEPTH, len(expectedCounts))] {
		actual, results := c.Perft(depth, depth, DEBUG)
		if expected != actual {
			t.Errorf("Expected legal move count (%d) does not equal computed count (%d) at depth %d for board\n%s\nPerft results:\n%s", expected, actual, depth, c.EBE.ToFEN(), results)
		}
	}
}

func TestPerftPosition6(t *testing.T) {
	expectedCounts := []int{1, 46, 2079, 89890, 3894594, 164075551, 6923051137, 287188994746, 11923589843526, 490154852788714}

	c := NewGame()
	c.SetStateFromFEN("r4rk1/1pp1qppp/p1np1n2/2b1p1B1/2B1P1b1/P1NP1N2/1PP1QPPP/R4RK1 w - - 0 10")

	// expectedCounts = []int{1, 45, 2066}
	// c.SetStateFromFEN("r3k2r/p1ppqpb1/1n2pnp1/1b1PN3/Pp2P3/2N2QPp/1PPBBP1P/R3K2R b KQkq a3 0 1")

	for depth, expected := range expectedCounts[:min(TEST_DEPTH, len(expectedCounts))] {
		actual, results := c.Perft(depth, depth, DEBUG)
		if expected != actual {
			t.Errorf("Expected legal move count (%d) does not equal computed count (%d) at depth %d for board\n%s\nPerft results:\n%s", expected, actual, depth, c.EBE.ToFEN(), results)
		}
	}
}

func BenchmarkPerft(b *testing.B) {
	f, err := os.Create("cpu.pprof")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	expectedCounts := []int{1, 20, 400, 8902, 197281, 4865609, 119060324}

	c := NewGame()

	for depth := range expectedCounts[:TEST_DEPTH] {
		c.Perft(depth, depth, DEBUG)
	}
}
