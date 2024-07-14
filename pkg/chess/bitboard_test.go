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

func CorrectPawnMoves(t *testing.T, bitboard BitBoard, expectedAttacksWhite, expectedMovesWhite, expectedAttacksBlack, expectedMovesBlack uint64) {
	actualAttacksWhite, actualMovesWhite := bitboard.PawnMoves(WHITE)
	actualAttacksBlack, actualMovesBlack := bitboard.PawnMoves(BLACK)

	actualAttacksWhite = actualAttacksWhite & bitboard.SidePieces(BLACK)
	actualAttacksBlack = actualAttacksBlack & bitboard.SidePieces(WHITE)

	if expectedAttacksWhite != actualAttacksWhite {
		t.Errorf("Expected pawn attacks for white do not match actual attacks\nExpected:\n%s\n\nActual:\n%s", To2DString(expectedAttacksWhite), To2DString(actualAttacksWhite))
	}

	if expectedAttacksBlack != actualAttacksBlack {
		t.Errorf("Expected pawn attacks for black do not match actual attacks\nExpected:\n%s\n\nActual:\n%s", To2DString(expectedAttacksBlack), To2DString(actualAttacksBlack))
	}

	if expectedMovesWhite != actualMovesWhite {
		t.Errorf("Expected pawn moves for white do not match actual moves\nExpected:\n%s\n\nActual:\n%s", To2DString(expectedMovesWhite), To2DString(actualMovesWhite))
	}

	if expectedMovesBlack != actualMovesBlack {
		t.Errorf("Expected pawn moves for black do not match actual moves\nExpected:\n%s\n\nActual:\n%s", To2DString(expectedMovesBlack), To2DString(actualMovesBlack))
	}
}

func CorrectMoves(t *testing.T, b BitBoard, side int, expectedMoves uint64, generator func(int) uint64) {
	moves := generator(side)

	if expectedMoves != moves {
		t.Errorf("Expected moves != actual moves for side %d\nExpected:\n%s\n\nActual:\n%s\n", side, To2DString(expectedMoves), To2DString(moves))
	}
}

func TestPawnMoveGeneration(t *testing.T) {
	ebe := DefaultBoard()
	bitboard := make(BitBoard)
	bitboard.FromEBE(ebe.Board)

	expectedAttacks := uint64(0)
	expectedMovesWhite := uint64(0b1111111111111111) << 16
	expectedMovesBlack := uint64(0b1111111111111111) << 32

	CorrectPawnMoves(t, bitboard, expectedAttacks, expectedMovesWhite, expectedAttacks, expectedMovesBlack)

	bitboard.Add(WHITE|PAWN, 24)
	bitboard.Remove(WHITE|PAWN, 8)

	bitboard.Add(WHITE|PAWN, 27)
	bitboard.Remove(WHITE|PAWN, 11)

	bitboard.Add(WHITE|PAWN, 31)
	bitboard.Remove(WHITE|PAWN, 15)

	bitboard.Add(BLACK|PAWN, 33)
	bitboard.Remove(BLACK|PAWN, 49)

	bitboard.Add(BLACK|PAWN, 36)
	bitboard.Remove(BLACK|PAWN, 52)

	bitboard.Add(BLACK|PAWN, 39)
	bitboard.Remove(BLACK|PAWN, 55)

	bitboard.UpdateSide(WHITE)
	bitboard.UpdateSide(BLACK)

	CorrectPawnMoves(t, bitboard, uint64(77309411328), uint64(40642150400), uint64(150994944), uint64(120315220852736))
}

func TestKnightMoveGeneration(t *testing.T) {
	bitboard := make(BitBoard)
	bitboard[WHITE|KNIGHT] = uint64(0b01000010)
	bitboard.UpdateSide(WHITE)
	bitboard.UpdateSide(BLACK)

	expectedThreatens := uint64(10819584)
	expectedCaptures := uint64(0)

	CorrectMoves(t, bitboard, WHITE, expectedCaptures|expectedThreatens, bitboard.KnightMoves)

	ebe := DefaultBoard()
	bitboard.FromEBE(ebe.Board)

	expectedThreatensWhite := uint64(10813440)
	expectedThreatensBlack := uint64(181419418583040)

	CorrectMoves(t, bitboard, WHITE, expectedCaptures|expectedThreatensWhite, bitboard.KnightMoves)
	CorrectMoves(t, bitboard, BLACK, expectedCaptures|expectedThreatensBlack, bitboard.KnightMoves)
}

