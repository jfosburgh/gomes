package chess

import (
	"fmt"
	"math"
	"time"
)

func (c *ChessGame) Search() ([]Move, []float64) {
	fmt.Printf("starting with max depth %d\n", c.MaxSearchDepth)
	options := c.GetLegalMoves()

	vals := make([]float64, len(options))
	c.Transpositions = make(map[EBEBoard]TranspositionNode)

	c.SearchStart = time.Now()
	c.SearchTimer = time.NewTimer(c.SearchTime)
	defer c.SearchTimer.Stop()

	for depth := range c.MaxSearchDepth {
		finished := true
		evaluated := 0
		skipped := 0
		if depth != 0 {
			options = c.PreOrder(options)
		}
		searchVals := make([]float64, len(options))
		for i, move := range options {
			select {
			case <-c.SearchTimer.C:
				finished = false
			default:
				c.MakeMove(move)
				v, e, s := c.Minimax(0, depth, math.Inf(-1), math.Inf(1))
				c.UnmakeMove(move)
				if e == -1 {
					finished = false
					fmt.Println("setting finished false")
				}

				searchVals[i] = v
				evaluated += e
				skipped += s
			}
			if !finished {
				break
			}
		}
		if !finished {
			fmt.Printf("max search time (%dms) reached before search to depth %d finished, exited without updating vals\n", c.SearchTime.Milliseconds(), depth)
			break
		}

		fmt.Printf("searched %d nodes, skipped %d to depth %d, %dms since search start\n", evaluated, skipped, depth, time.Since(c.SearchStart).Milliseconds())
		vals = searchVals
	}

	options, vals = sortMoves(options, vals, true)

	fmt.Println("search results")
	for i := range options {
		fmt.Printf(" > %s == %f\n", options[i], vals[i])
	}
	fmt.Printf("total search time: %dms\n", time.Since(c.SearchStart).Milliseconds())

	return options, vals
}

func sortMoves(options []Move, vals []float64, ascending bool) ([]Move, []float64) {
	for i := range len(options) - 1 {
		for j := 0; j < len(options)-i-1; j++ {
			if (vals[j] > vals[j+1] && ascending) || (vals[j] < vals[j+1] && !ascending) {
				temp := vals[j]
				vals[j] = vals[j+1]
				vals[j+1] = temp

				temp2 := options[j]
				options[j] = options[j+1]
				options[j+1] = temp2
			}
		}
	}

	return options, vals
}

func (c *ChessGame) PreOrder(moves []Move) []Move {
	vals := make([]float64, len(moves))
	for i := range len(moves) {
		c.MakeMove(moves[i])
		v, ok := c.Transpositions[c.EBE.Board]
		if !ok {
			// TODO: replace with guestimate based on move
			vals[i] = 0
		} else {
			vals[i] = v.Value
		}
		c.UnmakeMove(moves[i])
	}

	moves, _ = sortMoves(moves, vals, c.EBE.Active<<3 == BLACK)
	return moves
}

func (c *ChessGame) Minimax(depth, stopDepth int, alpha, beta float64) (float64, int, int) {
	if depth >= stopDepth || c.EBE.Halfmoves >= 100 {
		return c.Evaluate(depth), 1, 0
	}

	moves := c.GetLegalMoves()
	if len(moves) == 0 {
		return c.Evaluate(depth), 1, 0
	}

	evaluated := 0
	skipped := 0
	checked := 0

	moves = c.PreOrder(moves)

	if c.EBE.Active<<3 == WHITE {
		value := math.Inf(-1)
		for _, move := range moves {
			select {
			case <-c.SearchTimer.C:
				fmt.Println("boop")
				return 0, -1, 0
			default:
				c.MakeMove(move)
				v, e, s := c.Minimax(depth+1, stopDepth, alpha, beta)
				c.UnmakeMove(move)
				if e == -1 {
					return 0, -1, 0
				}

				value = max(value, v)
				evaluated += e
				skipped += s

				checked += 1
				if value > beta {
					break
				}
				alpha = max(alpha, value)
			}
		}
		return value, evaluated, skipped + len(moves) - checked
	} else {
		value := math.Inf(1)
		for _, move := range moves {
			select {
			case <-c.SearchTimer.C:
				return 0, -1, 0
			default:
				c.MakeMove(move)
				v, e, s := c.Minimax(depth+1, stopDepth, alpha, beta)
				c.UnmakeMove(move)
				if e == -1 {
					return 0, -1, 0
				}

				value = min(value, v)
				evaluated += e
				skipped += s

				checked += 1
				if value < alpha {
					break
				}
				beta = min(beta, value)
			}
		}
		return value, evaluated, skipped + len(moves) - checked
	}
}

func (c *ChessGame) Evaluate(currentDepth int) float64 {
	t, ok := c.Transpositions[c.EBE.Board]
	if !ok || t.Depth < currentDepth {
		score := float64(c.Material(c.EBE.Active<<3) - c.Material(enemy(c.EBE.Active<<3)))
		if c.EBE.Active<<3 == BLACK {
			score *= -1
		}
		c.Transpositions[c.EBE.Board] = TranspositionNode{
			Depth: currentDepth,
			Value: score,
		}
		return score
	}

	return t.Value
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
