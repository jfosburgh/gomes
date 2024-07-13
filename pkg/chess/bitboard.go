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
	selfBitboard := b.SidePieces(side)

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

	return potentialAttacks & (^selfBitboard), (singleAdvance | doubleAdvance) & (^b.AllPieces())
}

func (b BitBoard) KnightMoves(side int) uint64 {
	selfBitboard := b.SidePieces(side)

	moves := uint64(0)
	//ENE & ESE
	moves = moves | ((b[KNIGHT|side] & (^(rankMask(8) | fileMask(7) | fileMask(8)))) << (EAST + NORTHEAST))
	moves = moves | ((b[KNIGHT|side] & (^(rankMask(1) | fileMask(7) | fileMask(8)))) >> (WEST + NORTHWEST))

	//NNE & SSE
	moves = moves | ((b[KNIGHT|side] & (^(rankMask(8) | rankMask(7) | fileMask(8)))) << (NORTH + NORTHEAST))
	moves = moves | ((b[KNIGHT|side] & (^(rankMask(1) | rankMask(2) | fileMask(8)))) >> (NORTH + NORTHWEST))

	//WNW & WSW
	moves = moves | ((b[KNIGHT|side] & (^(rankMask(8) | fileMask(1) | fileMask(2)))) << (WEST + NORTHWEST))
	moves = moves | ((b[KNIGHT|side] & (^(rankMask(1) | fileMask(1) | fileMask(2)))) >> (EAST + NORTHEAST))

	//NNE & SSE
	moves = moves | ((b[KNIGHT|side] & (^(rankMask(8) | rankMask(7) | fileMask(1)))) << (NORTH + NORTHWEST))
	moves = moves | ((b[KNIGHT|side] & (^(rankMask(1) | rankMask(2) | fileMask(1)))) >> (NORTH + NORTHEAST))

	moves = moves & (^selfBitboard)

	return moves
}

func (b BitBoard) KingMoves(side int) uint64 {
	selfBitboard := b.SidePieces(side)

	moves := uint64(0)
	moves = moves | ((b[KING|side] & (^rankMask(8))) << NORTH)
	moves = moves | ((b[KING|side] & (^rankMask(1))) >> NORTH)
	moves = moves | ((b[KING|side] & (^fileMask(8))) << EAST)
	moves = moves | ((b[KING|side] & (^fileMask(1))) >> EAST)
	moves = moves | ((b[KING|side] & (^(rankMask(8) | fileMask(8)))) << NORTHEAST)
	moves = moves | ((b[KING|side] & (^(rankMask(1) | fileMask(8)))) >> NORTHWEST)
	moves = moves | ((b[KING|side] & (^(rankMask(8) | fileMask(1)))) << NORTHWEST)
	moves = moves | ((b[KING|side] & (^(rankMask(1) | fileMask(1)))) >> NORTHEAST)

	moves = moves & (^selfBitboard)

	return moves
}

func (b BitBoard) RookMoves(side int) uint64 {
	enemyBitboard := b.SidePieces(((^side >> 3) & 0b1) << 3)
	selfBitboard := b.SidePieces(side)

	moves := uint64(0)
	locs := toPieceLocations(b[side|ROOK])
	for _, loc := range locs {
		moves = moves | verticalCrossMasked(loc, (enemyBitboard|selfBitboard)&(^(0b1<<loc)))
	}

	moves = moves & (^selfBitboard)

	return moves
}

func (b BitBoard) BishopMoves(side int) uint64 {
	enemyBitboard := b.SidePieces(((^side >> 3) & 0b1) << 3)
	selfBitboard := b.SidePieces(side)

	moves := uint64(0)
	locs := toPieceLocations(b[side|BISHOP])
	for _, loc := range locs {
		moves = moves | diagonalCrossMasked(loc, (enemyBitboard|selfBitboard)&(^(0b1<<loc)))
	}

	moves = moves & (^selfBitboard)

	return moves
}

func (b BitBoard) QueenMoves(side int) uint64 {
	enemyBitboard := b.SidePieces(((^side >> 3) & 0b1) << 3)
	selfBitboard := b.SidePieces(side)

	moves := uint64(0)
	locs := toPieceLocations(b[side|QUEEN])
	for _, loc := range locs {
		moves = moves | diagonalCrossMasked(loc, (enemyBitboard|selfBitboard)&(^(0b1<<loc))) | verticalCrossMasked(loc, (enemyBitboard|selfBitboard)&(^(0b1<<loc)))
	}

	moves = moves & (^selfBitboard)

	return moves
}

