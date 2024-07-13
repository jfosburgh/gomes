package chess

import (
	// "fmt"
	"math"
)

type Move struct {
	Piece int
	Start int
	End   int

	Capture   int
	Castle    bool
	Promotion int

	Halfmoves       int
	CastlingRights  int
	EnPassantTarget int
}

func (c *ChessGame) GeneratePseudoLegal() []Move {
	// fmt.Printf("Generating moves for active player %d with board state\n%s\n", c.EBE.Active, c.EBE.Board)
	moves := []Move{}

	side := c.EBE.Active << 3
	moves = append(moves, c.GeneratePseudoLegalPawn(side)...)
	moves = append(moves, c.GeneratePseudoLegalRook(side)...)
	moves = append(moves, c.GeneratePseudoLegalKnight(side)...)
	moves = append(moves, c.GeneratePseudoLegalBishop(side)...)
	moves = append(moves, c.GeneratePseudoLegalQueen(side)...)
	moves = append(moves, c.GeneratePseudoLegalKing(side)...)

	return moves
}

func (c *ChessGame) GeneratePseudoLegalKing(side int) []Move {
	moves := []Move{}

	kingMoves := c.Bitboard.KingMoves(side)
	moveLocs := toPieceLocations(kingMoves)
	kingLoc := toPieceLocations(c.Bitboard[side|KING])[0]

	for _, moveLoc := range moveLocs {
		moves = append(moves, Move{
			Piece:   side | KING,
			Start:   kingLoc,
			End:     moveLoc,
			Capture: c.EBE.Board[moveLoc],

			Halfmoves:       c.EBE.Halfmoves,
			CastlingRights:  c.EBE.CastlingRights,
			EnPassantTarget: c.EBE.EnPassantTarget,
		})
	}

	castlingRights := c.EBE.CastlingRights >> 2
	if side == BLACK {
		castlingRights = c.EBE.CastlingRights & 0b0011
	}

	// kingside
	if castlingRights>>1 == 1 && c.EBE.Board[kingLoc+1] == EMPTY && c.EBE.Board[kingLoc+2] == EMPTY {
		moves = append(moves, Move{
			Piece:  side | KING,
			Start:  kingLoc,
			End:    kingLoc + 2,
			Castle: true,

			Halfmoves:       c.EBE.Halfmoves,
			CastlingRights:  c.EBE.CastlingRights,
			EnPassantTarget: c.EBE.EnPassantTarget,
		})
	}

	// queenside
	if castlingRights&0b01 == 1 && c.EBE.Board[kingLoc-1] == EMPTY && c.EBE.Board[kingLoc-2] == EMPTY && c.EBE.Board[kingLoc-3] == EMPTY {
		moves = append(moves, Move{
			Piece:  side | KING,
			Start:  kingLoc,
			End:    kingLoc - 2,
			Castle: true,

			Halfmoves:       c.EBE.Halfmoves,
			CastlingRights:  c.EBE.CastlingRights,
			EnPassantTarget: c.EBE.EnPassantTarget,
		})
	}

	return moves
}

func (c *ChessGame) GeneratePseudoLegalQueen(side int) []Move {
	moves := []Move{}

	queenMoves := c.Bitboard.QueenMoves(side)
	queenLocs := toPieceLocations(c.Bitboard[side|QUEEN])

	for _, queenLoc := range queenLocs {
		pieceMoves := queenMoves & (verticalCross(queenLoc) | diagonalCross(queenLoc))
		moveLocs := toPieceLocations(pieceMoves)

		for _, moveLoc := range moveLocs {
			moves = append(moves, Move{
				Piece:   side | QUEEN,
				Start:   queenLoc,
				End:     moveLoc,
				Capture: c.EBE.Board[moveLoc],

				Halfmoves:       c.EBE.Halfmoves,
				CastlingRights:  c.EBE.CastlingRights,
				EnPassantTarget: c.EBE.EnPassantTarget,
			})
		}
	}

	return moves
}

