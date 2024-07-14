package chess

import (
	"fmt"
	"math/bits"
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

	b.UpdateSide(WHITE)
	b.UpdateSide(BLACK)
}

func (b BitBoard) UpdateSide(side int) {
	pieces := b.SidePieces(side)
	b[side] = pieces
}

func (b BitBoard) AllPieces() uint64 {
	return b[WHITE] | b[BLACK]
}

func (b BitBoard) SidePieces(side int) uint64 {
	return b[side|PAWN] | b[side|ROOK] | b[side|KNIGHT] | b[side|BISHOP] | b[side|QUEEN] | b[side|KING]
}

func (b BitBoard) SideThreatens(side int) uint64 {
	threatens, _ := b.PawnMoves(side)
	threatens = threatens | b.RookMoves(side)
	threatens = threatens | b.KnightMoves(side)
	threatens = threatens | b.BishopMoves(side)
	threatens = threatens | b.QueenMoves(side)
	threatens = threatens | b.KingMoves(side)

	return threatens
}

func (b BitBoard) InCheck(side int) bool {
	return b[side|KING]&b.SideThreatens(enemy(side)) != 0
}

func (b BitBoard) PawnMoves(side int) (uint64, uint64) {
	selfBitboard := b[side]
	pawns := b[side|PAWN]
	allPieces := b.AllPieces()

	if side != WHITE {
		selfBitboard = rotate180(selfBitboard)
		pawns = rotate180(pawns)
		allPieces = rotate180(allPieces)
	}

	potentialAttacks := uint64(0)
	potentialAttacks = (^fileMask(1) & pawns) << NORTHWEST
	potentialAttacks = potentialAttacks | ((^fileMask(8) & pawns) << NORTHEAST)

	var singleAdvance, doubleAdvance uint64
	doubleAdvanceable := pawns & (rankMask(2) | rankMask(7))

	singleAdvance = pawns << 8
	doubleAdvance = (doubleAdvanceable << 8 & (^allPieces)) << 8

	if side != WHITE {
		return rotate180(potentialAttacks & (^selfBitboard)), rotate180((singleAdvance | doubleAdvance) & (^allPieces))
	}
	return potentialAttacks & (^selfBitboard), (singleAdvance | doubleAdvance) & (^allPieces)
}

func (b BitBoard) KnightMoves(side int) uint64 {
	selfBitboard := b[side]

	return getKnightMoves(b[KNIGHT|side]) & (^selfBitboard)
}

func (b BitBoard) KingMoves(side int) uint64 {
	selfBitboard := b[side]

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
	enemyBitboard := b[enemy(side)]
	selfBitboard := b[side]

	moves := uint64(0)
	locs := toPieceLocations(b[side|ROOK])
	for _, loc := range locs {
		moves = moves | verticalCrossMasked(loc, (enemyBitboard|selfBitboard)&(^(0b1<<loc)))
	}

	moves = moves & (^selfBitboard)

	return moves
}

func (b BitBoard) BishopMoves(side int) uint64 {
	enemyBitboard := b[enemy(side)]
	selfBitboard := b[side]

	moves := uint64(0)
	locs := toPieceLocations(b[side|BISHOP])
	for _, loc := range locs {
		moves = moves | diagonalCrossMasked(loc, (enemyBitboard|selfBitboard)&(^(0b1<<loc)))
	}

	moves = moves & (^selfBitboard)

	return moves
}

func (b BitBoard) QueenMoves(side int) uint64 {
	enemyBitboard := b[enemy(side)]
	selfBitboard := b[side]

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
	if bitboard == 0 {
		return locations
	}

	shifted := 0

	for {
		shift := bits.TrailingZeros64(bitboard)
		if shift == 64 {
			break
		}

		shifted += shift

		locations = append(locations, shifted)
		shifted += 1
		bitboard = bitboard >> (shift + 1)
	}

	return locations
}

func mirrorHorizontal(x uint64) uint64 {
	k1 := uint64(0x5555555555555555)
	k2 := uint64(0x3333333333333333)
	k4 := uint64(0x0f0f0f0f0f0f0f0f)
	x = ((x >> 1) & k1) | ((x & k1) << 1)
	x = ((x >> 2) & k2) | ((x & k2) << 2)
	x = ((x >> 4) & k4) | ((x & k4) << 4)
	return x
}

func flipVertical(x uint64) uint64 {
	return (x << 56) |
		((x << 40) & uint64(0x00ff000000000000)) |
		((x << 24) & uint64(0x0000ff0000000000)) |
		((x << 8) & uint64(0x000000ff00000000)) |
		((x >> 8) & uint64(0x00000000ff000000)) |
		((x >> 24) & uint64(0x0000000000ff0000)) |
		((x >> 40) & uint64(0x000000000000ff00)) |
		(x >> 56)
}

func flipDiagA1H8(x uint64) uint64 {
	t := uint64(0)
	k1 := uint64(0x5500550055005500)
	k2 := uint64(0x3333000033330000)
	k4 := uint64(0x0f0f0f0f00000000)
	t = k4 & (x ^ (x << 28))
	x ^= t ^ (t >> 28)
	t = k2 & (x ^ (x << 14))
	x ^= t ^ (t >> 14)
	t = k1 & (x ^ (x << 7))
	x ^= t ^ (t >> 7)
	return x
}

func rotate90Clockwise(x uint64) uint64 {
	return flipVertical(flipDiagA1H8(x))
}

func rotate180(x uint64) uint64 {
	return mirrorHorizontal(flipVertical(x))
}

func fileMask(file int) uint64 {
	return rotate90Clockwise(rankMask(file))
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

const (
	diag     = uint64(9241421688590303745)
	antiDiag = uint64(72624976668147840)
)

func diagonalCross(pos int) uint64 {
	res := uint64(0)

	rank := pos / 8
	file := pos % 8

	onDiag := 8 * (rank - file)
	if onDiag >= 0 {
		res = diag << onDiag
	} else {
		res = diag >> (onDiag * -1)
	}

	offDiag := 8 * (file - rank + (-7 + 2*rank))
	if offDiag >= 0 {
		res = res | antiDiag<<offDiag
	} else {
		res = res | antiDiag>>(offDiag*-1)
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

func getKnightMoves(knightLoc uint64) uint64 {
	moves := uint64(0)
	//ENE & ESE
	moves = moves | ((knightLoc & (^(rankMask(8) | fileMask(7) | fileMask(8)))) << (EAST + NORTHEAST))
	moves = moves | ((knightLoc & (^(rankMask(1) | fileMask(7) | fileMask(8)))) >> (WEST + NORTHWEST))

	//NNE & SSE
	moves = moves | ((knightLoc & (^(rankMask(8) | rankMask(7) | fileMask(8)))) << (NORTH + NORTHEAST))
	moves = moves | ((knightLoc & (^(rankMask(1) | rankMask(2) | fileMask(8)))) >> (NORTH + NORTHWEST))

	//WNW & WSW
	moves = moves | ((knightLoc & (^(rankMask(8) | fileMask(1) | fileMask(2)))) << (WEST + NORTHWEST))
	moves = moves | ((knightLoc & (^(rankMask(1) | fileMask(1) | fileMask(2)))) >> (EAST + NORTHEAST))

	//NNE & SSE
	moves = moves | ((knightLoc & (^(rankMask(8) | rankMask(7) | fileMask(1)))) << (NORTH + NORTHWEST))
	moves = moves | ((knightLoc & (^(rankMask(1) | rankMask(2) | fileMask(1)))) >> (NORTH + NORTHEAST))

	return moves
}

func enemy(side int) int {
	return ((^side >> 3) & 0b1) << 3
}
