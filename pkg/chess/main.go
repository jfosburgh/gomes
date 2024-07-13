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

func (c *ChessGame) Perft(depth, startDepth int) (int, string) {
	start := time.Now()
	if depth == 0 {
		return 1, ""
	}

	resultString := ""

	count := 0
	moves := c.GeneratePseudoLegal()
	for _, move := range moves {
		moveCount := 0
		active := c.EBE.Active << 3
		c.MakeMove(move)
		if !c.Bitboard.InCheck(active) {
			c, _ := c.Perft(depth-1, startDepth)
			moveCount += c
		}
		c.UnmakeMove(move)
		count += moveCount
		if depth == startDepth {
			resultString += fmt.Sprintf("%s%s: %d\n", int2algebraic(move.Start), int2algebraic(move.End), moveCount)
		}
	}

	if depth == startDepth {
		fmt.Printf("perft evaluated to depth of %d in %dms\n", startDepth, time.Since(start).Milliseconds())
	}

	return count, resultString
}
