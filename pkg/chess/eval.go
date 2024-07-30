package chess

import (
	"fmt"
	"slices"
)

func (c *ChessGame) Search() ([]Move, []int) {
	fmt.Printf("starting search at depth %d\n", c.SearchDepth)
	options := c.GetLegalMoves()
	vals := make([]int, len(options))

	evaluated := 0
	for i, move := range options {
		c.MakeMove(move)
		var e int
		vals[i], e = c.Minimax(c.SearchDepth - 1)
		evaluated += e
		c.UnmakeMove(move)
	}

	for i := range len(options) - 1 {
		for j := 0; j < len(options)-i-1; j++ {
			if vals[j] > vals[j+1] {
				temp := vals[j]
				vals[j] = vals[j+1]
				vals[j+1] = temp

				temp2 := options[j]
				options[j] = options[j+1]
				options[j+1] = temp2
			}
		}
	}

	fmt.Println("search results")
	for i := range options {
		fmt.Printf(" > %s == %d\n", options[i], vals[i])
	}
	fmt.Printf("searched %d nodes\n", evaluated)

	return options, vals
}

func (c *ChessGame) Minimax(depth int) (int, int) {
	if depth <= 0 || c.EBE.Halfmoves >= 100 {
		return c.Evaluate(), 1
	}

	moves := c.GetLegalMoves()
	if len(moves) == 0 {
		return c.Evaluate(), 1
	}
	vals := make([]int, len(moves))

	evaluated := 0
	for i, move := range moves {
		c.MakeMove(move)
		var e int
		vals[i], e = c.Minimax(depth - 1)
		evaluated += e
		c.UnmakeMove(move)
	}

	if c.EBE.Active<<3 == BLACK {
		return slices.Min(vals), evaluated
	}
	return slices.Max(vals), evaluated
}

func (c *ChessGame) Evaluate() int {
	score := c.Material(c.EBE.Active<<3) - c.Material(enemy(c.EBE.Active<<3))
	if c.EBE.Active<<3 == BLACK {
		score *= -1
	}

	return score
}

func (c *ChessGame) Material(side int) int {
	score := 2000 * len(toPieceLocations(c.Bitboard[side|KING]))
	score += 90 * len(toPieceLocations(c.Bitboard[side|QUEEN]))
	score += 50 * len(toPieceLocations(c.Bitboard[side|ROOK]))
	score += 30 * len(toPieceLocations(c.Bitboard[side|BISHOP]))
	score += 30 * len(toPieceLocations(c.Bitboard[side|KNIGHT]))

	pawnBoard := c.Bitboard[side|PAWN]
	pawns := toPieceLocations(pawnBoard)
	score += 10 * len(pawns)
	score -= 5 * len(toPieceLocations(pawnBoard&(pawnBoard<<8)))

	blocked := 0
	isolated := 0
	for _, pawn := range pawns {
		forward, diagLeft, diagRight := 8, 7, 9
		if side == BLACK {
			forward, diagLeft, diagRight = -8, -9, -7
		}

		b := true
		if pawn+forward >= 0 && pawn+forward < 64 {
			b = b && c.EBE.Board[pawn+forward] != 0
		}
		if pawn%8 > 0 && pawn+diagLeft >= 0 && pawn+diagRight < 64 {
			b = b && c.EBE.Board[pawn+diagLeft]&0b1000 != enemy(side)
		}
		if pawn%8 < 7 && pawn+diagRight >= 0 && pawn+diagRight < 64 {
			b = b && c.EBE.Board[pawn+diagRight]&0b1000 != enemy(side)
		}
		if b {
			blocked += 1
		}

		file := pawn%8 + 1
		mask := uint64(0)
		if file > 1 {
			mask = mask | fileMask(file-1)
		}
		if file < 8 {
			mask = mask | fileMask(file+1)
		}

		if pawnBoard&mask == 0 {
			isolated += 1
		}
	}

	score -= 5 * (blocked + isolated)

	flip := false
	if side != c.EBE.Active<<3 {
		flip = true
		c.EBE.Active = (^c.EBE.Active) & 0b1
	}

	score += len(c.GetLegalMoves())
	if flip {
		c.EBE.Active = (^c.EBE.Active) & 0b1
	}

	return score
}
