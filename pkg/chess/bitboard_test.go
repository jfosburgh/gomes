package chess

import (
	"fmt"
	"testing"
)

const (
	starting uint64 = 0b1111111111111111000000000000000000000000000000001111111111111111
)

func BitBoardEqual(t *testing.T, expected, actual BitBoard) {
	for piece := range expected {
		BitPieceEqual(t, piece, expected[piece], actual[piece])
	}
}

func BitPieceEqual(t *testing.T, piece int, expected, actual uint64) {
	if expected != actual {
		t.Errorf(fmt.Sprintf("Expected value for %s doesn't match actual value\nExpected:\n%s\n\nActual:\n%s", piece2String[piece], To2DString(expected), To2DString(actual)))
	}
}

func TestTo2DString(t *testing.T) {
	expected2D := "11111111\n11111111\n00000000\n00000000\n00000000\n00000000\n11111111\n11111111"
	actual2D := To2DString(starting)

	if expected2D != actual2D {
		t.Errorf(fmt.Sprintf("Expected 2d view doesn't match actual 2d view\nExpected:\n%s\n\nActual:\n%s", expected2D, actual2D))
	}
}

func TestStartingBitboard(t *testing.T) {
	ebe := DefaultBoard()
	bitboard := make(BitBoard)
	bitboard.FromEBE(ebe.Board)

	if bitboard.AllPieces() != starting {
		t.Errorf(fmt.Sprintf("Expected 2d view doesn't match actual 2d view\nExpected:\n%s\n\nActual:\n%s", To2DString(starting), To2DString(bitboard.AllPieces())))
	}

	expected := BitBoard{}
	expected[WHITE|PAWN] = uint64(65280)
	expected[WHITE|ROOK] = uint64(129)
	expected[WHITE|KNIGHT] = uint64(66)
	expected[WHITE|BISHOP] = uint64(36)
	expected[WHITE|QUEEN] = uint64(8)
	expected[WHITE|KING] = uint64(16)

	expected[BLACK|PAWN] = uint64(71776119061217280)
	expected[BLACK|ROOK] = uint64(9295429630892703744)
	expected[BLACK|KNIGHT] = uint64(4755801206503243776)
	expected[BLACK|BISHOP] = uint64(2594073385365405696)
	expected[BLACK|QUEEN] = uint64(576460752303423488)
	expected[BLACK|KING] = uint64(1152921504606846976)

	BitBoardEqual(t, expected, bitboard)
}