func (c *ChessGame) GeneratePseudoLegalBishop(side int) []Move {
	moves := []Move{}

	bishopMoves := c.Bitboard.BishopMoves(side)
	bishopLocs := toPieceLocations(c.Bitboard[side|BISHOP])

	for _, bishopLoc := range bishopLocs {
		pieceMoves := bishopMoves & diagonalCross(bishopLoc)
		moveLocs := toPieceLocations(pieceMoves)

		for _, moveLoc := range moveLocs {
			moves = append(moves, Move{
				Piece:   side | BISHOP,
				Start:   bishopLoc,
				End:     moveLoc,
				Capture: c.EBE.Board[moveLoc],

				Halfmoves:       c.EBE.Halfmoves,
				CastlingRights:  c.EBE.CastlingRights,
				EnPassantTarget: c.EBE.EnPassantTarget,
			})
		}
	}

	return moves
}

func (c *ChessGame) GeneratePseudoLegalRook(side int) []Move {
	moves := []Move{}

	rookMoves := c.Bitboard.RookMoves(side)
	rookLocs := toPieceLocations(c.Bitboard[side|ROOK])

	allPieces := c.Bitboard.AllPieces()

	for _, rookLoc := range rookLocs {
		pieceMoves := rookMoves & verticalCrossMasked(rookLoc, allPieces&(^(0b1<<rookLoc)))
		moveLocs := toPieceLocations(pieceMoves)

		for _, moveLoc := range moveLocs {
			moves = append(moves, Move{
				Piece:   side | ROOK,
				Start:   rookLoc,
				End:     moveLoc,
				Capture: c.EBE.Board[moveLoc],

				Halfmoves:       c.EBE.Halfmoves,
				CastlingRights:  c.EBE.CastlingRights,
				EnPassantTarget: c.EBE.EnPassantTarget,
			})
		}
	}

	return moves
}

func (c *ChessGame) GeneratePseudoLegalKnight(side int) []Move {
	moves := []Move{}

	knightMoves := c.Bitboard.KnightMoves(side)
	knightLocs := toPieceLocations(c.Bitboard[side|KNIGHT])

	for _, knightLoc := range knightLocs {
		pieceMoves := knightMoves & getKnightMoves(uint64(0b1<<knightLoc))
		moveLocs := toPieceLocations(pieceMoves)

		for _, moveLoc := range moveLocs {
			moves = append(moves, Move{
				Piece:   side | KNIGHT,
				Start:   knightLoc,
				End:     moveLoc,
				Capture: c.EBE.Board[moveLoc],

				Halfmoves:       c.EBE.Halfmoves,
				CastlingRights:  c.EBE.CastlingRights,
				EnPassantTarget: c.EBE.EnPassantTarget,
			})
		}
	}

	return moves
}