func (b BitBoard) Remove(piece, position int) {
	b[piece] = b[piece] & (^(0b1 << position))
}

func (b BitBoard) Add(piece, position int) {
	b[piece] = b[piece] | (0b1 << position)
}

func To2DString(board uint64) string {
	oneD := fmt.Sprintf("%064b", board)
	twoDArray := []string{}
	for i := range 8 {
		row := ""
		for j := (i+1)*8 - 1; j >= i*8; j-- {
			row += string(oneD[j])
		}
		twoDArray = append(twoDArray, row)
	}

	return strings.Join(twoDArray, "\n")
}

func toPieceLocations(bitboard uint64) []int {
	locations := []int{}
	shift := 0

	for shift < 63 {
		if (bitboard>>shift)&0b1 == 1 {
			locations = append(locations, shift)
		}

		shift += 1
	}

	return locations
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

func verticalCross(pos int) uint64 {
	rank := pos/8 + 1
	file := pos%8 + 1

	return rankMask(rank) | fileMask(file)
}

func verticalCrossMasked(pos int, pieces uint64) uint64 {
	rank := pos / 8
	file := pos % 8

	res := uint64(0)
	r, f := rank, file
	for {
		res = res | (0b1 << (r*8 + f))
		r += 1

		if r > 7 || (pieces>>((r-1)*8+f)&0b1 == 1) {
			break
		}
	}

	r, f = rank, file
	for {
		res = res | (0b1 << (r*8 + f))
		r -= 1

		if r < 0 || (pieces>>((r+1)*8+f)&0b1 == 1) {
			break
		}
	}

	r, f = rank, file
	for {
		res = res | (0b1 << (r*8 + f))
		f += 1

		if f > 7 || (pieces>>(r*8+f-1)&0b1 == 1) {
			break
		}
	}

	r, f = rank, file
	for {
		res = res | (0b1 << (r*8 + f))
		f -= 1

		if f < 0 || (pieces>>(r*8+f+1)&0b1 == 1) {
			break
		}
	}

	return res
}

func diagonalCross(pos int) uint64 {
	res := uint64(0)

	rank := pos / 8
	file := pos % 8

	r := rank
	f := file
	for {
		if r < 0 || f < 0 {
			break
		}

		res = res | (0b1 << (r*8 + f))
		r -= 1
		f -= 1
	}

	r = rank
	f = file
	for {
		if r < 0 || f > 7 {
			break
		}

		res = res | (0b1 << (r*8 + f))
		r -= 1
		f += 1
	}

	r = rank
	f = file
	for {
		if r > 7 || f > 7 {
			break
		}

		res = res | (0b1 << (r*8 + f))
		r += 1
		f += 1
	}

	r = rank
	f = file
	for {
		if r > 7 || f < 0 {
			break
		}

		res = res | (0b1 << (r*8 + f))
		r += 1
		f -= 1
	}

	return res
}

func diagonalCrossMasked(pos int, pieces uint64) uint64 {
	res := uint64(0)

	rank := pos / 8
	file := pos % 8

	r := rank
	f := file
	for {
		res = res | (0b1 << (r*8 + f))
		r -= 1
		f -= 1

		if r < 0 || f < 0 || (pieces>>((r+1)*8+f+1)&0b1 == 1) {
			break
		}
	}

	r = rank
	f = file
	for {
		res = res | (0b1 << (r*8 + f))
		r -= 1
		f += 1

		if r < 0 || f > 7 || (pieces>>((r+1)*8+f-1)&0b1 == 1) {
			break
		}
	}

	r = rank
	f = file
	for {
		res = res | (0b1 << (r*8 + f))
		r += 1
		f += 1

		if r > 7 || f > 7 || (pieces>>((r-1)*8+f-1)&0b1 == 1) {
			break
		}
	}

	r = rank
	f = file
	for {
		res = res | (0b1 << (r*8 + f))
		r += 1
		f -= 1

		if r > 7 || f < 0 || (pieces>>((r-1)*8+f+1)&0b1 == 1) {
			break
		}
	}

	return res
}
