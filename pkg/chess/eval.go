package chess

import (
	"fmt"
	"math"
)

func (c *ChessGame) Search() ([]Move, []float64) {
	fmt.Printf("starting search at depth %d\n", c.SearchDepth)
	options := c.GetLegalMoves()
	vals := make([]float64, len(options))

	evaluated := 0
	skipped := 0
	for i, move := range options {
		c.MakeMove(move)
		var e int
		v, e, s := c.Minimax(c.SearchDepth-1, math.Inf(-1), math.Inf(1))
		vals[i] = v
		evaluated += e
		skipped += s
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
		fmt.Printf(" > %s == %f\n", options[i], vals[i])
	}
	fmt.Printf("searched %d nodes, skipped %d\n", evaluated, skipped)

	return options, vals
}

func (c *ChessGame) Minimax(depth int, alpha, beta float64) (float64, int, int) {
	if depth <= 0 || c.EBE.Halfmoves >= 100 {
		return c.Evaluate(), 1, 0
	}

	moves := c.GetLegalMoves()
	if len(moves) == 0 {
		return c.Evaluate(), 1, 0
	}

	evaluated := 0
	skipped := 0
	checked := 0
	if c.EBE.Active<<3 == WHITE {
		value := math.Inf(-1)
		for _, move := range moves {
			c.MakeMove(move)
			v, e, s := c.Minimax(depth-1, alpha, beta)
			value = max(value, v)
			evaluated += e
			skipped += s
			c.UnmakeMove(move)

			checked += 1
			if value > beta {
				break
			}
			alpha = max(alpha, value)
		}
		return value, evaluated, skipped + len(moves) - checked
	} else {
		value := math.Inf(1)
		for _, move := range moves {
			c.MakeMove(move)
			v, e, s := c.Minimax(depth-1, alpha, beta)
			value = min(value, v)
			evaluated += e
			skipped += s
			c.UnmakeMove(move)

			checked += 1
			if value < alpha {
				break
			}
			beta = min(beta, value)
		}
		return value, evaluated, skipped + len(moves) - checked
	}
}

func (c *ChessGame) Evaluate() float64 {
	score := float64(c.Material(c.EBE.Active<<3) - c.Material(enemy(c.EBE.Active<<3)))
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