func TestKingMoveGeneration(t *testing.T) {
	bitboard := make(BitBoard)
	bitboard[WHITE|KING] = uint64(0b00001000)
	bitboard.UpdateSide(WHITE)
	bitboard.UpdateSide(BLACK)

	expectedThreatens := uint64(7188)
	expectedCaptures := uint64(0)

	CorrectMoves(t, bitboard, WHITE, expectedCaptures|expectedThreatens, bitboard.KingMoves)

	ebe := DefaultBoard()
	bitboard.FromEBE(ebe.Board)

	expectedThreatensWhite := uint64(0)
	expectedThreatensBlack := uint64(0)

	CorrectMoves(t, bitboard, WHITE, expectedCaptures|expectedThreatensWhite, bitboard.KingMoves)
	CorrectMoves(t, bitboard, BLACK, expectedCaptures|expectedThreatensBlack, bitboard.KingMoves)
}

func TestRookMoveGeneration(t *testing.T) {
	bitboard := make(BitBoard)
	bitboard[WHITE|ROOK] = uint64(0b10000000)
	bitboard.UpdateSide(WHITE)
	bitboard.UpdateSide(BLACK)

	expected := verticalCross(7) & (^bitboard[WHITE|ROOK])
	CorrectMoves(t, bitboard, WHITE, expected, bitboard.RookMoves)

	ebe := DefaultBoard()
	bitboard.FromEBE(ebe.Board)

	CorrectMoves(t, bitboard, WHITE, uint64(0), bitboard.RookMoves)
	CorrectMoves(t, bitboard, BLACK, uint64(0), bitboard.RookMoves)
}

func TestBishopMoveGeneration(t *testing.T) {
	bitboard := make(BitBoard)
	bitboard[WHITE|BISHOP] = uint64(0b00100000)
	bitboard.UpdateSide(WHITE)
	bitboard.UpdateSide(BLACK)

	expected := diagonalCross(5) & (^bitboard[WHITE|BISHOP])
	CorrectMoves(t, bitboard, WHITE, expected, bitboard.BishopMoves)

	ebe := DefaultBoard()
	bitboard.FromEBE(ebe.Board)

	CorrectMoves(t, bitboard, WHITE, uint64(0), bitboard.BishopMoves)
	CorrectMoves(t, bitboard, BLACK, uint64(0), bitboard.BishopMoves)
}

func TestQueenMoveGeneration(t *testing.T) {
	bitboard := make(BitBoard)
	bitboard[WHITE|QUEEN] = uint64(0b00001000)
	bitboard.UpdateSide(WHITE)
	bitboard.UpdateSide(BLACK)

	expected := (diagonalCross(3) | verticalCross(3)) & (^bitboard[WHITE|QUEEN])
	CorrectMoves(t, bitboard, WHITE, expected, bitboard.QueenMoves)

	ebe := DefaultBoard()
	bitboard.FromEBE(ebe.Board)

	CorrectMoves(t, bitboard, WHITE, uint64(0), bitboard.QueenMoves)
	CorrectMoves(t, bitboard, BLACK, uint64(0), bitboard.QueenMoves)
}

func TestVerticalCross(t *testing.T) {
	expected := uint64(144680345692733954)
	actual := verticalCross(17)

	if expected != actual {
		t.Errorf("Expected does not match actual\nExpected:\n%s\n\nActual:\n%s", To2DString(expected), To2DString(actual))
	}
}

func TestDiagonalCross(t *testing.T) {
	expected := uint64(4620710844311930120)
	actual := diagonalCross(17)

	if expected != actual {
		t.Errorf("Expected does not match actual\nExpected:\n%s\n\nActual:\n%s", To2DString(expected), To2DString(actual))
	}
}

func TestToPieceLocations(t *testing.T) {
	ebe := DefaultBoard()
	bitboard := make(BitBoard)
	bitboard.FromEBE(ebe.Board)

	expected := []int{1, 6}
	actual := toPieceLocations(bitboard[WHITE|KNIGHT])

	if !ArrayEqual(expected, actual) {
		t.Errorf("Expected locations (%+v) != actual locations (%+v) for pieces:\n%s", expected, actual, To2DString(bitboard[WHITE|KNIGHT]))
	}
}
