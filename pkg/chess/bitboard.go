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

func (b *BitBoard) AllPieces() uint64 {
	result := uint64(0)
	for _, pieceBoard := range *b {
		result = result | pieceBoard
	}

	return result
}
