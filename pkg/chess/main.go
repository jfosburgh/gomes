package chess

import (
	"fmt"
	"math/rand"
	"time"
)

type ChessGame struct {
	EBE         EBE
	Bitboard    BitBoard
	Moves       []Move
	Captured    []int
	SearchDepth int
}

func NewGame() *ChessGame {
	c := ChessGame{
		EBE:         DefaultBoard(),
		Bitboard:    make(BitBoard),
		SearchDepth: 1,
	}

	c.Bitboard.FromEBE(c.EBE.Board)

	return &c
}

func (c *ChessGame) SetStateFromFEN(fen string) {
	c.EBE.FromFEN(fen)
	c.Bitboard.FromEBE(c.EBE.Board)
	c.Moves = []Move{}
	c.Captured = []int{}
}

func copyBitboard(source, dest BitBoard) {
	for piece := range source {
		dest[piece] = source[piece]
	}
}

func copyBoard(source EBEBoard) EBEBoard {
	board := EBEBoard{}
	for i := range 64 {
		board[i] = source[i]
	}

	return board
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

		if c.EBE.Board[pawn+forward] != 0 && c.EBE.Board[diagLeft]&0b1000 != enemy(side) && c.EBE.Board[diagRight] != enemy(side) {
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

func (c *ChessGame) BestMove() Move {
	candidates := c.GetLegalMoves()

	// TODO: Change this to move search
	return candidates[rand.Intn(len(candidates))]
}

func (c *ChessGame) MoveFromLocations(start, end int) (Move, bool) {
	pseudoLegal := c.GeneratePseudoLegal()
	active := c.EBE.Active << 3
	for _, move := range pseudoLegal {
		if move.Start != start || move.End != end {
			continue
		}

		legal := false
		c.MakeMove(move)
		if !c.Bitboard.InCheck(active) {
			legal = true
		}
		c.UnmakeMove(move)

		return move, legal
	}

	return Move{}, false
}

func (c *ChessGame) GetLegalMoves() []Move {
	pseudoLegal := c.GeneratePseudoLegal()

	moves := []Move{}
	active := c.EBE.Active << 3
	for _, move := range pseudoLegal {
		c.MakeMove(move)
		if !c.Bitboard.InCheck(active) {
			moves = append(moves, move)
		}
		c.UnmakeMove(move)
	}

	return moves
}

func (c *ChessGame) GetMoveTargets(pieceLocation int) []int {
	fmt.Printf("getting moves for %d\n", pieceLocation)
	pseudoLegal := c.GeneratePseudoLegal()
	fmt.Printf("found %d pseudolegal moves\n", len(pseudoLegal))

	moves := []int{}
	active := c.EBE.Active << 3
	for _, move := range pseudoLegal {
		if move.Start != pieceLocation {
			fmt.Printf("discarding %+v, %d->%d\n", move, move.Start, move.End)
			continue
		}

		c.MakeMove(move)
		if !c.Bitboard.InCheck(active) {
			moves = append(moves, move.End)
		}
		c.UnmakeMove(move)
	}

	return moves
}

func (c *ChessGame) Perft(depth, startDepth int, debug bool) (int, string) {
	start := time.Time{}
	if depth == startDepth {
		start = time.Now()
	}
	if depth == 0 {
		return 1, ""
	}

	resultString := ""

	count := 0
	moves := c.GeneratePseudoLegal()
	if debug && depth == startDepth {
		fmt.Printf("starting search with board state:\nActive - %d\nCastling Rights - %04b\n%s\n", c.EBE.Active, c.EBE.CastlingRights, c.EBE.Board)
		for piece := range piece2String {
			fmt.Printf("%s: %+v, ", piece2String[piece], toPieceLocations(c.Bitboard[piece]))
		}
		fmt.Println(moves)
	}

	for _, move := range moves {
		moveCount := 0
		active := c.EBE.Active << 3

		startingBitboard := make(BitBoard)
		startingBoard := EBE{}
		if debug {
			copyBitboard(c.Bitboard, startingBitboard)
			startingBoard.Board = copyBoard(c.EBE.Board)
		}

		c.MakeMove(move)
		middleBitboard := make(BitBoard)
		middleBoard := EBE{}
		if debug {
			copyBitboard(c.Bitboard, middleBitboard)
			middleBoard.Board = copyBoard(c.EBE.Board)
			// fmt.Printf("Checking if %04b is in check with board state\n%s\n", active, c.EBE.Board)
		}
		if !c.Bitboard.InCheck(active) {
			c, _ := c.Perft(depth-1, startDepth, debug)
			moveCount += c
		}
		c.UnmakeMove(move)

		if debug {
			if startingBoard.Board != c.EBE.Board {
				panic(fmt.Sprintf("board before move %s%s doesn't match board after\nBefore:\n%s\nDuring:\n%s\nAfter:\n%s", int2algebraic(move.Start), int2algebraic(move.End), startingBoard.Board, middleBoard.Board, c.EBE.Board))
			}

			for piece := range startingBitboard {
				if startingBitboard[piece] != c.Bitboard[piece] {
					panic(fmt.Sprintf("Piece board for %s is different after %s%s\nStarting\n%s\nDuring\n%s\nEnding\n%s", piece2String[piece], int2algebraic(move.Start), int2algebraic(move.End), To2DString(startingBitboard[piece]), To2DString(middleBitboard[piece]), To2DString(c.Bitboard[piece])))
				}
			}
		}

		count += moveCount
		if depth == startDepth && moveCount != 0 {
			resultString += fmt.Sprintf("%s: %d\n", move, moveCount)
		}
	}

	if debug && depth == startDepth {
		fmt.Printf("\nending search with board state:\nActive - %d\nCastling Rights - %04b\n%s\n", c.EBE.Active, c.EBE.CastlingRights, c.EBE.Board)
		for piece := range piece2String {
			fmt.Printf("%s: %+v, ", piece2String[piece], toPieceLocations(c.Bitboard[piece]))
		}
		fmt.Println()
	}
	if depth == startDepth {
		fmt.Printf("perft evaluated to depth of %d in %dms\n", startDepth, time.Since(start).Milliseconds())
	}

	return count, resultString
}