func (c *ChessGame) GeneratePseudoLegalPawn(side int) []Move {
	// fmt.Printf("Generating pawn moves for side %d\nCurrent Board:\n%s\n", side, c.EBE.Board)
	moves := []Move{}
	pawnThreatens, pawnMoves := c.Bitboard.PawnMoves(side)
	pawnAttacks := pawnThreatens & c.Bitboard.SidePieces(enemy(side))
	if c.EBE.EnPassantTarget != -1 {
		pawnAttacks = pawnAttacks | (pawnAttacks & (0b1 << c.EBE.EnPassantTarget))
	}
	// fmt.Printf("Pawn moves:\n%s\n", To2DString(pawnMoves))
	// fmt.Printf("Pawn attacks:\n%s\n", To2DString(pawnAttacks))
	// fmt.Printf("Pawn threatens:\n%s\n", To2DString(pawnThreatens))

	attackLocs := toPieceLocations(pawnAttacks)
	attackOrigins := [2]int{NORTHEAST, NORTHWEST}
	if side == WHITE {
		attackOrigins = [2]int{SOUTHEAST, SOUTHWEST}
	}

	// fmt.Printf("generating attacks at locations %+v\n", attackLocs)

	for _, attackLoc := range attackLocs {
		for _, attackOrigin := range attackOrigins {
			if attackLoc%8 == 7 && (attackOrigin == NORTHEAST || attackOrigin == SOUTHEAST) {
				continue
			}
			if attackLoc%8 == 0 && (attackOrigin == NORTHWEST || attackOrigin == SOUTHWEST) {
				continue
			}

			if c.EBE.Board[attackLoc+attackOrigin] == side|PAWN {
				// fmt.Printf("adding pawn advance from %s (%d) to %s (%d)\n", int2algebraic(attackLoc+attackOrigin), attackLoc+attackOrigin, int2algebraic(attackLoc), attackLoc)
				if attackLoc >= 56 || attackLoc <= 7 {
					for _, promotion := range []int{KNIGHT, BISHOP, ROOK, QUEEN} {
						moves = append(moves, Move{
							Piece:     side | PAWN,
							Start:     attackLoc + attackOrigin,
							End:       attackLoc,
							Capture:   c.EBE.Board[attackLoc],
							Promotion: promotion,

							Halfmoves:       c.EBE.Halfmoves,
							CastlingRights:  c.EBE.CastlingRights,
							EnPassantTarget: c.EBE.EnPassantTarget,
						})
					}
				} else {
					moves = append(moves, Move{
						Piece:   side | PAWN,
						Start:   attackLoc + attackOrigin,
						End:     attackLoc,
						Capture: c.EBE.Board[attackLoc],

						Halfmoves:       c.EBE.Halfmoves,
						CastlingRights:  c.EBE.CastlingRights,
						EnPassantTarget: c.EBE.EnPassantTarget,
					})
				}
			}
		}
	}

	moveLocs := toPieceLocations(pawnMoves)
	moveOrigins := [2]int{NORTH, NORTH + NORTH}
	if side == WHITE {
		moveOrigins = [2]int{SOUTH, SOUTH + SOUTH}
	}

	// fmt.Printf("generating moves at locations %+v\n", moveLocs)
	for _, moveLoc := range moveLocs {
		for _, moveOrigin := range moveOrigins {
			if moveOrigin == SOUTH+SOUTH && moveLoc/8 != 3 {
				continue
			}
			if moveOrigin == NORTH+NORTH && moveLoc/8 != 4 {
				continue
			}
			if c.EBE.Board[moveLoc+moveOrigin] == side|PAWN {
				if moveLoc >= 56 || moveLoc <= 7 {
					for _, promotion := range []int{KNIGHT, BISHOP, ROOK, QUEEN} {
						moves = append(moves, Move{
							Piece:     side | PAWN,
							Start:     moveLoc + moveOrigin,
							End:       moveLoc,
							Promotion: promotion,

							Halfmoves:       c.EBE.Halfmoves,
							CastlingRights:  c.EBE.CastlingRights,
							EnPassantTarget: c.EBE.EnPassantTarget,
						})
					}
				} else {
					// fmt.Printf("adding pawn advance from %s (%d) to %s (%d)\n", int2algebraic(moveLoc+moveOrigin), moveLoc+moveOrigin, int2algebraic(moveLoc), moveLoc)
					moves = append(moves, Move{
						Piece: side | PAWN,
						Start: moveLoc + moveOrigin,
						End:   moveLoc,

						Halfmoves:       c.EBE.Halfmoves,
						CastlingRights:  c.EBE.CastlingRights,
						EnPassantTarget: c.EBE.EnPassantTarget,
					})
				}
			}
		}
	}

	return moves
}

