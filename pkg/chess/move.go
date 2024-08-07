package chess

import (
	"fmt"
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

func (m Move) String() string {
	s := fmt.Sprintf("%s%s", int2algebraic(m.Start), int2algebraic(m.End))
	if m.Promotion != 0 {
		s += piece2String[m.Promotion]
	}

	return s
}

// func (c *ChessGame) MoveFromLocations(start, end, promotion int) Move {
// 	move := Move{
// 		Piece: c.EBE.Board[start],
// 		Start: start,
// 		End: end,
//
// 		Halfmoves: c.EBE.Halfmoves,
// 		CastlingRights: c.EBE.CastlingRights,
// 		EnPassantTarget: c.EBE.EnPassantTarget,
// 	}
//
// 	return move
// }
//
// func (c *ChessGame) MoveFromAlgebraic(start, end string) Move {
// 	move := Move{}
//
// 	return move
// }

func (c *ChessGame) GeneratePseudoLegal() []Move {
	// fmt.Printf("Generating moves for active player %d and castling rights %04b with board state\n%s\n", c.EBE.Active, c.EBE.CastlingRights, c.EBE.Board)
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

	if c.EBE.CastlingRights == 0 || c.Bitboard.InCheck(side) {
		return moves
	}

	castlingRights := c.EBE.CastlingRights >> 2
	if side == BLACK {
		castlingRights = c.EBE.CastlingRights & 0b0011
	}

	threatened := c.Bitboard.SideThreatens(enemy(side))

	// kingside
	if castlingRights>>1 == 1 && c.EBE.Board[kingLoc+1] == EMPTY && c.EBE.Board[kingLoc+2] == EMPTY && threatened&((0b1<<(kingLoc+1))|(0b1<<(kingLoc+2))) == 0 {
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
	if castlingRights&0b01 == 1 && c.EBE.Board[kingLoc-1] == EMPTY && c.EBE.Board[kingLoc-2] == EMPTY && c.EBE.Board[kingLoc-3] == EMPTY && threatened&((0b1<<(kingLoc-1))|(0b1<<(kingLoc-2))) == 0 {
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
	queenLocs := toPieceLocations(c.Bitboard[side|QUEEN])
	enemyBoard, selfBoard := c.Bitboard[enemy(side)], c.Bitboard[side]

	for _, queenLoc := range queenLocs {
		queenMoves := getQueenMoves(enemyBoard, selfBoard, 0b1<<queenLoc)
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
	bishopLocs := toPieceLocations(c.Bitboard[side|BISHOP])
	enemyBoard, selfBoard := c.Bitboard[enemy(side)], c.Bitboard[side]

	for _, bishopLoc := range bishopLocs {
		bishopMoves := getBishopMoves(enemyBoard, selfBoard, 0b1<<bishopLoc)
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
		// fmt.Printf("adding en passant target at %s\n", int2algebraic(c.EBE.EnPassantTarget))
		pawnAttacks = pawnAttacks | (pawnThreatens & (0b1 << c.EBE.EnPassantTarget))
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
							Promotion: side | promotion,

							Halfmoves:       c.EBE.Halfmoves,
							CastlingRights:  c.EBE.CastlingRights,
							EnPassantTarget: c.EBE.EnPassantTarget,
						})
					}
				} else {
					captured := c.EBE.Board[attackLoc]
					if attackLoc == c.EBE.EnPassantTarget {
						if side == WHITE {
							captured = c.EBE.Board[attackLoc-8]
						} else {
							captured = c.EBE.Board[attackLoc+8]
						}
					}
					moves = append(moves, Move{
						Piece:   side | PAWN,
						Start:   attackLoc + attackOrigin,
						End:     attackLoc,
						Capture: captured,

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
			if moveOrigin == SOUTH+SOUTH && (moveLoc/8 != 3 || c.EBE.Board[moveLoc+SOUTH] != EMPTY) {
				continue
			}
			if moveOrigin == NORTH+NORTH && (moveLoc/8 != 4 || c.EBE.Board[moveLoc+NORTH] != EMPTY) {
				continue
			}
			if c.EBE.Board[moveLoc+moveOrigin] == side|PAWN {
				if moveLoc >= 56 || moveLoc <= 7 {
					for _, promotion := range []int{KNIGHT, BISHOP, ROOK, QUEEN} {
						moves = append(moves, Move{
							Piece:     side | PAWN,
							Start:     moveLoc + moveOrigin,
							End:       moveLoc,
							Promotion: side | promotion,

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

func (c *ChessGame) RemovePiece(piece, location int) {
	c.EBE.Board[location] = EMPTY
	c.Bitboard.Remove(piece, location)
}

func (c *ChessGame) PlacePiece(piece, location int) {
	c.EBE.Board[location] = piece
	c.Bitboard.Add(piece, location)
}

func (c *ChessGame) ReplacePiece(oldPiece, newPiece, location int) {
	c.EBE.Board[location] = newPiece
	c.Bitboard.Remove(oldPiece, location)
	c.Bitboard.Add(newPiece, location)
}

func (c *ChessGame) MakeMove(move Move) {
	c.RemovePiece(move.Piece, move.Start)
	pieceToPlace := move.Piece
	if move.Promotion != 0 {
		pieceToPlace = move.Promotion
	}

	if move.Capture == 0 {
		c.PlacePiece(pieceToPlace, move.End)
	} else {
		c.Captured = append(c.Captured, move.Capture)
		if move.EnPassantTarget == move.End {
			c.PlacePiece(pieceToPlace, move.End)
			if c.EBE.Active == 0 {
				// fmt.Printf("capturing en passant target at %s in move %s with board state \n%s\n", int2algebraic(move.End-8), move, c.EBE.Board)
				c.RemovePiece(move.Capture, move.End-8)
				// fmt.Printf("final state:\n%s\n", c.EBE.Board)
			} else {
				// fmt.Printf("capturing en passant target at %s in move %s with board state \n%s\n", int2algebraic(move.End+8), move, c.EBE.Board)
				c.RemovePiece(move.Capture, move.End+8)
				// fmt.Printf("final state:\n%s\n", c.EBE.Board)
			}
		} else {
			c.ReplacePiece(move.Capture, pieceToPlace, move.End)
		}
	}

	if move.Castle {
		switch {
		// white king side
		case move.Piece>>3 == 0 && move.End == 6:
			c.RemovePiece(WHITE|ROOK, 7)
			c.PlacePiece(WHITE|ROOK, 5)
		// white queen side
		case move.Piece>>3 == 0 && move.End == 2:
			c.RemovePiece(WHITE|ROOK, 0)
			c.PlacePiece(WHITE|ROOK, 3)
		// black king side
		case move.Piece>>3 == 1 && move.End == 62:
			c.RemovePiece(BLACK|ROOK, 63)
			c.PlacePiece(BLACK|ROOK, 61)
		// black queen side
		case move.Piece>>3 == 1 && move.End == 58:
			c.RemovePiece(BLACK|ROOK, 56)
			c.PlacePiece(BLACK|ROOK, 59)
		}
	}

	c.Moves = append(c.Moves, move)

	if c.EBE.Active<<3 == BLACK {
		c.EBE.Moves += 1
	}

	c.Bitboard.UpdateSide(c.EBE.Active << 3)
	c.EBE.Active = ^c.EBE.Active & 0b1
	if move.Capture != 0 {
		c.Bitboard.UpdateSide(c.EBE.Active << 3)
	}

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

		if move.End == 7 {
			c.EBE.CastlingRights = c.EBE.CastlingRights & 0b0111
		}

		if move.End == 0 {
			c.EBE.CastlingRights = c.EBE.CastlingRights & 0b1011
		}

		if move.End == 56 {
			c.EBE.CastlingRights = c.EBE.CastlingRights & 0b1110
		}

		if move.End == 63 {
			c.EBE.CastlingRights = c.EBE.CastlingRights & 0b1101
		}
	}

	if move.Piece&0b0111 == PAWN && math.Abs(float64(move.End)-float64(move.Start)) == 16 {
		c.EBE.EnPassantTarget = move.Start + (move.End-move.Start)/2
	} else {
		c.EBE.EnPassantTarget = -1
	}
}

func (c *ChessGame) UnmakeMove(move Move) {
	c.PlacePiece(move.Piece, move.Start)
	pieceToRemove := move.Piece
	if move.Promotion != 0 {
		pieceToRemove = move.Promotion
	}

	if move.Capture == 0 {
		c.RemovePiece(pieceToRemove, move.End)
	} else {
		if move.End == move.EnPassantTarget {
			c.RemovePiece(pieceToRemove, move.End)
			if c.EBE.Active == 1 {
				// fmt.Printf("replacing en passant target at %s in move %s with board state \n%s\n", int2algebraic(move.End-8), move, c.EBE.Board)
				c.PlacePiece(move.Capture, move.End-8)
				// fmt.Printf("final state:\n%s\n", c.EBE.Board)
			} else {
				// fmt.Printf("replacing en passant target at %s in move %s with board state \n%s\n", int2algebraic(move.End+8), move, c.EBE.Board)
				c.PlacePiece(move.Capture, move.End+8)
				// fmt.Printf("final state:\n%s\n", c.EBE.Board)
			}
		} else {
			c.ReplacePiece(pieceToRemove, move.Capture, move.End)
		}
		c.Captured = c.Captured[:len(c.Captured)-1]
	}

	if move.Castle {
		switch {
		// white king side
		case move.Piece>>3 == 0 && move.End == 6:
			c.RemovePiece(WHITE|ROOK, 5)
			c.PlacePiece(WHITE|ROOK, 7)
		// white queen side
		case move.Piece>>3 == 0 && move.End == 2:
			c.RemovePiece(WHITE|ROOK, 3)
			c.PlacePiece(WHITE|ROOK, 0)
		// black king side
		case move.Piece>>3 == 1 && move.End == 62:
			c.RemovePiece(BLACK|ROOK, 61)
			c.PlacePiece(BLACK|ROOK, 63)
		// black queen side
		case move.Piece>>3 == 1 && move.End == 58:
			c.RemovePiece(BLACK|ROOK, 59)
			c.PlacePiece(BLACK|ROOK, 56)
		}
	}

	c.Moves = c.Moves[:len(c.Moves)-1]
	if c.EBE.Active<<3 == WHITE {
		c.EBE.Moves -= 1
	}

	if move.Capture != 0 {
		c.Bitboard.UpdateSide(c.EBE.Active << 3)
	}
	c.EBE.Active = ^c.EBE.Active & 0b01
	c.Bitboard.UpdateSide(c.EBE.Active << 3)

	// reset cached values
	c.EBE.Halfmoves = move.Halfmoves
	c.EBE.CastlingRights = move.CastlingRights
	c.EBE.EnPassantTarget = move.EnPassantTarget
}
