package chess

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	StartingBoard = EBEBoard{
		WHITE | ROOK, WHITE | KNIGHT, WHITE | BISHOP, WHITE | QUEEN, WHITE | KING, WHITE | BISHOP, WHITE | KNIGHT, WHITE | ROOK,
		WHITE | PAWN, WHITE | PAWN, WHITE | PAWN, WHITE | PAWN, WHITE | PAWN, WHITE | PAWN, WHITE | PAWN, WHITE | PAWN,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY, EMPTY,
		BLACK | PAWN, BLACK | PAWN, BLACK | PAWN, BLACK | PAWN, BLACK | PAWN, BLACK | PAWN, BLACK | PAWN, BLACK | PAWN,
		BLACK | ROOK, BLACK | KNIGHT, BLACK | BISHOP, BLACK | QUEEN, BLACK | KING, BLACK | BISHOP, BLACK | KNIGHT, BLACK | ROOK,
	}
)

type EBE struct {
	Board           EBEBoard
	Active          int
	CastlingRights  int
	EnPassantTarget int
	Halfmoves       int
	Moves           int
}

type EBEBoard [64]int

func (b EBEBoard) String() string {
	board := "-----------------"
	for rank := 7; rank >= 0; rank-- {
		board += "\n|"
		for file := 0; file <= 7; file++ {
			board += fmt.Sprintf("%s|", piece2String[b[rank*8+file]])
		}
		board += "\n-----------------"
	}

	return board
}

func DefaultBoard() EBE {
	return EBE{
		Board:           StartingBoard,
		Active:          0b0,
		CastlingRights:  0b1111,
		EnPassantTarget: -1,
		Halfmoves:       0,
		Moves:           1,
	}
}

func (b *EBE) ToFEN() string {
	fen := ""

	empty := 0
	rank := 7
	file := 0

	for rank*8+file >= 0 {
		switch b.Board[rank*8+file] {
		case EMPTY:
			empty += 1
		default:
			if empty > 0 {
				fen += strconv.Itoa(empty)
				empty = 0
			}
			fen += piece2String[b.Board[rank*8+file]]
		}

		file += 1
		if file > 7 {
			rank -= 1
			file = 0

			if empty > 0 {
				fen += strconv.Itoa(empty)
				empty = 0
			}
			fen += "/"
		}
	}
	fen = fen[:len(fen)-1]

	fen += " "
	if b.Active == 0 {
		fen += "w "
	} else {
		fen += "b "
	}

	fen += castlingRightsToString(b.CastlingRights)

	fen += " "

	if b.EnPassantTarget == -1 {
		fen += "-"
	} else {
		fen += int2algebraic(b.EnPassantTarget)
	}

	fen = fmt.Sprintf("%s %d %d", fen, b.Halfmoves, b.Moves)

	return fen
}

func castlingRightsToString(castlingRights int) string {
	s := ""
	if castlingRights == 0 {
		return "-"
	}

	if (castlingRights>>3)&0b1 == 1 {
		s += "K"
	}
	if (castlingRights>>2)&0b1 == 1 {
		s += "Q"
	}
	if (castlingRights>>1)&0b1 == 1 {
		s += "k"
	}
	if castlingRights&0b1 == 1 {
		s += "q"
	}

	return s
}

func (b *EBE) FromFEN(fen string) {
	fenParts := strings.Split(fen, " ")
	rank := 7
	file := 0

	b.Board = [64]int{}

	placements := strings.Split(fenParts[0], "")
	for i := range placements {
		if placements[i] == "/" {
			rank -= 1
			file = 0
			continue
		}

		empty, ok := strconv.Atoi(placements[i])
		if ok == nil {
			file += empty
			continue
		}

		b.Board[rank*8+file] = string2Piece[placements[i]]
		file += 1
	}

	if fenParts[1] == "w" {
		b.Active = 0b0
	} else {
		b.Active = 0b1
	}

	castlingChars := strings.Split(fenParts[2], "")
	for _, char := range castlingChars {
		switch char {
		case "-":
			break
		case "K":
			b.CastlingRights = b.CastlingRights | (0b1 << 3)
		case "Q":
			b.CastlingRights = b.CastlingRights | (0b1 << 2)
		case "k":
			b.CastlingRights = b.CastlingRights | (0b1 << 1)
		case "q":
			b.CastlingRights = b.CastlingRights | 0b1
		}
	}

	enPassantPos := fenParts[3]
	if enPassantPos == "-" {
		b.EnPassantTarget = -1
	} else {
		b.EnPassantTarget = algebraic2Int(enPassantPos)
	}

	halfmoves, _ := strconv.Atoi(fenParts[4])
	b.Halfmoves = halfmoves

	moves, _ := strconv.Atoi(fenParts[5])
	b.Moves = moves
}
