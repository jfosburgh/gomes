package chess

import "math"

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

func (c *ChessGame) MakeMove(move Move) {
	c.EBE.Board[move.Start] = EMPTY
	c.EBE.Board[move.End] = move.Piece

	c.Bitboard.Add(move.Piece, move.End)
	c.Bitboard.Remove(move.Piece, move.Start)

	if move.Promotion != 0 {
		c.EBE.Board[move.End] = move.Promotion
		c.Bitboard.Remove(move.Piece, move.End)
		c.Bitboard.Add(move.Promotion, move.End)
	}

	if move.Capture != 0 {
		c.Captured = append(c.Captured, move.Capture)
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
	c.EBE.Active = ^c.EBE.Active & 0b01

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
