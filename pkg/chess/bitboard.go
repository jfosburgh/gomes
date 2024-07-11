package chess

import (
	"fmt"
	"strings"
)

type BitBoard map[int]uint64

func (b BitBoard) FromEBE(ebe EBEBoard) {
	for piece := range piece2String {
		if piece != EMPTY {
			b[piece] = 0
		}
	}

	for i := range ebe {
		piece := ebe[i]
		if piece != EMPTY {
			b[piece] = b[piece] | (0b1 << i)
		}
	}
}

func To2DString(board uint64) string {
	oneD := fmt.Sprintf("%064b", board)
	twoDArray := []string{}
	for i := range 8 {
		twoDArray = append(twoDArray, oneD[i*8:(i+1)*8])
	}

	return strings.Join(twoDArray, "\n")
}

func (b BitBoard) AllPieces() uint64 {
	result := uint64(0)
	for _, pieceBoard := range b {
		result = result | pieceBoard
	}

	return result
}

func (b BitBoard) SidePieces(side int) uint64 {
	result := uint64(0)
	for piece, pieceBoard := range b {
		if (piece >> 3) == (side >> 3) {
			result = result | pieceBoard
		}
	}

	return result
}

func (b BitBoard) PawnMoves(side int) (uint64, uint64) {
	enemyBitboard := b.SidePieces(((^side >> 3) & 0b1) << 3)
	selfBitboard := b.SidePieces(side)
	validTargets := enemyBitboard & (^selfBitboard)

	potentialAttacks := uint64(0)
	if side == WHITE {
		potentialAttacks = (^fileMask(1) & b[side|PAWN]) << NORTHWEST
		potentialAttacks = potentialAttacks | ((^fileMask(8) & b[side|PAWN]) << NORTHEAST)
	} else {
		potentialAttacks = (^fileMask(1) & b[side|PAWN]) >> NORTHWEST
		potentialAttacks = potentialAttacks | ((^fileMask(8) & b[side|PAWN]) >> NORTHEAST)
	}

	var singleAdvance, doubleAdvance uint64
	pawns := b[side|PAWN]
	doubleAdvanceable := pawns & (rankMask(2) | rankMask(7))
	if side == WHITE {
		singleAdvance = pawns << 8
		doubleAdvance = doubleAdvanceable << 16
	} else {
		singleAdvance = pawns >> 8
		doubleAdvance = doubleAdvanceable >> 16
	}

	return potentialAttacks & validTargets, (singleAdvance | doubleAdvance) & (^b.AllPieces())
}

func (b BitBoard) Remove(piece, position int) {
	b[piece] = b[piece] & (^(0b1 << position))
}

func (b BitBoard) Add(piece, position int) {
	b[piece] = b[piece] | (0b1 << position)
}

func fileMask(file int) uint64 {
	res := uint64(0)
	for i := range 8 {
		res = res | 1<<(i*8+file-1)
	}

	return res
}

func rankMask(rank int) uint64 {
	return 0b11111111 << ((rank - 1) * 8)
}
