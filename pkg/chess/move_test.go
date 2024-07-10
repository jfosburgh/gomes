package chess

import "testing"

func GameEquals(t *testing.T, expected, actual *ChessGame) {
	EBEEqual(t, expected.EBE, actual.EBE)
	// BitBoardEqual(t, expected.Bitboard, actual.Bitboard)
}

func TestMakeUnmakeMove(t *testing.T) {
	fenStates := []string{
		StartingFEN,
		"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2",
		"rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2",
	}

	moves := []Move{
		{
			Piece: WHITE | PAWN,
			Start: algebraic2Int("e2"),
			End:   algebraic2Int("e4"),
		},
		{
			Piece: BLACK | PAWN,
			Start: algebraic2Int("c7"),
			End:   algebraic2Int("c5"),
		},
		{
			Piece: WHITE | KNIGHT,
			Start: algebraic2Int("g1"),
			End:   algebraic2Int("f3"),
		},
	}

	expected := []*ChessGame{}
	for _, fen := range fenStates {
		newGame := NewGame()
		newGame.SetStateFromFEN(fen)
		expected = append(expected, newGame)
	}

	actual := NewGame()
	for i := range moves {
		moves[i].Halfmoves = actual.EBE.Halfmoves
		moves[i].CastlingRights = actual.EBE.CastlingRights
		moves[i].EnPassantTarget = actual.EBE.EnPassantTarget

		actual.MakeMove(moves[i])
		GameEquals(t, expected[i+1], actual)
	}

	for i := len(moves) - 1; i >= 0; i-- {
		actual.UnmakeMove(moves[i])
		GameEquals(t, expected[i], actual)
	}
}
