package chess

import (
	"cmp"
	"fmt"
	"slices"
	"testing"
)

func ArrayEqual[A cmp.Ordered](expected, actual []A) bool {
	if len(expected) != len(actual) {
		return false
	}

	slices.Sort(expected)
	slices.Sort(actual)

	for i := range expected {
		if expected[i] != actual[i] {
			return false
		}
	}

	return true
}

func EBEEqual(t *testing.T, expected, actual EBE) {
	if expected.Board != actual.Board {
		t.Errorf(fmt.Sprintf("Boards don't match\nExpected:\n%s\n\nActual:\n%s", expected.Board, actual.Board))
	}

	if expected.Active != actual.Active {
		t.Errorf(fmt.Sprintf("Expected active player (%d) != actual active player (%d)", expected.Active, actual.Active))
	}

	if expected.CastlingRights != actual.CastlingRights {
		t.Errorf(fmt.Sprintf("Expected castling rights (%s) != actual castling rights (%s)", castlingRightsToString(expected.CastlingRights), castlingRightsToString(actual.CastlingRights)))
	}

	if !ArrayEqual(expected.EnPassantTargets, actual.EnPassantTargets) {
		expectedString := ""
		for _, pos := range expected.EnPassantTargets {
			expectedString += int2algebraic(pos)
		}

		actualString := ""
		for _, pos := range expected.EnPassantTargets {
			actualString += int2algebraic(pos)
		}

		t.Errorf(fmt.Sprintf("Expected en passant targets (%s) != actual en passant targets (%s)", expectedString, actualString))
	}

	if expected.Halfmoves != actual.Halfmoves {
		t.Errorf(fmt.Sprintf("Expected halfmoves (%d) != actual halfmoves (%d)", expected.Halfmoves, actual.Halfmoves))
	}

	if expected.Moves != actual.Moves {
		t.Errorf(fmt.Sprintf("Expected moves (%d) != actual moves (%d)", expected.Moves, actual.Moves))
	}
}

func TestStartingBoard(t *testing.T) {
	expected := DefaultBoard()

	actual := EBE{}
	actual.FromFEN(StartingFEN)

	EBEEqual(t, expected, actual)
}

func TestStartingFEN(t *testing.T) {
	actual := DefaultBoard()
	actualFEN := actual.ToFEN()

	if StartingFEN != actualFEN {
		t.Errorf(fmt.Sprintf("Expected FEN string != actual FEN string\nExpected:\n%s\n\nActual:\n%s", StartingFEN, actualFEN))
	}
}

func TestFENDecodeEncode(t *testing.T) {
	expected := "rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2"

	board := EBE{}
	board.FromFEN(expected)

	actual := board.ToFEN()
	if expected != actual {
		t.Errorf(fmt.Sprintf("Expected FEN string != actual FEN string\nExpected:\n%s\n\nActual:\n%s", expected, actual))
	}
}
