package chess

import (
	"fmt"
	"time"
)

type ChessGame struct {
	EBE      EBE
	Bitboard BitBoard
	Moves    []Move
	Captured []int
}

func NewGame() *ChessGame {
	c := ChessGame{
		EBE:      DefaultBoard(),
		Bitboard: make(BitBoard),
	}

	c.Bitboard.FromEBE(c.EBE.Board)

	return &c
}

func (c *ChessGame) SetStateFromFEN(fen string) {
	c.EBE.FromFEN(fen)
	c.Bitboard.FromEBE(c.EBE.Board)
	c.Moves = []Move{}
	c.Captured = []int{}
}

func copyBitboard(source, dest BitBoard) {
	for piece := range source {
		dest[piece] = source[piece]
	}
}

func copyBoard(source EBEBoard) EBEBoard {
	board := EBEBoard{}
	for i := range 64 {
		board[i] = source[i]
	}

	return board
}

func (c *ChessGame) Perft(depth, startDepth int) (int, string) {
	start := time.Now()
	if depth == 0 {
		return 1, ""
	}

	resultString := ""

	count := 0
	moves := c.GeneratePseudoLegal()
	if depth == startDepth {
		fmt.Printf("starting search with board state:\nActive - %d\nCastling Rights - %04b\n%s\n", c.EBE.Active, c.EBE.CastlingRights, c.EBE.Board)
		for piece := range piece2String {
			fmt.Printf("%s: %+v, ", piece2String[piece], toPieceLocations(c.Bitboard[piece]))
		}
		fmt.Println(moves)
	}

	for _, move := range moves {
		moveCount := 0
		active := c.EBE.Active << 3

		startingBitboard := make(BitBoard)
		copyBitboard(c.Bitboard, startingBitboard)

		startingBoard := EBE{}
		startingBoard.Board = copyBoard(c.EBE.Board)

		c.MakeMove(move)
		if !c.Bitboard.InCheck(active) {
			c, _ := c.Perft(depth-1, startDepth)
			moveCount += c
		}
		c.UnmakeMove(move)

		if startingBoard.Board != c.EBE.Board {
			panic(fmt.Sprintf("board before move %s%s doesn't match board after\nBefore:\n%s\nAfter:\n%s", int2algebraic(move.Start), int2algebraic(move.End), startingBoard.Board, c.EBE.Board))
		}

		for piece := range startingBitboard {
			if startingBitboard[piece] != c.Bitboard[piece] {
				panic(fmt.Sprintf("Piece board for %s is different after %s%s\nStarting\n%s\nEnding\n%s", piece2String[piece], int2algebraic(move.Start), int2algebraic(move.End), To2DString(startingBitboard[piece]), To2DString(c.Bitboard[piece])))
			}
		}

		count += moveCount
		if depth == startDepth {
			resultString += fmt.Sprintf("%s: %d\n", move, moveCount)
		}
	}

	if depth == startDepth {
		fmt.Printf("\nending search with board state:\nActive - %d\nCastling Rights - %04b\n%s\n", c.EBE.Active, c.EBE.CastlingRights, c.EBE.Board)
		for piece := range piece2String {
			fmt.Printf("%s: %+v, ", piece2String[piece], toPieceLocations(c.Bitboard[piece]))
		}
	}
	if depth == startDepth {
		fmt.Printf("\nperft evaluated to depth of %d in %dms\n", startDepth, time.Since(start).Milliseconds())
	}

	return count, resultString
}