func (c *ChessGame) MakeMove(move Move) {
	c.EBE.Board[move.Start] = EMPTY
	c.EBE.Board[move.End] = move.Piece

	c.Bitboard.Add(move.Piece, move.End)
	c.Bitboard.Remove(move.Piece, move.Start)

	if move.Capture != 0 {
		c.Captured = append(c.Captured, move.Capture)
		c.Bitboard.Remove(move.Capture, move.End)
	}

	if move.Promotion != 0 {
		c.EBE.Board[move.End] = move.Promotion
		c.Bitboard.Remove(move.Piece, move.End)
		c.Bitboard.Add(move.Promotion, move.End)
	}

	if move.Castle {
		switch {
		// white king side
		case move.Piece>>3 == 0 && move.End == 6:
			c.EBE.Board[7] = EMPTY
			c.EBE.Board[5] = WHITE | ROOK

			c.Bitboard.Remove(7, WHITE|ROOK)
			c.Bitboard.Add(5, WHITE|ROOK)
		// white queen side
		case move.Piece>>3 == 0 && move.End == 2:
			c.EBE.Board[0] = EMPTY
			c.EBE.Board[3] = WHITE | ROOK

			c.Bitboard.Remove(0, WHITE|ROOK)
			c.Bitboard.Add(3, WHITE|ROOK)
		// black king side
		case move.Piece>>3 == 1 && move.End == 62:
			c.EBE.Board[63] = EMPTY
			c.EBE.Board[61] = BLACK | ROOK

			c.Bitboard.Remove(63, BLACK|ROOK)
			c.Bitboard.Add(61, BLACK|ROOK)
		// black queen side
		case move.Piece>>3 == 1 && move.End == 58:
			c.EBE.Board[56] = EMPTY
			c.EBE.Board[59] = WHITE | ROOK

			c.Bitboard.Remove(56, BLACK|ROOK)
			c.Bitboard.Add(59, BLACK|ROOK)
		}
	}

	c.Moves = append(c.Moves, move)

	if c.EBE.Active<<3 == BLACK {
		c.EBE.Moves += 1
	}
	c.EBE.Active = ^c.EBE.Active & 0b1

	if move.Capture == 0 && move.Piece&0b0111 != PAWN {
		c.EBE.Halfmoves += 1
	} else {
		c.EBE.Halfmoves = 0
	}

	if c.EBE.CastlingRights != 0 {
		if move.Piece == WHITE|KING {
			c.EBE.CastlingRights = c.EBE.CastlingRights & 0b0011
		}

		if move.Piece == BLACK|KING {
			c.EBE.CastlingRights = c.EBE.CastlingRights & 0b1100
		}

		if move.Piece == WHITE|ROOK && move.Start == 7 {
			c.EBE.CastlingRights = c.EBE.CastlingRights & 0b0111
		}

		if move.Piece == WHITE|ROOK && move.Start == 0 {
			c.EBE.CastlingRights = c.EBE.CastlingRights & 0b1011
		}

		if move.Piece == BLACK|ROOK && move.Start == 63 {
			c.EBE.CastlingRights = c.EBE.CastlingRights & 0b1101
		}

		if move.Piece == BLACK|ROOK && move.Start == 56 {
			c.EBE.CastlingRights = c.EBE.CastlingRights & 0b1110
		}
	}

	if move.Piece&0b0111 == PAWN && math.Abs(float64(move.End)-float64(move.Start)) == 16 {
		c.EBE.EnPassantTarget = move.Start + (move.End-move.Start)/2
	} else {
		c.EBE.EnPassantTarget = -1
	}
}

func (c *ChessGame) UnmakeMove(move Move) {
	c.EBE.Board[move.Start] = move.Piece
	c.EBE.Board[move.End] = EMPTY

	c.Bitboard.Add(move.Piece, move.Start)
	c.Bitboard.Remove(move.Piece, move.End)

	if move.Castle {
		switch {
		// white king side
		case move.Piece>>3 == 0 && move.End == 6:
			c.EBE.Board[5] = EMPTY
			c.EBE.Board[7] = WHITE | ROOK

			c.Bitboard.Remove(5, WHITE|ROOK)
			c.Bitboard.Add(7, WHITE|ROOK)
		// white queen side
		case move.Piece>>3 == 0 && move.End == 2:
			c.EBE.Board[3] = EMPTY
			c.EBE.Board[0] = WHITE | ROOK

			c.Bitboard.Remove(3, WHITE|ROOK)
			c.Bitboard.Add(0, WHITE|ROOK)
		// black king side
		case move.Piece>>3 == 1 && move.End == 62:
			c.EBE.Board[61] = EMPTY
			c.EBE.Board[63] = BLACK | ROOK

			c.Bitboard.Remove(61, BLACK|ROOK)
			c.Bitboard.Add(63, BLACK|ROOK)
		// black queen side
		case move.Piece>>3 == 1 && move.End == 58:
			c.EBE.Board[59] = EMPTY
			c.EBE.Board[56] = WHITE | ROOK

			c.Bitboard.Remove(59, BLACK|ROOK)
			c.Bitboard.Add(56, BLACK|ROOK)
		}
	}

	if move.Capture != 0 {
		c.EBE.Board[move.End] = move.Capture
		c.Bitboard.Add(move.Capture, move.End)
		c.Captured = c.Captured[:len(c.Captured)-1]
	}

	c.Moves = c.Moves[:len(c.Moves)-1]
	if c.EBE.Active<<3 == WHITE {
		c.EBE.Moves -= 1
	}
	c.EBE.Active = ^c.EBE.Active & 0b01

	// reset cached values
	c.EBE.Halfmoves = move.Halfmoves
	c.EBE.CastlingRights = move.CastlingRights
	c.EBE.EnPassantTarget = move.EnPassantTarget
}
