package chess

import (
	"fmt"
	"strconv"
)

const (
	WHITE = 0b0000
	BLACK = 0b1000

	EMPTY  = 0b0000
	PAWN   = 0b0001
	KNIGHT = 0b0010
	BISHOP = 0b0011
	ROOK   = 0b0100
	QUEEN  = 0b0101
	KING   = 0b0110

	StartingFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
)

var (
	piece2String = map[int]string{
		BLACK | PAWN:   "p",
		BLACK | KNIGHT: "n",
		BLACK | BISHOP: "b",
		BLACK | ROOK:   "r",
		BLACK | QUEEN:  "q",
		BLACK | KING:   "k",
		WHITE | PAWN:   "P",
		WHITE | KNIGHT: "N",
		WHITE | BISHOP: "B",
		WHITE | ROOK:   "R",
		WHITE | QUEEN:  "Q",
		WHITE | KING:   "K",
		EMPTY:          " ",
	}

	string2Piece = map[string]int{
		"p": BLACK | PAWN,
		"n": BLACK | KNIGHT,
		"b": BLACK | BISHOP,
		"r": BLACK | ROOK,
		"q": BLACK | QUEEN,
		"k": BLACK | KING,
		"P": WHITE | PAWN,
		"N": WHITE | KNIGHT,
		"B": WHITE | BISHOP,
		"R": WHITE | ROOK,
		"Q": WHITE | QUEEN,
		"K": WHITE | KING,
		" ": EMPTY,
	}
)

func algebraic2Int(pos [2]byte) int {
	file := int(rune(pos[0]) - 'a')
	rank, _ := strconv.Atoi(string(pos[1]))

	return rank*8 + file
}

func int2algebraic(pos int) string {
	rank := pos / 8
	file := rank % 8

	alg := fmt.Sprint('a' + file)
	return alg + fmt.Sprint(rank)
}
